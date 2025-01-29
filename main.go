package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Client представляет подключенного пользователя
type Client struct {
	conn *websocket.Conn
	room string
}

var (
	clients   = make(map[*Client]bool)
	clientsMu sync.Mutex
)

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	client := &Client{conn: conn, room: "default"} // Пока одна комната для всех
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	log.Printf("New client connected. Total: %d", len(clients))

	for {
		var msg struct {
			Type string `json:"type"` // "offer", "answer", "candidate"
			Data string `json:"data"`
			To   string `json:"to"` // ID получателя (пока не используем)
		}

		if err := conn.ReadJSON(&msg); err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		log.Println("Message received:", msg)

		// Рассылаем сообщение всем, кроме отправителя
		clientsMu.Lock()
		for c := range clients {
			if c != client {
				if err := c.conn.WriteJSON(msg); err != nil {
					log.Printf("Write error: %v", err)
				}
			}
		}
		clientsMu.Unlock()
	}

	clientsMu.Lock()
	delete(clients, client)
	clientsMu.Unlock()
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
