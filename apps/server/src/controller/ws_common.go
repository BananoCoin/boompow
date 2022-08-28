package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

var ActiveHub *Hub

const (
	// Time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	PingPeriod = (PongWait * 9) / 10

	// Maximum message size allowed from peer.
	MaxMessageSize = 512
)

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	Hub *Hub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan []byte

	// IP Address
	IPAddress string

	Email string
}

var Upgrader = websocket.Upgrader{}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Keep a few buckets for clients to distribute work
	ClientsA map[string]*Client
	ClientsB map[string]*Client
	ClientsC map[string]*Client

	// Outbound messages to the client
	Broadcast  chan []byte
	BroadcastA chan []byte
	BroadcastB chan []byte
	BroadcastC chan []byte

	// Inbound messages from client
	Response chan ClientWSMessage

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Channel to broadcast stats to
	StatsChan *chan repository.WorkMessage

	sync.RWMutex
	LastBroadCastBucket string
}

func NewHub(statsChan *chan repository.WorkMessage) *Hub {
	return &Hub{
		Broadcast:           make(chan []byte, 100),
		BroadcastA:          make(chan []byte, 100),
		BroadcastB:          make(chan []byte, 100),
		BroadcastC:          make(chan []byte, 100),
		Response:            make(chan ClientWSMessage),
		Register:            make(chan *Client),
		Unregister:          make(chan *Client),
		ClientsA:            make(map[string]*Client),
		ClientsB:            make(map[string]*Client),
		ClientsC:            make(map[string]*Client),
		StatsChan:           statsChan,
		LastBroadCastBucket: "A",
	}
}

func (h *Hub) BlockAwardedWorker(blockAwardedChan <-chan serializableModels.ClientMessage) {
	for ba := range blockAwardedChan {
		for k := range h.ClientsA {
			if h.ClientsA[k].Email == ba.ProviderEmail {
				bytes, err := json.Marshal(ba)
				if err != nil {
					klog.Errorf("Error marshalling block awarded message %s", err)
					break
				}
				fmt.Printf("Awarding to %s", k)
				h.ClientsA[k].Send <- bytes
			}
		}
		for k := range h.ClientsB {
			if h.ClientsB[k].Email == ba.ProviderEmail {
				bytes, err := json.Marshal(ba)
				if err != nil {
					klog.Errorf("Error marshalling block awarded message %s", err)
					break
				}
				fmt.Printf("Awarding to %s", k)
				h.ClientsB[k].Send <- bytes
			}
		}
		for k := range h.ClientsC {
			if h.ClientsC[k].Email == ba.ProviderEmail {
				bytes, err := json.Marshal(ba)
				if err != nil {
					klog.Errorf("Error marshalling block awarded message %s", err)
					break
				}
				fmt.Printf("Awarding to %s", k)
				h.ClientsC[k].Send <- bytes
			}
		}
	}
}

func (h *Hub) UnregisterClient(client *Client) {
	h.Unregister <- client
	if _, ok := h.ClientsA[client.IPAddress]; ok {
		delete(h.ClientsA, client.IPAddress)
		close(client.Send)
		// Keep global state of connected clients
		database.GetRedisDB().RemoveConnectedClient(client.IPAddress)
	} else if _, ok := h.ClientsB[client.IPAddress]; ok {
		delete(h.ClientsB, client.IPAddress)
		close(client.Send)
		// Keep global state of connected clients
		database.GetRedisDB().RemoveConnectedClient(client.IPAddress)
	} else if _, ok := h.ClientsC[client.IPAddress]; ok {
		delete(h.ClientsC, client.IPAddress)
		close(client.Send)
		// Keep global state of connected clients
		database.GetRedisDB().RemoveConnectedClient(client.IPAddress)
	}
}

