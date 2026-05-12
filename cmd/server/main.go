package main

import (
	"fmt"
	"net/http"
	"os"

	"flamingo_chat/internal/app"
	"flamingo_chat/internal/server"
	ws "flamingo_chat/internal/transport/websocket"
)

func main() {
	// 0. Initialize metrics — set the Google Sheets URL to send metrics to.
	metricsURL := os.Getenv("METRICS_URL")

	//temporary hardcode for testing.
	// metricsURL := ""

	server.InitMetrics(metricsURL)
	fmt.Println("[MAIN] Metrics initialized with URL:", metricsURL)

	// 1. Create the server (business logic + state)
	srv := server.NewServer()

	// 2. Create the hub (WebSocket connection registry)
	hub := ws.NewHub()

	// 3. Create the app (glue between transport and server)
	// Wire sets up all the callbacks — nothing leaks into main.
	application := app.New(srv, hub)
	application.Wire()

	// 4. Start background workers
	srv.StartCleanupWorker()
	go hub.Run()

	// 5. HTTP routes
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ws.ServeWS(hub, application.HandleMessage, w, r)
	})

	// Insider Joke lol.
	http.HandleFunc("/OishikHealth", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Health Report. - Oishik"))
	})

	fmt.Println("[SERVER] Running on :8080")
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}
	http.ListenAndServe(":"+port, nil)
}
