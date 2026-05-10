// internal/transport/websocket/client.go
package websocket

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// how long we wait for a pong from the client before disconnecting
	pongWait = 60 * time.Second

	// how often we ping the client (must be less than pongWait)
	pingPeriod = 50 * time.Second

	// max size of a single incoming message in bytes
	maxMessageSize = 4096
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	// Allow all origins for now.
	// In production you'd check the Origin header here.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client represents one connected WebSocket user.
// The Hub keeps track of all Clients.
type Client struct {
	hub *Hub

	// the actual WebSocket connection
	conn *websocket.Conn

	// userID is set after the client sends EventInit and we know who they are
	userID string

	// send is a buffered channel of outgoing messages.
	// WritePump reads from this and sends to the WebSocket.
	// Nothing writes directly to conn except WritePump.
	send chan []byte
}

// ServeWS upgrades an HTTP request to a WebSocket connection,
// creates a Client, registers it with the Hub, and starts its pumps.
// This is the function you attach to your HTTP router in main.go.
func ServeWS(hub *Hub, onMessage func(client *Client, data []byte), w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("[WS] Upgrade error:", err)
		return
	}

	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
		// userID is empty until EventInit is processed
	}

	// Register with hub so the hub knows this client exists
	hub.register <- client

	// Start the two goroutines.
	// These run concurrently for the lifetime of the connection.
	go client.writePump()
	go client.readPump(onMessage)
}

// readPump runs in its own goroutine.
// It reads raw bytes from the WebSocket and passes them to onMessage.
// When the connection closes (or errors), it unregisters the client.
func (c *Client) readPump(onMessage func(client *Client, data []byte)) {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))

	// Every time we get a pong from the client, reset the deadline.
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			// websocket.IsUnexpectedCloseError filters out normal close codes
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Println("[WS] Read error:", err)
			}
			break // exit loop → defer runs → client unregistered
		}

		// Hand the raw bytes to the app layer.
		// The app layer decides what to do with them.
		onMessage(c, data)
	}
}

// writePump runs in its own goroutine.
// It reads from the send channel and writes to the WebSocket.
// It also sends periodic pings to detect dead connections.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {

		case data, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if !ok {
				// Hub closed the channel — connection is done
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				fmt.Println("[WS] Write error:", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				// Client didn't respond — connection is dead
				return
			}
		}
	}
}

// Send puts data onto the client's send channel.
// The writePump will pick it up and send it over the WebSocket.
// Call this from anywhere — it's safe because we never touch conn directly.
func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		// Channel is full — client is too slow or dead.
		// The hub will handle cleanup.
		fmt.Println("[WS] Send buffer full for user:", c.userID)
	}
}

// SetUserID is called by the app layer once we know who this client is.
func (c *Client) SetUserID(id string) {
	c.userID = id
}

// GetUserID returns the client's user ID.
func (c *Client) GetUserID() string {
	return c.userID
}
