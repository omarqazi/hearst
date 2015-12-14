package controller

import (
	"encoding/json"
	"fmt"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
)

type ThreadController struct {
}

func (tc ThreadController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tc.GetThread(rid(r), w, r)
	default:
		tc.HandleUnknown(w, r)
	}
}

func (tc ThreadController) GetThread(tid string, w http.ResponseWriter, r *http.Request) {
	thread, err := datastore.GetThread(tid)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "thread not found")
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	if err := encoder.Encode(thread); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "Error marshaling response json")
		return
	}
}

func (tc ThreadController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
