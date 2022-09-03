package controller

import (
	"bytes"
	"net/http"
	"time"

	"github.com/bananocoin/boompow/apps/server/src/middleware"
	"github.com/bananocoin/boompow/libs/utils/net"
	"github.com/gorilla/websocket"
	"k8s.io/klog/v2"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type ClientWSMessage struct {
	ClientEmail string `json:"email"`
	msg         []byte
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(PongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(PongWait)); return nil })
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				klog.Errorf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		msgObj := ClientWSMessage{ClientEmail: c.Email, msg: message}
		c.Hub.Response <- msgObj
	}
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(PingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if !ok {
				// The hub closed the channel.
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func WorkerChl(hub *Hub, w http.ResponseWriter, r *http.Request) {
	provider := middleware.AuthorizedProvider(r.Context())
	// Only PROVIDER type users can provide work
	if provider == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Unauthorized"))
		return
	}

	clientIP := net.GetIPAddress(r)

	// Block hetzner datacenters
	if net.IsIPInHetznerRange(clientIP) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("403 - Forbidden"))
		return
	}

	// Block IPs already connected
	for c := range hub.Clients {
		if c.IPAddress == clientIP {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("403 - Forbidden"))
			return
		}
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		klog.Error(err)
		return
	}
	client := &Client{Hub: hub, Conn: conn, Send: make(chan []byte, 256), IPAddress: clientIP, Email: provider.User.Email}
	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
