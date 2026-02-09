package worker

import "fmt"

// Phase 4:
// Worker processes pending

func worker(jobQueue <-chan Event) {
	for event := range jobQueue {
		fmt.Println("Worker delivering event")
		deliverWebhook(event)
	}
}
