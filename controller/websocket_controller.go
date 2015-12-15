package controller

import (
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocketController struct {
}

func (wsc WebSocketController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "error upgrading connection to WebSocket", 500)
		return
	}

	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			conn.Close()
			return
		}
		if err = conn.WriteMessage(messageType, p); err != nil {
			conn.Close()
			return
		}
	}
}
