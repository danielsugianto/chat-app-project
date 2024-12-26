package chat

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Room struct {
	mu       sync.Mutex
	users    map[*websocket.Conn]string
	messages []string
}

func NewRoom() *Room {
	return &Room{
		users:    make(map[*websocket.Conn]string),
		messages: make([]string, 0),
	}
}

func (r *Room) Join(conn *websocket.Conn, username string) {
	r.mu.Lock()
	r.users[conn] = username
	r.mu.Unlock()

	log.Printf("%s joined the room", username)
}

func (r *Room) Broadcast(sender *websocket.Conn, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get the sender's username
	senderUsername := r.users[sender]
	formattedMessage := senderUsername + ": " + message

	// Store the message in memory
	r.messages = append(r.messages, formattedMessage)

	// Send the message to all users except the sender
	for conn := range r.users {
		if conn != sender {
			err := conn.WriteMessage(websocket.TextMessage, []byte(formattedMessage))
			if err != nil {
				log.Printf("Error sending message: %v", err)
				conn.Close()
				delete(r.users, conn)
			}
		}
	}
}

func (r *Room) Leave(conn *websocket.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.users, conn)
	log.Printf("User left the room. Active members: %d", len(r.users))
}
