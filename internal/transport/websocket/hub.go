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

	// broadcast delivers a message to every connected user
	broadcast chan []byte

	//
	registerUser chan registerRequest // routes RegisterUser through Run()

	// onDisconnect is called when a client disconnects.
	// The app layer sets this to do cleanup (end chats, etc).
	onDisconnect func(userID string)
}

type targetedMessage struct {
	userID string
	data   []byte
}

type registerRequest struct {
	userID string
	client *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:      make(map[string]*Client),
		register:     make(chan *Client),
		unregister:   make(chan *Client),
		send:         make(chan targetedMessage, 256),
		broadcast:    make(chan []byte, 64),
		registerUser: make(chan registerRequest, 32),
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

		case req := <-h.registerUser:
			// If this userID already has a connection, evict it cleanly.
			// This handles the case where a user opens a second tab or
			// reconnects before the old connection is fully closed.
			if old, ok := h.clients[req.userID]; ok && old != req.client {
				fmt.Println("[HUB] Evicting duplicate connection for:", req.userID)
				close(old.send)
				// Do NOT fire onDisconnect here — the user is not gone,
				// they just reconnected. Firing it would reset their status.
			}
			h.clients[req.userID] = req.client
			fmt.Println("[HUB] User registered:", req.userID)

		case client := <-h.unregister:
			userID := client.GetUserID()

			if userID != "" {
				// Only clean up and fire onDisconnect if this specific
				// client instance is still the registered one.
				// If it was already evicted (duplicate connection), skip.
				if current, ok := h.clients[userID]; ok && current == client {
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

		case data := <-h.broadcast:
			for _, client := range h.clients {
				client.Send(data)
			}

		}
	}
}

// RegisterUser maps a userID to a client after init is processed.
// Routes through the Run() goroutine so the clients map is never
// touched from outside. Handles duplicate connections safely.
func (h *Hub) RegisterUser(userID string, client *Client) {
	h.registerUser <- registerRequest{userID: userID, client: client}
}

// SendToUser delivers data to a connected user by their ID.
// Safe to call from any goroutine.
func (h *Hub) SendToUser(userID string, data []byte) {
	h.send <- targetedMessage{userID: userID, data: data}
}

// Broadcast delivers data to every connected user.
// Safe to call from any goroutine.
func (h *Hub) Broadcast(data []byte) {
	h.broadcast <- data
}

// SetOnDisconnect sets the callback for when a user disconnects.
// Call this before Run().
func (h *Hub) SetOnDisconnect(fn func(userID string)) {
	h.onDisconnect = fn
}
