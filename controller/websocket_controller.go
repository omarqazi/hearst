package controller

import (
	"github.com/gorilla/websocket"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"strconv"
	"time"
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
	defer conn.Close()

	broadcastChannel := make(chan interface{}, 10)
	go wsc.ProcessCommands(conn, broadcastChannel)
	conn.SetPongHandler(func(appData string) error {
		return nil
	})

	ticker := time.Tick(15 * time.Second)

	for {
		select {
		case responseItem := <-broadcastChannel:
			if err = conn.WriteJSON(responseItem); err != nil {
				return
			}
		case <-ticker:
			deadline := time.Now().Add(15 * time.Second)
			if err = conn.WriteControl(websocket.PingMessage, []byte{}, deadline); err != nil {
				return
			}
		}
	}
}

func (wsc WebSocketController) ProcessCommands(conn *websocket.Conn, broadcast chan interface{}) {
	defer conn.Close()

	for {
		var request map[string]string
		if err := conn.ReadJSON(&request); err != nil {
			return
		}

		if request["model"] == "mailbox" { // If mailbox
			wsc.HandleMailbox(request, conn, broadcast)
		} else if request["model"] == "thread" {
			wsc.HandleThread(request, conn, broadcast)
		} else if request["model"] == "message" {
			wsc.HandleMessage(request, conn, broadcast)
		} else if request["model"] == "threadmember" {
			wsc.HandleThreadMember(request, conn, broadcast)
		} else { // if request type unknown
			wsc.UnknownRequest(request, conn, broadcast)
		}
	}
}

func (wsc WebSocketController) HandleMailbox(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	var mailbox datastore.Mailbox

	if _, ok := request["uuid"]; ok { // Get by UUID
		go wsc.GetMailbox(request, conn, broadcast)
	} else if action, ok := request["action"]; ok && action == "insert" { // Insert
		if err := conn.ReadJSON(&mailbox); err != nil {
			return
		}
		go wsc.InsertMailbox(request, conn, broadcast, mailbox)
	} else if action == "update" { // Insert
		if err := conn.ReadJSON(&mailbox); err != nil {
			return
		}
		go wsc.UpdateMailbox(request, conn, broadcast, mailbox)
	} else if action == "delete" {
		go wsc.DeleteMailbox(request, conn, broadcast)
	} else { // if not enough information
		wsc.ErrorResponse("invalid mailbox request", conn, broadcast)
	}
}

func (wsc WebSocketController) HandleThread(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	var thread datastore.Thread

	if _, ok := request["uuid"]; ok { // Get by UUID
		go wsc.GetThread(request, conn, broadcast)
	} else if action, ok := request["action"]; ok && action == "insert" {
		if err := conn.ReadJSON(&thread); err != nil {
			return
		}
		go wsc.InsertThread(request, conn, broadcast, thread)
	} else if action == "update" {
		if err := conn.ReadJSON(&thread); err != nil {
			return
		}
		go wsc.UpdateThread(request, conn, broadcast, thread)
	} else if action == "delete" {
		wsc.DeleteThread(request, conn, broadcast)
	} else if action == "list" {
		go wsc.ListThread(request, conn, broadcast)
	} else {
		wsc.ErrorResponse("invalid thread request", conn, broadcast)
	}
}

func (wsc WebSocketController) HandleMessage(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	var message datastore.Message
	if _, ok := request["uuid"]; ok {
		go wsc.GetMessage(request, conn, broadcast)
	} else if action, ok := request["action"]; ok && action == "insert" {
		if err := conn.ReadJSON(&message); err != nil {
			return
		}
		go wsc.InsertMessage(request, conn, broadcast, message)
	}
}

func (wsc WebSocketController) HandleThreadMember(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	_, threadOk := request["thread_id"]
	_, mailboxOk := request["mailbox_id"]
	action, ok := request["action"]
	var member datastore.ThreadMember

	if threadOk && mailboxOk && ok {
		switch action {
		case "get":
			go wsc.GetThreadMember(request, conn, broadcast)
		case "insert":
			if err := conn.ReadJSON(&member); err != nil {
				return
			}
			go wsc.InsertThreadMember(request, conn, broadcast, member)
		case "update":
			if err := conn.ReadJSON(&member); err != nil {
				return
			}
			go wsc.UpdateThreadMember(request, conn, broadcast, member)
		case "delete":
			go wsc.DeleteThreadMember(request, conn, broadcast)
		default:
			wsc.ErrorResponse("invalid action", conn, broadcast)
		}
	} else {
		wsc.ErrorResponse("thread_id, mailbox_id and action required", conn, broadcast)
	}
}

func (wsc WebSocketController) ListThread(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	threadId, ok := request["thread_id"]
	if !ok {
		wsc.ErrorResponse("thread id required", conn, broadcast)
		return
	}

	thread, err := datastore.GetThread(threadId)
	if err != nil {
		wsc.ErrorResponse("thread not found", conn, broadcast)
		return
	}

	limitString, ok := request["limit"]
	limit, err := strconv.Atoi(limitString)
	if !ok || err != nil {
		limit = 50
	}

	messages, err := thread.RecentMessages(limit)
	if err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	followString, ok := request["follow"]
	shouldFollow := ok && followString == "true"
	var changeEvents chan datastore.Event
	if shouldFollow {
		changeEvents = datastore.Stream.EventChannel("message-insert-" + thread.Id)
	}

	wo(broadcast, messages)

	if shouldFollow {
		for evt := range changeEvents {
			if ok := wo(broadcast, []datastore.Event{evt}); !ok {
				return
			}
		}
	}

	return
}

