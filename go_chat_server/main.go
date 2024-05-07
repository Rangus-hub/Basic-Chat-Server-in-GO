package main

import (
	//"encoding/json" // Import the encoding/json package
	"fmt"
	"net/http"
	"text/template"

	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)

func homePage(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Welcome to the Chat Room!")

	tmpl, err := template.ParseFiles("./Templates/home.html")
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Execute the template
	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

}

func handleConnections(w http.ResponseWriter, r *http.Request) {

	// Upgrade HTTP connection to WebSocket connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error upgrading to WebSocket:", err)
		return
	}
	defer conn.Close()

	// Register client connection
	clients[conn] = true

	start := time.Now()

	for {
		// Read message from WebSocket connection
		var msg Message

		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error reading message:", err)
			break // Exit loop on error
		}

		elapsed := time.Since(start)

		//Printing message
		fmt.Printf("Received message from %s: %s\nTime Taken: %d\n\n", msg.Username, msg.Message, elapsed.Seconds())

		// Forward message to broadcast channel
		broadcast <- msg
	}

	// Unregister client connection
	delete(clients, conn)
}

// func handleConnections(w http.ResponseWriter, r *http.Request) {
// 	conn, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		fmt.Println(err)
// 		return
// 	}
// 	defer conn.Close()

// 	clients[conn] = true

// 	for {
// 		// Read message from client
// 		_, messageBytes, err := conn.ReadMessage()
// 		if err != nil {
// 			fmt.Println(err)
// 			delete(clients, conn)
// 			return
// 		}

// 		// Decode JSON message into Message struct
// 		var message Message
// 		err = json.Unmarshal(messageBytes, &message)
// 		if err != nil {
// 			fmt.Println("Error decoding JSON:", err)
// 			continue
// 		}

// 		// Process message (in this example, just print it)
// 		fmt.Printf("Received message from client: %+v\n", message)

// 		// Broadcast message to all clients (optional)
// 		broadcast <- message
// 	}
// }

func handleMessages() {

	for {
		msg := <-broadcast

		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				fmt.Println(err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func main() {

	http.HandleFunc("/", homePage)

	http.HandleFunc("/ws", handleConnections)

	go handleMessages()

	fmt.Println("Server started on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}