func (h *Hub) AddClientToBucket(client *Client) {
	// See if already in a bucket
	added := false
	if _, ok := h.ClientsA[client.IPAddress]; ok {
		added = true
		h.ClientsA[client.IPAddress] = client
	} else if _, ok := h.ClientsB[client.IPAddress]; ok {
		added = true
		h.ClientsB[client.IPAddress] = client
	} else if _, ok := h.ClientsC[client.IPAddress]; ok {
		added = true
		h.ClientsC[client.IPAddress] = client
	}
	if added {
		return
	}

	// If not in a bucket, add to a balanced bucket
	if len(h.ClientsA) < len(h.ClientsB) {
		if len(h.ClientsA) < len(h.ClientsC) {
			// Add to A
			h.ClientsA[client.IPAddress] = client
		}
	} else if len(h.ClientsB) < len(h.ClientsC) {
		// Add to B
		h.ClientsB[client.IPAddress] = client
	} else {
		// Add to C
		h.ClientsC[client.IPAddress] = client
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Determine if user is already in a bucket
			h.AddClientToBucket(client)
			// Keep global state of connected clients
			database.GetRedisDB().AddConnectedClient(client.IPAddress)
		case client := <-h.Unregister:
			h.UnregisterClient(client)
		case message := <-h.Response:
			// Try to unmarshal as ClientWorkResponse
			var workResponse serializableModels.ClientWorkResponse
			err := json.Unmarshal(message.msg, &workResponse)
			// If this channel exists, send response
			activeChannel := ActiveChannels.Get(workResponse.RequestID)
			if activeChannel != nil {
				// Validate this work
				if !validation.IsWorkValid(activeChannel.Hash, activeChannel.DifficultyMultiplier, workResponse.Result) {
					klog.Errorf("Received invalid work for %s", activeChannel.Hash)
					// ! TODO - penalize this bad client
					continue
				}
				// Send work cancel command to all clients
				workCancel := &serializableModels.ClientMessage{
					MessageType: serializableModels.WorkCancel,
					Hash:        activeChannel.Hash,
				}
				bytes, err := json.Marshal(workCancel)
				if err != nil {
					klog.Errorf("Failed to marshal work cancel command: %v", err)
				} else {
					go func() { ActiveHub.Broadcast <- bytes }()
				}
				// Credit this client for this work
				// Except for some services people can abuse, like BananoVault
				if activeChannel.RequesterEmail != "vault@banano.cc" {
					statsMessage := repository.WorkMessage{
						BlockAward:           activeChannel.BlockAward,
						ProvidedByEmail:      message.ClientEmail,
						RequestedByEmail:     activeChannel.RequesterEmail,
						Hash:                 activeChannel.Hash,
						Result:               workResponse.Result,
						DifficultyMultiplier: activeChannel.DifficultyMultiplier,
					}
					*h.StatsChan <- statsMessage
				} else {
					// Still cache
					database.GetRedisDB().CacheWork(activeChannel.Hash, workResponse.Result)
				}
				WriteChannelSafe(activeChannel.Chan, message.msg)
			} else {
				klog.V(3).Infof("Received work response for hash %s, but no channel exists", workResponse.Hash)
			}
			// Error de-serializing
			if err != nil {
				klog.Errorf("Error unmarshalling work response: %s", err)
				continue
			}
		case message := <-h.BroadcastA:
			for k := range h.ClientsA {
				select {
				case h.ClientsA[k].Send <- message:
				default:
					close(h.ClientsA[k].Send)
					delete(h.ClientsA, k)
				}
			}
		case message := <-h.BroadcastB:
			for k := range h.ClientsB {
				select {
				case h.ClientsB[k].Send <- message:
				default:
					close(h.ClientsB[k].Send)
					delete(h.ClientsB, k)
				}
			}
		case message := <-h.BroadcastC:
			for k := range h.ClientsC {
				select {
				case h.ClientsC[k].Send <- message:
				default:
					close(h.ClientsC[k].Send)
					delete(h.ClientsC, k)
				}
			}
		case message := <-h.Broadcast:
			h.BroadcastA <- message
			h.BroadcastB <- message
			h.BroadcastC <- message
		}
	}
}

// Recover from panic if the channel is closed
func WriteChannelSafe(out chan []byte, msg []byte) (err error) {

	defer func() {
		// recover from panic caused by writing to a closed channel
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			return
		}
	}()

	out <- msg // write on possibly closed channel

	return err
}

