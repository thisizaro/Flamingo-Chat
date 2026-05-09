package main

import (
	"fmt"
	"net/http"

	"flamingo_chat/internal/app"
	"flamingo_chat/internal/server"
	ws "flamingo_chat/internal/transport/websocket"
)

func main() {

	coreServer := server.NewServer()

	hub := ws.NewHub(nil)

	appHandler := app.NewHandler(coreServer, hub)

	hub.Handler = appHandler

	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, w, r)
	})

	fmt.Println("[SERVER] Running on :8080")
	http.ListenAndServe(":8080", nil)
}
