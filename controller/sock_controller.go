package controller

import (
	"crypto/rsa"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/omarqazi/hearst/auth"
	"github.com/omarqazi/hearst/datastore"
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

type SockRequest struct {
	Conn        *websocket.Conn    // WebSocket connection object
	Request     map[string]string  // Sock controller request -- tells sock controller what to do
	HTTPRequest *http.Request      // HTTP request used to establish the sock connection
	Client      *datastore.Mailbox // The authorized mailbox of the client that is making the request
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

	mb, err := sc.IdentifyClient(conn)
	if err != nil {
		return
	}

	mb.StillConnected()
	conn.SetPongHandler(func(appData string) error {
		mb.StillConnected()
		return nil
	})

	go sc.HandleReads(conn, responses, r, &mb)
	sc.HandleWrites(conn, responses, time.Tick(pingTime), &mb)
}

// Function HandleReads reads json requests from the socket, processes them
// through the apropriate controller, and sends the response to the client
func (sc SockController) HandleReads(conn *websocket.Conn, responses chan interface{}, r *http.Request, mb *datastore.Mailbox) (err error) {
	defer close(responses)
	var request map[string]string

	req := SockRequest{
		Conn:        conn,
		Request:     request,
		HTTPRequest: r,
		Client:      mb,
	}

	for {
		if err = conn.ReadJSON(&request); err != nil {
			return
		}
		req.Request = request

		mb.StillConnected()

		switch request["action"] {
		case "create":
			err = sc.HandleCreate(req, responses)
		case "read":
			responses <- map[string]string{"ye": "reading"}
		case "update":
			responses <- map[string]string{"ye": "updating"}
		case "delete":
			responses <- map[string]string{"ye": "deleting"}
		default:
			responses <- map[string]string{"error": "invalid action"}
		}

		if err != nil {
			responses <- map[string]string{"error": err.Error()}
			return
		}
	}
}

func (sc SockController) HandleCreate(req SockRequest, responses chan interface{}) (err error) {
	var dbo datastore.Recordable

	switch req.Request["model"] {
	case "mailbox":
		dbo = &datastore.Mailbox{}
	case "thread":
		thread := datastore.Thread{}
		thread.RequireId()
		dbo = &thread
		adminMember := &datastore.ThreadMember{
			ThreadId:          thread.Id,
			MailboxId:         req.Client.Id,
			AllowRead:         true,
			AllowWrite:        true,
			AllowNotification: true,
		}

		if memberErr := thread.AddMember(adminMember); memberErr != nil {
			responses <- map[string]string{"error": "error adding thread member to new thread"}
			return
		}
	case "message":
		dbo = &datastore.Message{}
	case "threadmember":
		dbo = &datastore.ThreadMember{}
	default:
		return errors.New("Error during create: invalid model type")
	}

	if err = req.Conn.ReadJSON(&dbo); err != nil {
		return
	}

	if !req.Client.CanWrite(dbo.PermissionThreadId()) {
		responses <- map[string]string{"error": "you do not have permission to create this object"}
		return
	}

	if insertErr := dbo.Insert(); insertErr != nil {
		responses <- map[string]string{"error": "could not create object"}
		return
	}

	responses <- dbo

	return
}

// Function HandleWrites coordinates all write operations on the socket by
// listening to multiple channels and writing any received data
func (sc SockController) HandleWrites(conn *websocket.Conn, jsonWrites <-chan interface{}, pingWrites <-chan time.Time, mb *datastore.Mailbox) (err error) {
	for err = nil; err == nil; { // Loop until there's an error (like client close)
		mb.StillConnected()
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

// Function IdentifyClient attempts to identify the user over a web socket connection
// It must have exclusive access to io on the connection until it returns
func (sc SockController) IdentifyClient(conn *websocket.Conn) (mb datastore.Mailbox, err error) {
	var authRequest map[string]string
	if err = conn.ReadJSON(&authRequest); err != nil { // read auth request from connection
		conn.WriteJSON(map[string]string{"error": "client failed to identify itself"})
		return
	}

	switch authRequest["auth"] {
	case "session":
		mailboxId := authRequest["mailbox"]
		token := authRequest["token"]

		if mb, err = datastore.GetMailbox(mailboxId); err != nil {
			return
		}

		pubKey, er := auth.PublicKeyFromString(mb.PublicKey)
		if er != nil {
			return mb, er
		}

		session, erx := auth.ParseSession(token)
		if erx != nil {
			return mb, erx
		}

		err = session.Valid(pubKey, &serverSessionKey.PublicKey)
	case "temp", "new":
		var key *rsa.PrivateKey
		mb, key, err = datastore.NewMailboxWithKey()
		if err != nil {
			return
		}

		token, erx := mb.SessionToken(24*time.Hour, key, serverSessionKey)
		if erx != nil {
			err = erx
			return
		}

		if err = mb.Insert(); err != nil {
			return
		}
		authResponse := map[string]string{"mailbox_id": mb.Id, "session_token": token}
		if authRequest["auth"] == "new" {
			authResponse["private_key"] = auth.StringForPrivateKey(key)
		}
		conn.WriteJSON(authResponse)
	default:
		err = errors.New("invalid auth type")
		conn.WriteJSON(map[string]string{"error": "invalid auth type"})
	}

	return
}

// Function HandlePong is called when the client responds to a ping
func (sc SockController) HandlePong(appData string) error {
	return nil
}