package main

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

type Message struct { // struct is used internally between channels, sent to sort before converting to WSPacket
	clientId string
	data     []byte
}

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[string]*Client

	// Inbound messages from the clients.
	broadcast chan []byte

	sendTarget chan Message

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	s *discordgo.Session
}

func newHub(s *discordgo.Session) *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		sendTarget: make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[string]*Client),
		s:          s,
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client.id] = client
		case client := <-h.unregister:
			if _, ok := h.clients[client.id]; ok {
				delete(h.clients, client.id)
				close(client.send)
				log.Printf("%s disconnected", client.id)
			}
		case message := <-h.broadcast:
			for _, client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client.id)
				}
			}
		case message := <-h.sendTarget:
			client, ok := h.clients[message.clientId]
			if ok {
				select {
				case client.send <- message.data:
				default:
					delete(h.clients, client.id)
					close(client.send)
				}
			}
		}

	}
}
