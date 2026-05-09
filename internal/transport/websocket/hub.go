package websocket

import (
	"fmt"

	"flamingo_chat/internal/app"
)

type Hub struct {
	Clients map[string]*Client

	Register   chan *Client
	Unregister chan *Client

	Handler *app.Handler
}

func NewHub(handler *app.Handler) *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Handler:    handler,
	}
}

func (h *Hub) Run() {

	for {
		select {

		case client := <-h.Register:
			h.Clients[client.ID] = client

			fmt.Println("[WS] Client registered:", client.ID)

		case client := <-h.Unregister:
			delete(h.Clients, client.ID)
			close(client.Send)

			fmt.Println("[WS] Client unregistered:", client.ID)
		}
	}
}

func (h *Hub) SendToClient(userID string, data []byte) {

	client, ok := h.Clients[userID]
	if !ok {
		fmt.Println("[WS] Client not connected:", userID)
		return
	}

	client.Send <- data
}
