package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

// type Payload struct {
// 	EventType string `json:"event_type"`
// 	Data      Data   `json:"data"`
// }

// type Data struct {
// 	ID    string `json:"id"`
// 	Email string `json:"email"`
// }

func main() {
	mux := http.NewServeMux()
	webhook := webhookHandler()
	mux.Handle("POST /webhook", webhook)

	fmt.Println("Server is started at 9000")
	err := http.ListenAndServe(":9000", mux)
	if err != nil {
		log.Fatal(err)
	}
}

func webhookHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Ensure the request method is POST
		if r.Method != http.MethodPost {
			http.Error(w, "Metho not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Read the request body
		defer r.Body.Close() // Closes the body
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusBadRequest)
			return
		}

		log.Printf("Received webhook payload %s\n", string(body))

		// Unmarshaling and structs is commented cause webhook receivers should be schema-agnostic
		// Unmarshal the JSON data into the payload struct
		// var data Payload
		// err = json.Unmarshal(body, &data)
		// if err != nil {
		// 	http.Error(w, "Error pasing JSON Payload", http.StatusBadRequest)
		// 	return
		// }
		// fmt.Printf("Received webhook \n%s\ndata{%s,\n%s}", data.EventType, data.Data.ID, data.Data.Email)
	})
}
