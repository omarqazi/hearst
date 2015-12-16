package controller

import (
	"github.com/gorilla/websocket"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"strconv"
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
		var request map[string]string
		err := conn.ReadJSON(&request)
		if err != nil {
			conn.Close()
			return
		}

		keepConnection := true
		if request["model"] == "mailbox" { // If mailbox
			if _, ok := request["uuid"]; ok { // Get by UUID
				keepConnection = wsc.GetMailbox(request, conn)
			} else if action, ok := request["action"]; ok && action == "insert" { // Insert
				keepConnection = wsc.InsertMailbox(request, conn)
			} else if action == "update" { // Insert
				keepConnection = wsc.UpdateMailbox(request, conn)
			} else if action == "delete" {
				keepConnection = wsc.DeleteMailbox(request, conn)
			} else { // if not enough information
				keepConnection = wsc.ErrorResponse("invalid mailbox request", conn)
			}
		} else if request["model"] == "thread" {
			if _, ok := request["uuid"]; ok { // Get by UUID
				keepConnection = wsc.GetThread(request, conn)
			} else if action, ok := request["action"]; ok && action == "insert" {
				keepConnection = wsc.InsertThread(request, conn)
			} else if action == "update" {
				keepConnection = wsc.UpdateThread(request, conn)
			} else if action == "delete" {
				keepConnection = wsc.DeleteThread(request, conn)
			} else if action == "list" {
				keepConnection = wsc.ListThread(request, conn)
			} else {
				keepConnection = wsc.ErrorResponse("invalid thread request", conn)
			}
		} else if request["model"] == "message" {
			if _, ok := request["uuid"]; ok {
				keepConnection = wsc.GetMessage(request, conn)
			} else if action, ok := request["action"]; ok && action == "insert" {
				keepConnection = wsc.InsertMessage(request, conn)
			}
		} else if request["model"] == "threadmember" {
			_, threadOk := request["thread_id"]
			_, mailboxOk := request["mailbox_id"]
			action, ok := request["action"]

			if threadOk && mailboxOk && ok {
				switch action {
				case "get":
					keepConnection = wsc.GetThreadMember(request, conn)
				case "insert":
					keepConnection = wsc.InsertThreadMember(request, conn)
				case "update":
					keepConnection = wsc.UpdateThreadMember(request, conn)
				case "delete":
					keepConnection = wsc.DeleteThreadMember(request, conn)
				default:
					keepConnection = wsc.ErrorResponse("invalid action", conn)
				}
			} else {
				keepConnection = wsc.ErrorResponse("thread_id, mailbox_id and action required", conn)
			}
		} else { // if request type unknown
			// write back request
			keepConnection = wsc.UnknownRequest(request, conn)
		}

		if !keepConnection {
			conn.Close()
			return
		}
	}
}

func (wsc WebSocketController) ListThread(request map[string]string, conn *websocket.Conn) bool {
	threadId, ok := request["thread_id"]
	if !ok {
		return wsc.ErrorResponse("thread id required", conn)
	}

	thread, err := datastore.GetThread(threadId)
	if err != nil {
		return wsc.ErrorResponse("thread not found", conn)
	}

	limitString, ok := request["limit"]
	limit, err := strconv.Atoi(limitString)
	if !ok || err != nil {
		limit = 50
	}

	messages, err := thread.RecentMessages(limit)
	if err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	followString, ok := request["follow"]
	shouldFollow := ok && followString == "true"
	var changeEvents chan datastore.Event
	if shouldFollow {
		changeEvents = datastore.Stream.EventChannel("message-insert-" + thread.Id)
	}

	if err = conn.WriteJSON(messages); err != nil {
		return false
	}

	if shouldFollow {
		for evt := range changeEvents {
			if err = conn.WriteJSON([]datastore.Event{evt}); err != nil {
				return false
			}
		}
	}

	return true
}

