package websocket

import (
	"fmt"

	"flamingo_chat/internal/events"

	gws "github.com/gorilla/websocket"
)

type Client struct {
	ID   string
	Conn *gws.Conn
	Send chan []byte
}

func (c *Client) ReadPump(hub *Hub) {

	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()

		if err != nil {
			fmt.Println("[WS] Read failed:", c.ID)
			break
		}

		event, err := events.ParseEvent(message)
		if err != nil {
			fmt.Println("[WS] Parse error")
			continue
		}

		hub.Handler.Handle(c.ID, event)
	}
}

func (c *Client) WritePump() {

	defer c.Conn.Close()

	for {
		message, ok := <-c.Send
		if !ok {
			return
		}

		if err := c.Conn.WriteMessage(gws.TextMessage, message); err != nil {
			fmt.Println("[WS] Write failed:", c.ID)
			return
		}
	}
}
