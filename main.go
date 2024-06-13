package main

import (
	"chatserver/data"
	"chatserver/spotify"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

func messageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var msg data.Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	response := data.Message{Data: "Received: " + msg.Data}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade:", err)
		return
	}
	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read:", err)
			break
		}
		log.Printf("Received: %s", message)

		if err := conn.WriteMessage(messageType, message); err != nil {
			log.Println("Write:", err)
			break
		}
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file")
	}

	router := mux.NewRouter()

	//---WebSocket Routes--------
	router.HandleFunc("/message", messageHandler)
	router.HandleFunc("/ws", handleWebSocket)

	//---Spotify Routes----------
	router.HandleFunc("/login", spotify.Login).Methods("GET", "OPTIONS")
	router.HandleFunc("/callback", spotify.Callback).Methods("GET", "OPTIONS")
	router.HandleFunc("/refresh_token", spotify.Refresh).Methods("GET", "OPTIONS")

	fmt.Println("Server is listening on 5001...")
	log.Fatal(http.ListenAndServe(":5001", router))
}
