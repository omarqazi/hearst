package postoffice

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type Client struct {
	Domain     string
	httpClient *http.Client
	Socket     *websocket.Conn
}

func NewClient(domain string) (c Client, err error) {
	c.Domain = domain
	return
}

func (c *Client) ConnectSocket() (err error) {
	socketUrl := "ws://" + c.Domain + "/sock/"
	var dialer *websocket.Dialer

	c.Socket, _, err = dialer.Dial(socketUrl, nil)
	return
}

func (c *Client) EnsureConnected() (err error) {
	if c.Socket == nil {
		err = c.ConnectSocket()
	}
	return
}

func (c *Client) Authenticate(authPayload map[string]string) (result map[string]string, err error) {
	c.EnsureConnected()

	if err = c.Socket.WriteJSON(authPayload); err != nil {
		return
	}

	err = c.Socket.ReadJSON(&result)
	return
}
