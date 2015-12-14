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
	switch r.Method {
	case "GET":
		mc.GetMessage(rid(r), w, r)
	default:
		mc.HandleUnknown(w, r)
	}
}

func (mc MessageController) GetMessage(mid string, w http.ResponseWriter, r *http.Request) {
	message, err := datastore.GetMessage(mid)
	if err != nil {
		http.Error(w, "message not found", 404)
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	if err := encoder.Encode(message); err != nil {
		http.Error(w, "error marshaling response JSON", 500)
	}
}

func (mc MessageController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