func (wsc WebSocketController) GetMailbox(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	mb, err := datastore.GetMailbox(request["uuid"])
	if err != nil { // If mailbox not found
		wsc.ErrorResponse("not found", conn, broadcast)
		return
	}
	wo(broadcast, mb)
	return
}

func (wsc WebSocketController) GetThread(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	thread, err := datastore.GetThread(request["uuid"])
	if err != nil {
		wsc.ErrorResponse("not found", conn, broadcast)
		return
	}
	wo(broadcast, thread)
	return
}

func (wsc WebSocketController) GetMessage(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	message, err := datastore.GetMessage(request["uuid"])
	if err != nil {
		wsc.ErrorResponse("not found", conn, broadcast)
		return
	}
	wo(broadcast, message)
	return
}

func (wsc WebSocketController) GetThreadMember(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		wsc.ErrorResponse("thread not found", conn, broadcast)
		return
	}

	member, err := thread.GetMember(request["mailbox_id"])
	if err != nil {
		wsc.ErrorResponse("member not found", conn, broadcast)
		return
	}
	wo(broadcast, member)
	return
}

func (wsc WebSocketController) InsertMailbox(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, mailbox datastore.Mailbox) {
	if err := mailbox.Insert(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	mb, erx := datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}
	wo(broadcast, mb)
	return
}

func (wsc WebSocketController) InsertThread(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, thread datastore.Thread) {
	if err := thread.Insert(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	tr, erx := datastore.GetThread(thread.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}
	wo(broadcast, tr)
}

func (wsc WebSocketController) InsertMessage(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, message datastore.Message) {
	if err := message.Insert(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	msg, erx := datastore.GetMessage(message.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}

	wo(broadcast, msg)
	return
}

func (wsc WebSocketController) InsertThreadMember(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, member datastore.ThreadMember) {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		wsc.ErrorResponse("thread not found", conn, broadcast)
		return
	}

	mailbox, err := datastore.GetMailbox(request["mailbox_id"])
	if err != nil {
		wsc.ErrorResponse("mailbox not found", conn, broadcast)
		return
	}

	member.MailboxId = mailbox.Id
	if err = thread.AddMember(&member); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	wo(broadcast, member)
	return
}

func (wsc WebSocketController) UpdateMailbox(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, mailbox datastore.Mailbox) {
	mb, erx := datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}

	if mailbox.PublicKey == "" {
		mailbox.PublicKey = mb.PublicKey
	}

	if mailbox.DeviceId == "" {
		mailbox.DeviceId = mb.DeviceId
	}

	if err := mailbox.Update(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	mb, erx = datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}
	wo(broadcast, mb)
	return
}

func (wsc WebSocketController) UpdateThread(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, thread datastore.Thread) {
	tr, erx := datastore.GetThread(thread.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}

	if thread.Subject == "" {
		thread.Subject = tr.Subject
	}

	if err := thread.Update(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}

	tr, erx = datastore.GetThread(thread.Id)
	if erx != nil {
		wsc.ErrorResponse(erx.Error(), conn, broadcast)
		return
	}
	wo(broadcast, tr)
	return
}

func (wsc WebSocketController) UpdateThreadMember(request map[string]string, conn *websocket.Conn, broadcast chan interface{}, mem datastore.ThreadMember) {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		wsc.ErrorResponse("thread not found", conn, broadcast)
		return
	}

	member, err := thread.GetMember(request["mailbox_id"])
	if err != nil {
		wsc.ErrorResponse("member not found", conn, broadcast)
		return
	}

	mem.ThreadId = member.ThreadId
	mem.MailboxId = member.MailboxId

	if err = mem.UpdatePermissions(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}
	wo(broadcast, mem)
	return
}

func (wsc WebSocketController) DeleteMailbox(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	if uuid, ok := request["delete_mailbox"]; ok {
		mailbox := datastore.Mailbox{Id: uuid}
		if err := mailbox.Delete(); err != nil {
			wsc.ErrorResponse(err.Error(), conn, broadcast)
			return
		}

		wo(broadcast, mailbox)
		return
	}

	wsc.ErrorResponse("provide id to delete in delete_mailbox parameter", conn, broadcast)
	return
}

func (wsc WebSocketController) DeleteThread(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	if uuid, ok := request["delete_thread"]; ok {
		thread := datastore.Thread{Id: uuid}
		if err := thread.Delete(); err != nil {
			wsc.ErrorResponse(err.Error(), conn, broadcast)
			return
		}

		wo(broadcast, thread)
		return
	}

	wsc.ErrorResponse("provide id to delete in delete_thread parameter", conn, broadcast)
	return
}

func (wsc WebSocketController) DeleteThreadMember(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		wsc.ErrorResponse("thread not found", conn, broadcast)
		return
	}

	member, err := thread.GetMember(request["mailbox_id"])
	if err != nil {
		wsc.ErrorResponse("member not found", conn, broadcast)
		return
	}

	if err := member.Remove(); err != nil {
		wsc.ErrorResponse(err.Error(), conn, broadcast)
		return
	}
	wo(broadcast, member)
	return
}

func (wsc WebSocketController) ErrorResponse(message string, conn *websocket.Conn, broadcast chan interface{}) {
	wo(broadcast, map[string]string{"error": message})
}

func (wsc WebSocketController) UnknownRequest(request map[string]string, conn *websocket.Conn, broadcast chan interface{}) {
	wo(broadcast, request)
}

func wo(broadcast chan interface{}, obj interface{}) bool {
	select {
	case broadcast <- obj:
		return true
	default:
		return false
	}
}
