package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"chat-app-project/internal/auth"
	"chat-app-project/internal/chat"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Action struct {
	Type    string `json:"type"`    // "create_room", "join_room", "direct_message"
	Target  string `json:"target"`  // Room name or username
	Message string `json:"message"` // Message content
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}
	roomManager := chat.NewRoomManager()
	users := make(map[string]*websocket.Conn) // Track all users by username

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		username := r.URL.Query().Get("username")
		password := r.URL.Query().Get("password")

		if err := auth.Authenticate(username, password); err != nil {
			http.Error(w, "Unauthorized: "+err.Error(), http.StatusUnauthorized)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("WebSocket upgrade failed:", err)
			return
		}
		defer conn.Close()

		users[username] = conn
		defer delete(users, username)

		for {
			var action Action
			err := conn.ReadJSON(&action)
			if err != nil {
				log.Println("Error reading JSON:", err)
				break
			}

			switch action.Type {
			case "create_room":
				room := roomManager.GetRoom(action.Target)
				room.Join(conn, username)
				log.Printf("User %s created and joined room: %s", username, action.Target)

			case "join_room":
				room := roomManager.GetRoom(action.Target)
				room.Join(conn, username)
				log.Printf("User %s joined room: %s", username, action.Target)

			case "direct_message":
				targetConn, exists := users[action.Target]
				if exists {
					message := username + " (DM): " + action.Message
					err := targetConn.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						log.Printf("Error sending DM: %v", err)
					}
				} else {
					err := conn.WriteMessage(websocket.TextMessage, []byte("User not found"))
					if err != nil {
						log.Printf("Error sending DM failure message: %v", err)
					}
				}

			case "message":
				room := roomManager.GetRoom(action.Target)
				broadcastMsg := action.Message
				room.Broadcast(conn, broadcastMsg)

			case "logout":
				log.Printf("User %s logged out", username)

				// Remove user from all rooms
				for _, room := range roomManager.ListRooms() {
					room.Leave(conn)
				}

				// Remove user from users map
				delete(users, username)

				// Close connection (this also exits the handler loop)
				conn.Close()
				return

			default:
				log.Printf("Unknown action type: %s", action.Type)
			}
		}
	})

	port := os.Getenv("SERVER_PORT")
	log.Println("Server started on :" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
