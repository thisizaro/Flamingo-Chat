package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// metricsSheetURL is your Google Apps Script deployment URL.
// Set this before starting the server — either hardcode it here
// or load it from an environment variable (recommended for Render).
var metricsSheetURL = ""

type metricEntry struct {
	Time   string `json:"time"`
	Event  string `json:"event"`
	UserID string `json:"user_id,omitempty"`
	Detail string `json:"detail,omitempty"`
}

// InitMetrics sets the Google Sheets URL to send metrics to.
// Call this from main.go before starting the server.
// On Render: set an environment variable METRICS_URL and pass it here.
func InitMetrics(url string) {
	metricsSheetURL = url
	fmt.Println("[METRICS] Initialized, sending to Google Sheets")
}

// LogMetric sends one metric event to Google Sheets.
// Runs in a goroutine so it never blocks the server.
// If the URL is not set or the request fails, it logs and moves on.
func LogMetric(event, userID, detail string) {
	if metricsSheetURL == "" {
		return
	}

	entry := metricEntry{
		Time:   time.Now().Format("2006-01-02 15:04:05"),
		Event:  event,
		UserID: userID,
		Detail: detail,
	}

	// Fire and forget — runs in background, never blocks
	go func() {
		body, err := json.Marshal(entry)
		if err != nil {
			fmt.Println("[METRICS] Failed to marshal:", err)
			return
		}

		resp, err := http.Post(metricsSheetURL, "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("[METRICS] Failed to send:", err)
			return
		}
		defer resp.Body.Close()
		fmt.Println(
			"[METRICS] Sent:",
			event,
			"user:",
			userID,
			"status:",
			resp.Status,
		)
	}()
}
