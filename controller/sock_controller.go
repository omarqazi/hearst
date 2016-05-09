package controller

import (
	"crypto/rsa"
	"errors"
	"github.com/gorilla/websocket"
	"github.com/omarqazi/hearst/auth"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"strconv"
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
			err = sc.HandleRead(req, responses)
		case "update":
			err = sc.HandleUpdate(req, responses)
		case "delete":
			err = sc.HandleDelete(req, responses)
		case "list":
			err = sc.HandleList(req, responses)
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
	rid := req.Request["rid"]

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
			responses <- map[string]string{"error": "error adding thread member to new thread", "rid": rid}
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

	go func() {
		if !req.Client.CanWrite(dbo.PermissionThreadId()) {
			responses <- map[string]string{"error": "you do not have permission to create this object", "rid": rid}
			return
		}

		if insertErr := dbo.Insert(); insertErr != nil {
			responses <- map[string]string{"error": "could not create object", "rid": rid}
			return
		}

		if len(rid) > 0 {
			responses <- map[string]interface{}{
				"rid":     rid,
				"payload": dbo,
			}
		} else {
			responses <- dbo
		}
	}()

	return
}

func (sc SockController) HandleRead(req SockRequest, responses chan interface{}) (err error) {
	var dbo datastore.Recordable

	switch req.Request["model"] {
	case "mailbox":
		dbo = &datastore.Mailbox{Record: datastore.Rec(req.Request["id"])}
	case "thread":
		dbo = &datastore.Thread{Record: datastore.Rec(req.Request["id"])}
	case "message":
		dbo = &datastore.Message{Id: req.Request["id"]}
	case "threadmember":
		dbo = &datastore.ThreadMember{MailboxId: req.Request["mailbox_id"], ThreadId: req.Request["thread_id"]}
	default:
		return errors.New("Error during read: invalid model type")
	}

	go func() {
		if loadErr := dbo.Load(); loadErr != nil {
			responses <- map[string]string{"error": "unable to load object from datastore"}
			return
		}

		if req.Client.CanRead(dbo.PermissionThreadId()) {
			responses <- dbo
		} else {
			responses <- map[string]string{"error": "client not authorized to read this object"}
		}
	}()

	return
}

func (sc SockController) HandleList(req SockRequest, responses chan interface{}) (err error) {
	switch req.Request["model"] {
	case "thread":
		err = sc.HandleListThread(req, responses)
	case "threadmember":
		err = sc.HandleListThreadMember(req, responses)
	}
	return
}

func (sc SockController) HandleListThread(req SockRequest, responses chan interface{}) (err error) {
	go func() {
		threadId := req.Request["id"]

		thread, err := datastore.GetThread(threadId)
		if err != nil {
			responses <- map[string]string{"error": "thread not found"}
			return
		}

		if !req.Client.CanRead(thread.Id) {
			responses <- map[string]string{"error": "not authorized to list thread", "thread_id": threadId}
			return
		}

		limitString, ok := req.Request["limit"]
		limit, err := strconv.Atoi(limitString)
		if !ok || err != nil {
			limit = 50
		}

		topic := req.Request["topic"]

		messages, err := thread.RecentMessagesWithTopic(topic, limit)
		if err != nil {
			responses <- map[string]string{"error": "error retrieving recent messages", "thread_id": thread.Id}
			return
		}

		responses <- messages

		shouldFollow := req.Request["follow"] == "true"
		var changeEvents chan datastore.Event
		if shouldFollow {
			changeEvents = datastore.Stream.EventChannel("message-insert-" + thread.Id)
		}

		if shouldFollow && req.Client.CanFollow(thread.Id) {
			for evt := range changeEvents {
				responses <- []datastore.Event{evt}
			}
		}
	}()

	return
}

func (sc SockController) HandleListThreadMember(req SockRequest, responses chan interface{}) (err error) {
	go func() {
		threadId, hasThreadId := req.Request["thread_id"]
		mailboxId, hasMailboxId := req.Request["mailbox_id"]

		if hasThreadId {
			thread := datastore.Thread{Record: datastore.Rec(threadId)}
			members, err := thread.GetAllMembers()
			if err != nil {
				responses <- map[string]string{"error": "unable to get members for thread", "thread_id": threadId}
			} else {
				responses <- members
			}
		} else if hasMailboxId {
			mailbox := datastore.Mailbox{Record: datastore.Rec(mailboxId)}
			members, err := mailbox.GetAllThreads()
			if err != nil {
				responses <- map[string]string{"error": "unable to get threads for mailbox", "mailbox_id": mailboxId}
			} else {
				responses <- members
			}
		} else {
			responses <- map[string]string{"error": "neither thread_id nor mailbox_id required"}
		}
	}()

	return
}

func (sc SockController) HandleUpdate(req SockRequest, responses chan interface{}) (err error) {
	var dbo datastore.Recordable

	switch req.Request["model"] {
	case "mailbox":
		dbo = &datastore.Mailbox{}
	case "thread":
		dbo = &datastore.Thread{}
	case "message":
		dbo = &datastore.Message{}
	case "threadmember":
		dbo = &datastore.ThreadMember{}
	}

	if err = req.Conn.ReadJSON(&dbo); err != nil {
		return
	}

	go func() {
		if req.Client.CanWrite(dbo.PermissionThreadId()) {
			if updateErr := dbo.Update(); updateErr != nil {
				responses <- map[string]string{"error": "could not update object"}
				return
			}
		}

		if loadErr := dbo.Load(); loadErr == nil {
			responses <- dbo
		}
	}()

	return
}

func (sc SockController) HandleDelete(req SockRequest, responses chan interface{}) (err error) {
	var dbo datastore.Recordable

	switch req.Request["model"] {
	case "mailbox":
		dbo = &datastore.Mailbox{Record: datastore.Rec(req.Request["id"])}
	case "thread":
		dbo = &datastore.Thread{Record: datastore.Rec(req.Request["id"])}
	case "message":
		dbo = &datastore.Message{Id: req.Request["id"]}
	case "threadmember":
		dbo = &datastore.ThreadMember{MailboxId: req.Request["mailbox_id"], ThreadId: req.Request["thread_id"]}
	default:
		return errors.New("Error during read: invalid model type")
	}

	go func() {
		if req.Client.CanWrite(dbo.PermissionThreadId()) {
			if deleteErr := dbo.Delete(); deleteErr != nil {
				responses <- map[string]string{"error": "could not delete object"}
			}
			responses <- dbo
			return
		} else {
			responses <- map[string]string{"error": "not authorized to delete object"}
		}
	}()

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
