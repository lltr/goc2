package main

import (
	"flag"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

var addr = flag.String("addr", ":8081", "http service address")

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
