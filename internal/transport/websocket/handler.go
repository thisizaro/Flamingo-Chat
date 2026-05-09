package websocket

import (
	"fmt"
	"net/http"

	gws "github.com/gorilla/websocket"
)

var upgrader = gws.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request) {

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("[WS] Upgrade failed:", err)
		return
	}

	clientID := r.URL.Query().Get("user_id")

	client := &Client{
		ID:   clientID,
		Conn: conn,
		Send: make(chan []byte),
	}

	hub.Register <- client

	fmt.Println("[WS] Connected:", clientID)

	go client.WritePump()
	go client.ReadPump(hub)
}
