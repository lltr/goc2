package main

import (
	"flag"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// Bot parameters
var (
	addr      *string
	guildId   *string
	channelId *string
	botToken  *string
)

func init() {
	addr = flag.String("addr", ":8081", "http service address")

	if len(os.Args) > 3 {
		tokenArg := os.Args[1]
		botToken = flag.String("token", tokenArg, "Bot access token")
		guildIdArg := os.Args[2]
		guildId = flag.String("guild", guildIdArg, "Guild ID")
		channelIdArg := os.Args[3]
		channelId = flag.String("channel", channelIdArg, "Channel ID for response")

	} else {
		log.Fatalln("Error missing args: <bot_token> <guild_id> <channel_id>")
	}
}

var s *discordgo.Session

func main() {
	flag.Parse()

	s, err := discordgo.New("Bot " + *botToken)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}

	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, *guildId, v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

	defer s.Close()

	// start http stuff
	router := mux.NewRouter()
	hub := newHub(s)
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

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, myApiHandler)
		}
	})
	log.Println("Commands loaded!")

	err = server.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	log.Println("Press Ctrl+C to exit")
	<-stop

	log.Println("Gracefully shutting down.")
	s.Close()
}
