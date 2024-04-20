package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var addr = flag.String("addr", ":8081", "http service address")

type Input struct {
	AgentId string `json:"agentId"`
	Input   string `json:"input"`
}
type ApiHandler struct {
	Hub *Hub
}

func (ah *ApiHandler) GetAgentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	agentsConnected := ""
	keys := reflect.ValueOf(ah.Hub.clients).MapKeys()
	if len(keys) > 0 {
		agentsConnected = fmt.Sprintf("%v agents(s) online: \n", len(keys))
		strkeys := make([]string, len(keys))
		for i := 0; i < len(keys); i++ {
			strkeys[i] = "`" + keys[i].String() + "`"
		}
		//fmt.Print(strings.Join(strkeys, ","))
		agentsConnected += strings.Join(strkeys, ",")
	} else {
		agentsConnected = "No agents connected"
	}

	fmt.Fprintf(w, agentsConnected)
}

func (ah *ApiHandler) InputHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Unmarshal the JSON into a Message struct
	var input Input
	if err := json.Unmarshal(body, &input); err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	// Print the message to the server console
	fmt.Printf("Received input: %s\n", input.Input)
	fmt.Printf("Received agent: %s\n", input.AgentId)

	encodedTransferPacket := encodeTransferPacket("command", input.Input)
	ah.Hub.sendTarget <- Message{clientId: input.AgentId, data: encodedTransferPacket}

	// Respond to the client
	fmt.Fprintf(w, "Received: %s", input.Input)
}

func main() {
	flag.Parse()

	router := mux.NewRouter()
	hub := newHub()
	go hub.run()
	myApiHandler := &ApiHandler{Hub: hub}
	router.HandleFunc("/agents", myApiHandler.GetAgentsHandler).Methods("GET")
	router.HandleFunc("/input", myApiHandler.InputHandler).Methods("POST")

	router.HandleFunc("/ws/{clientId}", func(w http.ResponseWriter, r *http.Request) {
		clientId := mux.Vars(r)["clientId"]
		serveWs(hub, w, r, clientId)
	})
	server := &http.Server{
		Handler:           router,
		Addr:              *addr,
		ReadHeaderTimeout: 3 * time.Second,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
