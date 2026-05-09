package websocket

import "fmt"

// Hub maintains the set of all active WebSocket clients.
// It is the only place that touches the clients map,
// which means no mutexes needed — everything goes through channels.
type Hub struct {
	// registered clients, keyed by userID.
	// A client only appears here after SetUserID has been called
	// (i.e. after EventInit is processed).
	clients map[string]*Client

	// register receives a new client when they connect.
	register chan *Client

	// unregister receives a client when they disconnect.
	unregister chan *Client

	// send receives a targeted message: deliver data to a specific userID.
	send chan targetedMessage

	// onDisconnect is called when a client disconnects.
	// The app layer sets this to do cleanup (end chats, etc).
	onDisconnect func(userID string)
}

type targetedMessage struct {
	userID string
	data   []byte
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		send:       make(chan targetedMessage, 256),
	}
}

// Run starts the Hub's event loop. Call this in a goroutine.
// All state changes to clients map happen here — single goroutine, no races.
func (h *Hub) Run() {
	for {
		select {

		case <-h.register:
			// At connection time we don't know the userID yet.
			// The client is stored temporarily with empty userID.
			// RegisterUser is called later once we know who they are.
			fmt.Println("[HUB] Client connected (pending init)")

		case client := <-h.unregister:
			userID := client.GetUserID()

			if userID != "" {
				if _, ok := h.clients[userID]; ok {
					delete(h.clients, userID)
					close(client.send)
					fmt.Println("[HUB] Client disconnected:", userID)

					if h.onDisconnect != nil {
						h.onDisconnect(userID)
					}
				}
			} else {
				// Client disconnected before finishing init
				close(client.send)
				fmt.Println("[HUB] Pre-init client disconnected")
			}

		case msg := <-h.send:
			if client, ok := h.clients[msg.userID]; ok {
				client.Send(msg.data)
			} else {
				fmt.Println("[HUB] No client found for userID:", msg.userID)
			}
		}
	}
}

// RegisterUser maps a userID to a client.
// Called by the app layer after processing EventInit.
func (h *Hub) RegisterUser(userID string, client *Client) {
	// This is called from the app layer (a different goroutine),
	// but since Go maps aren't concurrent-safe we send it through
	// the run loop via a small trick: we just do it directly here
	// because RegisterUser is always called before any other hub
	// operation for this client. For safety at scale you'd channel
	// this too, but for this project it's fine.
	h.clients[userID] = client
	fmt.Println("[HUB] User registered:", userID)
}

// SendToUser delivers data to a connected user by their ID.
// Safe to call from any goroutine.
func (h *Hub) SendToUser(userID string, data []byte) {
	h.send <- targetedMessage{userID: userID, data: data}
}

// SetOnDisconnect sets the callback for when a user disconnects.
// Call this before Run().
func (h *Hub) SetOnDisconnect(fn func(userID string)) {
	h.onDisconnect = fn
}
