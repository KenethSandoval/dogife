package main

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// ClientManager monitoring of connected clients, new registrations
// disconnected and message from point to point
type ClientManager struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
}

// Client unique identifier, socket connection and message waiting `send`
type Client struct {
	id     string
	socket *websocket.Conn
	send   chan []byte
}

// Message json format
type Message struct {
	Sender    string `json:"sender,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Content   string `json:"content,omitempty"`
}

var manager = ClientManager{
	broadcast:  make(chan []byte),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	clients:    make(map[*Client]bool),
}

// start if the `manager.register` has data it will be added to the client map
// after adding it a message is sent in json format to the other clients
// without including the connected client
//
// If the client is disconnected, he `manager.unregister` have data and the client's
// data will be closed in the channel and will be deleted from the `ClientManager`
// and send a json
//
// If the `manager.broadcast` has data, it means that we are trying
// to send and receive messages.
// Let's walk through each managed client sending the message to each of them.
// If for some reason the channel is clogged or the message cannot be sent
// we assume the client has been disconnected and remove it.
func (manager *ClientManager) start() {
	for {
		select {
		case conn := <-manager.register:
			manager.clients[conn] = true
			jsonMessage, _ := json.Marshal(&Message{Content: "/A new socket has connected."})
			manager.send(jsonMessage, conn)
		case conn := <-manager.unregister:
			if _, ok := manager.clients[conn]; ok {
				close(conn.send)
				delete(manager.clients, conn)
				jsonMessage, _ := json.Marshal(&Message{Content: "/A socket has disconnected."})
				manager.send(jsonMessage, conn)
			}
		case message := <-manager.broadcast:
			for conn := range manager.clients {
				select {
				case conn.send <- message:
				default:
					close(conn.send)
					delete(manager.clients, conn)
				}
			}
		}
	}
}

func (manager *ClientManager) send(message []byte, ignore *Client) {

}