func (wsc WebSocketController) GetMailbox(request map[string]string, conn *websocket.Conn) bool {
	mb, err := datastore.GetMailbox(request["uuid"])
	if err != nil { // If mailbox not found
		return wsc.ErrorResponse("not found", conn)
	}

	// Write mailbox response
	if err = conn.WriteJSON(mb); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) GetThread(request map[string]string, conn *websocket.Conn) bool {
	thread, err := datastore.GetThread(request["uuid"])
	if err != nil {
		return wsc.ErrorResponse("not found", conn)
	}

	if err = conn.WriteJSON(thread); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) GetMessage(request map[string]string, conn *websocket.Conn) bool {
	message, err := datastore.GetMessage(request["uuid"])
	if err != nil {
		return wsc.ErrorResponse("not found", conn)
	}

	if err = conn.WriteJSON(message); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) GetThreadMember(request map[string]string, conn *websocket.Conn) bool {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		return wsc.ErrorResponse("thread not found", conn)
	}

	member, err := thread.GetMember(request["mailbox_id"])
	if err != nil {
		return wsc.ErrorResponse("member not found", conn)
	}

	if err = conn.WriteJSON(member); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) InsertMailbox(request map[string]string, conn *websocket.Conn) bool {
	var mailbox datastore.Mailbox
	err := conn.ReadJSON(&mailbox)
	if err != nil {
		return false
	}

	if err := mailbox.Insert(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	mb, erx := datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if err = conn.WriteJSON(mb); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) InsertThread(request map[string]string, conn *websocket.Conn) bool {
	var thread datastore.Thread
	err := conn.ReadJSON(&thread)
	if err != nil {
		return false
	}

	if err := thread.Insert(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	tr, erx := datastore.GetThread(thread.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if err = conn.WriteJSON(tr); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) InsertMessage(request map[string]string, conn *websocket.Conn) bool {
	var message datastore.Message
	err := conn.ReadJSON(&message)
	if err != nil {
		return false
	}

	if err := message.Insert(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	msg, erx := datastore.GetMessage(message.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if err = conn.WriteJSON(msg); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) InsertThreadMember(request map[string]string, conn *websocket.Conn) bool {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		return wsc.ErrorResponse("thread not found", conn)
	}

	mailbox, err := datastore.GetMailbox(request["mailbox_id"])
	if err != nil {
		return wsc.ErrorResponse("mailbox not found", conn)
	}

	var member datastore.ThreadMember
	err = conn.ReadJSON(&member)
	if err != nil {
		return false
	}
	member.MailboxId = mailbox.Id

	if err = thread.AddMember(&member); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	if err = conn.WriteJSON(member); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) UpdateMailbox(request map[string]string, conn *websocket.Conn) bool {
	var mailbox datastore.Mailbox
	err := conn.ReadJSON(&mailbox)
	if err != nil {
		return false
	}

	mb, erx := datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if mailbox.PublicKey == "" {
		mailbox.PublicKey = mb.PublicKey
	}

	if mailbox.DeviceId == "" {
		mailbox.DeviceId = mb.DeviceId
	}

	if err := mailbox.Update(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	mb, erx = datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if err = conn.WriteJSON(mb); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) UpdateThread(request map[string]string, conn *websocket.Conn) bool {
	var thread datastore.Thread
	err := conn.ReadJSON(&thread)
	if err != nil {
		return false
	}

	tr, erx := datastore.GetThread(thread.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if thread.Subject == "" {
		thread.Subject = tr.Subject
	}

	if err := thread.Update(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	tr, erx = datastore.GetThread(thread.Id)
	if erx != nil {
		return wsc.ErrorResponse(erx.Error(), conn)
	}

	if err = conn.WriteJSON(tr); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) UpdateThreadMember(request map[string]string, conn *websocket.Conn) bool {
	var mem datastore.ThreadMember
	err := conn.ReadJSON(&mem)
	if err != nil {
		return false
	}

	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		return wsc.ErrorResponse("thread not found", conn)
	}

	member, err := thread.GetMember(request["mailbox_id"])
	if err != nil {
		return wsc.ErrorResponse("member not found", conn)
	}

	mem.ThreadId = member.ThreadId
	mem.MailboxId = member.MailboxId

	if err = mem.UpdatePermissions(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	return true
}

func (wsc WebSocketController) DeleteMailbox(request map[string]string, conn *websocket.Conn) bool {
	if uuid, ok := request["delete_mailbox"]; ok {
		mailbox := datastore.Mailbox{Id: uuid}
		if err := mailbox.Delete(); err != nil {
			return wsc.ErrorResponse(err.Error(), conn)
		}

		if err := conn.WriteJSON(mailbox); err != nil {
			return false
		}

		return true
	}

	return wsc.ErrorResponse("provide id to delete in delete_mailbox parameter", conn)
}

func (wsc WebSocketController) DeleteThread(request map[string]string, conn *websocket.Conn) bool {
	if uuid, ok := request["delete_thread"]; ok {
		thread := datastore.Thread{Id: uuid}
		if err := thread.Delete(); err != nil {
			return wsc.ErrorResponse(err.Error(), conn)
		}

		if err := conn.WriteJSON(thread); err != nil {
			return false
		}

		return true
	}

	return wsc.ErrorResponse("provide id to delete in delete_thread parameter", conn)
}

func (wsc WebSocketController) DeleteThreadMember(request map[string]string, conn *websocket.Conn) bool {
	thread, err := datastore.GetThread(request["thread_id"])
	if err != nil {
		return wsc.ErrorResponse("thread not found", conn)
	}

	member, err := thread.GetMember(request["mailbox_id"])
	if err != nil {
		return wsc.ErrorResponse("member not found", conn)
	}

	if err := member.Remove(); err != nil {
		return wsc.ErrorResponse(err.Error(), conn)
	}

	if err := conn.WriteJSON(member); err != nil {
		return false
	}

	return true
}

func (wsc WebSocketController) ErrorResponse(message string, conn *websocket.Conn) bool {
	errorPayload := map[string]string{"error": message}
	if err := conn.WriteJSON(errorPayload); err != nil {
		return false
	}
	return true
}

func (wsc WebSocketController) UnknownRequest(request map[string]string, conn *websocket.Conn) bool {
	if err := conn.WriteJSON(request); err != nil {
		return false
	}

	return true
}
