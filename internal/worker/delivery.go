package worker

import (
	"bytes"
	"encoding/json"
	"net/http"
)

// Phase 4:
// Worker processes pending deliveries from database

func deliverWebhook(event Event) {
	payload, err := json.Marshal(event)
	if err != nil {
		return
	}
	targetURL := "http://localhost:9000/webhook"
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
