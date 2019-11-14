package main

import (
	"log"
	"net/http"
	"path"
	"text/template"
	"time"

	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan Message)           // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Message Define our message object
type Message struct {
	ID      string    `json:"id"`
	Author  string    `json:"author"`
	Avatar  string    `json:"avatar"`
	Message string    `json:"message"`
	Image   string    `json:"images"`
	Time    time.Time `json:"time"`
}

func main() {

	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)
	// index
	http.HandleFunc("/home", home)

	// chatRoom
	http.HandleFunc("/chatRoom", chatRoom)

	http.HandleFunc("/sample", sample)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a websocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		var msg Message
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- msg
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func render(w http.ResponseWriter, r *http.Request, fp string) {
	tmpl, err := template.ParseFiles(fp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("../public", "home.html")
	render(w, r, fp)
}

func chatRoom(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("../public", "chatRoom.html")
	render(w, r, fp)
}

func sample(w http.ResponseWriter, r *http.Request) {
	fp := path.Join("../public", "sample.html")
	render(w, r, fp)
}
