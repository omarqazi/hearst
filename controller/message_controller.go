package controller

import (
	"encoding/json"
	"fmt"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
)

type MessageController struct {
}

func (mc MessageController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mb, err := authorizedMailbox(r)
	if err != nil {
		http.Error(w, "session token invalid", 403)
		return
	}

	switch r.Method {
	case "GET":
		mc.GetMessage(rid(r), w, r, &mb)
	case "POST":
		mc.PostMessage(w, r, &mb)
	default:
		mc.HandleUnknown(w, r)
	}
}

const messageLimit = 500

func (mc MessageController) GetMessage(tid string, w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	thread, err := datastore.GetThread(tid)
	if err != nil {
		http.Error(w, "thread not found", 404)
		return
	}

	userThreadMember, err := thread.GetMember(mb.Id)
	if err != nil || !userThreadMember.AllowRead {
		http.Error(w, "acess denied: not thread member", 403)
		return
	}

	topic := r.URL.Query().Get("topic")
	if topic == "" {
		topic = "%"
	}

	recentMessages, err := thread.RecentMessagesWithTopic(topic, messageLimit)
	if err != nil {
		http.Error(w, "error finding recent messages", 500)
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	if err := encoder.Encode(recentMessages); err != nil {
		http.Error(w, "error marshaling response JSON", 500)
	}
}

func (mc MessageController) PostMessage(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	var message datastore.Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		http.Error(w, "error parsing request body", 400)
		return
	}

	if message.ThreadId == "" {
		message.ThreadId = rid(r)
	}

	thread := datastore.Thread{
		Id: message.ThreadId,
	}
	userMember, err := thread.GetMember(mb.Id)
	if err != nil || !userMember.AllowWrite {
		http.Error(w, "access denied: not member of thread", 403)
		return
	}

	if err := message.Insert(); err != nil {
		http.Error(w, "error inserting message", 500)
		return
	}

	mc.GetMessage(message.ThreadId, w, r, mb)
}

func (mc MessageController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
