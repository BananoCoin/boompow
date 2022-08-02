package controller

import (
	"time"

	"github.com/bbedward/boompow-server-ng/src/database"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
)

var ActiveHub *Hub

// What a work request looks like from the client perspective
type ClientWorkRequest struct {
	Hash                 string `json:"hash"`
	DifficutlyMultiplier int    `json:"difficulty_multiplier"`
}

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
}

var Upgrader = websocket.Upgrader{}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	Clients map[*Client]bool

	// Outbound messages to the client
	Broadcast chan []byte

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[*Client]bool),
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
		case message := <-h.Broadcast:
			glog.Info(string(message))
			for client := range h.Clients {
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
