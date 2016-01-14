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
	_, err := authorizedMailbox(r)
	if err != nil {
		http.Error(w, "session token invalid", 403)
		return
	}

	switch r.Method {
	case "GET":
		mc.GetMessage(rid(r), w, r)
	case "POST":
		mc.PostMessage(w, r)
	default:
		mc.HandleUnknown(w, r)
	}
}

const messageLimit = 500

func (mc MessageController) GetMessage(tid string, w http.ResponseWriter, r *http.Request) {
	thread, err := datastore.GetThread(tid)
	if err != nil {
		http.Error(w, "thread not found", 404)
		return
	}

	recentMessages, err := thread.RecentMessages(messageLimit)
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

func (mc MessageController) PostMessage(w http.ResponseWriter, r *http.Request) {
	var message datastore.Message
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&message); err != nil {
		http.Error(w, "error parsing request body", 400)
		return
	}

	if message.ThreadId == "" {
		message.ThreadId = rid(r)
	}

	if err := message.Insert(); err != nil {
		http.Error(w, "error inserting message", 500)
		return
	}

	mc.GetMessage(message.ThreadId, w, r)
}

func (mc MessageController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
