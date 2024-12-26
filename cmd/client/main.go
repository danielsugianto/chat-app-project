package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

type Action struct {
	Type    string `json:"type"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

var mutex sync.Mutex

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		conn, _, err := loginPage(reader)
		if err != nil {
			log.Printf("Failed to connect: %v\n", err)
			return
		}
		if conn != nil {
			startChat(conn, reader)
		}
	}
}
func loginPage(reader *bufio.Reader) (*websocket.Conn, *http.Response, error) {

	var conn *websocket.Conn
	var err error
	var resp *http.Response

	for {
		// Ask for username and password
		log.Print("Enter username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)

		log.Print("Enter password: ")
		password, _ := reader.ReadString('\n')
		password = strings.TrimSpace(password)

		port := os.Getenv("SERVER_PORT")
		host := os.Getenv("SERVER_HOST")

		fmt.Println(port)
		fmt.Println(host)
		// Create WebSocket URL with query params
		u := url.URL{
			Scheme:   "ws",
			Host:     host + ":" + port,
			Path:     "/ws",
			RawQuery: "username=" + username + "&password=" + password,
		}

		// Connect to the WebSocket server
		conn, resp, err = websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			if resp != nil && resp.StatusCode == 401 {
				log.Println("Authentication failed: Invalid username or password.")
			} else {
				return nil, nil, err
			}
		} else {
			log.Println("Login successful!")
			break
		}
	}
	return conn, resp, nil
}

// StartChat handles the chat session
func startChat(conn *websocket.Conn, reader *bufio.Reader) {

	defer conn.Close()
	// Goroutine to handle incoming messages
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Read error:", err)
				break
			}

			// Lock to prevent cursor overlap
			mutex.Lock()

			// Clear the "Message: " line and print the incoming message
			fmt.Print("\033[2K\r")   // Clear current line
			fmt.Println(string(msg)) // Print the new message

			// Reprint the "Message: " prompt
			fmt.Print("Message: ")
			mutex.Unlock()
		}
	}()

	// Main menu loop
	for {
		log.Println("\nMenu:")
		log.Println("1. Create Room")
		log.Println("2. Join Room")
		log.Println("3. Direct Message")
		log.Println("4. Logout")
		log.Print("Choose an option: ")
		option, _ := reader.ReadString('\n')
		option = strings.TrimSpace(option)

		var action Action
		switch option {
		case "1":
			log.Print("Enter room name to create: ")
			room, _ := reader.ReadString('\n')
			room = strings.TrimSpace(room)
			action = Action{Type: "create_room", Target: room}
			err := conn.WriteJSON(action)
			if err != nil {
				log.Println("Write error:", err)
				break
			}
			enterChatMode(conn, reader, room)

		case "2":
			log.Print("Enter room name to join: ")
			room, _ := reader.ReadString('\n')
			room = strings.TrimSpace(room)
			action = Action{Type: "join_room", Target: room}
			err := conn.WriteJSON(action)
			if err != nil {
				log.Println("Write error:", err)
				break
			}
			enterChatMode(conn, reader, room)

		case "3":
			log.Print("Enter recipient username: ")
			recipient, _ := reader.ReadString('\n')
			recipient = strings.TrimSpace(recipient)
			log.Print("Enter message: ")
			message, _ := reader.ReadString('\n')
			message = strings.TrimSpace(message)
			action = Action{Type: "direct_message", Target: recipient, Message: message}
			err := conn.WriteJSON(action)
			if err != nil {
				log.Println("Write error:", err)
				break
			}

		case "4":
			action = Action{Type: "logout"}
			err := conn.WriteJSON(action)
			if err != nil {
				log.Println("Write error:", err)
				break
			}
			log.Println("Logged out successfully!")
			return
		default:
			log.Println("Invalid option. Please try again.")
		}
	}
}
func enterChatMode(conn *websocket.Conn, reader *bufio.Reader, room string) {
	log.Printf("You are now in the room: %s. Type '/leave' to exit.", room)

	for {
		// Reprint the "Message: " prompt
		fmt.Print("Message: ")
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)

		// Leave room command
		if text == "/leave" {
			log.Printf("Leaving room: %s", room)
			return
		}

		// Send message to the room
		action := Action{Type: "message", Target: room, Message: text}
		err := conn.WriteJSON(action)
		if err != nil {
			log.Println("Write error:", err)
			break
		}
	}
}