// Channels for reach specific work request
var ActiveChannels = models.NewSyncArray()

// Timeout waiting for work response from client
const WORK_TIMEOUT_S = time.Second * 30

// Get last broadcast bucket thread-safe
func (h *Hub) GetBroadcastBucket() string {
	h.RLock()
	defer h.RUnlock()
	if h.LastBroadCastBucket == "A" {
		h.LastBroadCastBucket = "B"
		return "B"
	}
	if h.LastBroadCastBucket == "B" {
		h.LastBroadCastBucket = "C"
		return "C"
	}
	h.LastBroadCastBucket = "A"
	return "A"
}

// Method to handle a work request response
// 1) Broadcast to every client
// 2) Create a channel for the response
// 3) Wait for response on the channel until timeout
func (h *Hub) BroadcastWorkRequestAndWait(workRequest *serializableModels.ClientMessage) (*serializableModels.ClientWorkResponse, error) {
	// Serialize
	bytes, err := json.Marshal(workRequest)
	if err != nil {
		return nil, err
	}
	// Create channel for this hash
	activeChannelObj := models.ActiveChannelObject{
		BlockAward:           workRequest.BlockAward,
		RequesterEmail:       workRequest.RequesterEmail,
		RequestID:            workRequest.RequestID,
		Hash:                 workRequest.Hash,
		DifficultyMultiplier: workRequest.DifficultyMultiplier,
		Chan:                 make(chan []byte),
	}
	ActiveChannels.Put(activeChannelObj)

	broadcastBucket := h.GetBroadcastBucket()
	switch broadcastBucket {
	case "A":
		go func() { ActiveHub.BroadcastA <- bytes }()
	case "B":
		go func() { ActiveHub.BroadcastB <- bytes }()
	case "C":
		go func() { ActiveHub.BroadcastC <- bytes }()
	}
	select {
	case response := <-activeChannelObj.Chan:
		var workResponse serializableModels.ClientWorkResponse
		err := json.Unmarshal(response, &workResponse)
		if err != nil {
			return nil, err
		}
		// Close channel
		close(activeChannelObj.Chan)
		ActiveChannels.Delete(workRequest.RequestID)
		return &workResponse, nil
	case <-time.After(1 * time.Second):
		klog.Errorf("Work request timed out %s, sending to all buckets", workRequest.Hash)
		// Close channel
		close(activeChannelObj.Chan)
		ActiveChannels.Delete(workRequest.RequestID)
		return h.BroadcastWorkRequestAndWaitToEverybody(workRequest)
	}
}

func (h *Hub) BroadcastWorkRequestAndWaitToEverybody(workRequest *serializableModels.ClientMessage) (*serializableModels.ClientWorkResponse, error) {
	// Serialize
	bytes, err := json.Marshal(workRequest)
	if err != nil {
		return nil, err
	}
	// Create channel for this hash
	activeChannelObj := models.ActiveChannelObject{
		BlockAward:           workRequest.BlockAward,
		RequesterEmail:       workRequest.RequesterEmail,
		RequestID:            workRequest.RequestID,
		Hash:                 workRequest.Hash,
		DifficultyMultiplier: workRequest.DifficultyMultiplier,
		Chan:                 make(chan []byte),
	}
	ActiveChannels.Put(activeChannelObj)

	go func() { ActiveHub.Broadcast <- bytes }()
	select {
	case response := <-activeChannelObj.Chan:
		var workResponse serializableModels.ClientWorkResponse
		err := json.Unmarshal(response, &workResponse)
		if err != nil {
			return nil, err
		}
		// Close channel
		close(activeChannelObj.Chan)
		ActiveChannels.Delete(workRequest.RequestID)
		return &workResponse, nil
	// 30
	case <-time.After(WORK_TIMEOUT_S):
		klog.Errorf("Work request timed out %s", workRequest.Hash)
		// Close channel
		close(activeChannelObj.Chan)
		ActiveChannels.Delete(workRequest.RequestID)
		return nil, errors.New("timeout")
	}
}
