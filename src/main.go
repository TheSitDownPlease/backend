package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var broadcast = make(chan RoomAction)        // broadcast channel

// Configure the upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var wsAddr = flag.String("addr", "localhost:8080", "http service address")

// Message Define our message object
type Message struct {
	ID      string `json:"id"`
	Author  string `json:"author"`
	Avatar  string `json:"avatar"`
	Message string `json:"message"`
	Image   string `json:"images"`
	Time    int64  `json:"time"`
}

// RoomAction the room action
type RoomAction struct {
	Action  string  `json:"action"`
	Message Message `json:"data"`
}

func main() {
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	http.HandleFunc("/save_msg", saveMessage)

	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)
	// Start listening for incoming chat messages
	go handleMessages()
	// Start the server on localhost port 8000 and log any errors
	fmt.Println("Go Server is start at localhost:8000")
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
		var action RoomAction
		// Read in a new message as JSON and map it to a Message object
		err := ws.ReadJSON(&action)
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// Send the newly received message to the broadcast channel
		broadcast <- action

		insertMessageToDB(&action.Message)
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		action := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(action)
			if err != nil {
				log.Printf("error: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

func saveMessage(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var action RoomAction
	err = json.Unmarshal(body, &action)

	if err != nil {
		log.Fatal("json unmarshal fail: ", err)
	}

	fmt.Println("before insert db")
	insertMessageToDB(&action.Message)
	fmt.Println("after insert db")

	// fakeMsg := Message{
	// 	Author:  "Blues",
	// 	ID:      "this is id",
	// 	Avatar:  "fake.avatar.url",
	// 	Message: "this is generate from fake message",
	// 	Image:   "fake.image.url",
	// 	Time:    time.Now().Unix(),
	// }
	// fmt.Println("before insert db")
	// insertMessageToDB(&action.Message)
	// fmt.Println("after insert db")
	//w.Write([]byte("<h1>Hello World!</h1>"))
}

func insertMessageToDB(msg *Message) {
	// Open up our database connection.
	// I've set up a database on my local machine using phpmyadmin.
	// The database is called testDb
	db, err := sql.Open("mysql", "admin:herb123456@tcp(database-1.cgn5wkvotooq.ap-east-1.rds.amazonaws.com:3306)/irepair")

	//db, err := sql.Open("mysql", "master:xHi52E5R09aKMRr6blFH@tcp(localhost:3306)/irepair")

	// if there is an error opening the connection, handle it
	if err != nil {
		panic(err.Error())
	}

	// defer the close till after the main function has finished
	// executing
	defer db.Close()

	// perform a db.Query insert
	insertTime := stampToDate(strconv.FormatInt(msg.Time, 10))
	insert, err := db.Query("INSERT INTO Messages (author, avatar, message, image, time, id, username) VALUES (?, ?, ?, ?, ?, null, ?)",
		msg.Author, msg.Avatar, msg.Message, msg.Image, insertTime, msg.ID)

	// if there is an error inserting, handle it
	if err != nil {
		panic(err.Error())
	}
	// be careful deferring Queries if you are using transactions
	defer insert.Close()
}

func stampToDate(strTime string) string {

	formatTime := strTime
	if len(strTime) > 10 {
		formatTime = strTime[0:10]
	}

	i, _ := strconv.ParseInt(formatTime, 10, 64)
	tm := time.Unix(i, 0)

	return tm.Format("2006-01-02 15:04:05")
}
