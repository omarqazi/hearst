package controller

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const pingTime = time.Duration(15 * time.Second)

var socketizer = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// Struct SockController exposes the Hearst API through a WebSocket interface.
// It parses requests and routes them to the appropriate controller code.
type SockController struct {
}

// Upgrade incoming HTTP connections to WebSocket
func (sc SockController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := socketizer.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "error upgrading connection", 500)
		return
	}
	defer conn.Close()

	conn.SetPongHandler(sc.HandlePong)
	responses := make(chan interface{}, 10)

	go sc.HandleReads(conn, responses, r)
	sc.HandleWrites(conn, responses, time.Tick(pingTime))
}

// Function HandleReads reads json requests from the socket, processes them
// through the apropriate controller, and sends the response to the client
func (sc SockController) HandleReads(conn *websocket.Conn, responses chan interface{}, r *http.Request) (err error) {
	var request map[string]string
	defer close(responses)

	for {
		if err = conn.ReadJSON(&request); err != nil {
			return
		}

		responses <- request
	}
}

// Function HandleWrites coordinates all write operations on the socket by
// listening to multiple channels and writing any received data
func (sc SockController) HandleWrites(conn *websocket.Conn, jsonWrites <-chan interface{}, pingWrites <-chan time.Time) (err error) {
	for err = nil; err == nil; { // Loop until there's an error (like client close)
		select { // listen to all channels
		case j, more := <-jsonWrites: // If we get a json response object, send it down
			if more {
				err = conn.WriteJSON(j)
			} else {
				err = errors.New("write channel closed")
			}
		case <-pingWrites: // if it's time to ping the client, send a ping
			err = conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(pingTime))
		}
	}

	return
}

// Function HandlePong is called when the client responds to a ping
func (sc SockController) HandlePong(appData string) error {
	log.Println("Got pong:", appData)
	return nil
}
