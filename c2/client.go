package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub                        *Hub
	id                         string
	conn                       *websocket.Conn
	send                       chan []byte
	discordSendMessageRequests chan []byte
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		outputTransferPacket := decodeTransferPacket(message)
		//formattedCmdReply := fmt.Sprintf("%s - %s", c.id, outputTransferPacket.Payload)
		formattedCmdReplyForDiscord := fmt.Sprintf("%s ```%s```", c.id, outputTransferPacket.Payload)

		c.discordSendMessageRequests <- []byte(formattedCmdReplyForDiscord)
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()
	for {
		select {
		case message, _ := <-c.send:
			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) requestPump() {
	for {
		select {
		case discordSendMessageRequest := <-c.discordSendMessageRequests:
			log.Println(string(discordSendMessageRequest))
			_, err := c.hub.s.ChannelMessageSend(*channelId, string(discordSendMessageRequest))
			if err != nil {
				log.Printf("Error sending discord msg: %v", err)
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request, clientId string) {
	log.Printf("%s connected", clientId)
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, id: clientId, conn: conn, send: make(chan []byte, 256), discordSendMessageRequests: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
	go client.requestPump()
}
