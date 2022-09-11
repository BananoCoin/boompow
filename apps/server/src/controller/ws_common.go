package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/bananocoin/boompow/apps/server/src/database"
	"github.com/bananocoin/boompow/apps/server/src/models"
	"github.com/bananocoin/boompow/apps/server/src/repository"
	serializableModels "github.com/bananocoin/boompow/libs/models"
	"github.com/bananocoin/boompow/libs/utils/validation"
	"github.com/gorilla/websocket"
	"golang.org/x/exp/slices"
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
	// Registered clients.
	Clients map[*Client]bool

	// Outbound messages to the client
	Broadcast chan []byte

	// Inbound messages from client
	Response chan ClientWSMessage

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Channel to broadcast stats to
	StatsChan *chan repository.WorkMessage
}

func NewHub(statsChan *chan repository.WorkMessage) *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Response:   make(chan ClientWSMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
		StatsChan:  statsChan,
	}
}

func (h *Hub) BlockAwardedWorker(blockAwardedChan <-chan serializableModels.ClientMessage) {
	for ba := range blockAwardedChan {
		for c := range h.Clients {
			if c.Email == ba.ProviderEmail {
				bytes, err := json.Marshal(ba)
				if err != nil {
					klog.Errorf("Error marshalling block awarded message %s", err)
					break
				}
				fmt.Printf("Awarding to %s", c.IPAddress)
				database.GetRedisDB().UpdateClientScore(c.IPAddress, int(ba.DifficultyMultiplier))
				c.Send <- bytes
			}
		}
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			// Keep global state of connected clients
			database.GetRedisDB().AddConnectedClient(client.IPAddress)
		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
				// Keep global state of connected clients
				database.GetRedisDB().RemoveConnectedClient(client.IPAddress)
			}
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
				if activeChannel.RequesterEmail == "vault@banano.cc" {
					activeChannel.BlockAward = false
				}
				statsMessage := repository.WorkMessage{
					BlockAward:           activeChannel.BlockAward,
					ProvidedByEmail:      message.ClientEmail,
					RequestedByEmail:     activeChannel.RequesterEmail,
					Hash:                 activeChannel.Hash,
					Result:               workResponse.Result,
					DifficultyMultiplier: activeChannel.DifficultyMultiplier,
				}
				*h.StatsChan <- statsMessage
				WriteChannelSafe(activeChannel.Chan, message.msg)
			} else {
				klog.V(3).Infof("Received work response for hash %s, but no channel exists", workResponse.Hash)
			}
			// Error de-serializing
			if err != nil {
				klog.Errorf("Error unmarshalling work response: %s", err)
				continue
			}
		case message := <-h.Broadcast:
			toExclude, err := database.GetRedisDB().FilterOverperformingClients()
			if err != nil {
				klog.Errorf("Error filtering overperforming clients: %v", err)
				toExclude = []string{}
			}
			if len(h.Clients) < 5 {
				toExclude = []string{}
				klog.V(3).Infof("Not enough clients to exclude any")
			}
			for client := range h.Clients {
				if len(toExclude) > 0 && slices.Contains(toExclude, client.IPAddress) {
					continue
				}
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
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

// Method to handle a work request response
// 1) Broadcast to every client
// 2) Create a channel for the response
// 3) Wait for response on the channel until timeout
func BroadcastWorkRequestAndWait(workRequest serializableModels.ClientMessage) (*serializableModels.ClientWorkResponse, error) {
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
