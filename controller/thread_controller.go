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
	if subcat := urlSubcategory(r); subcat == "members" {
		tc.RouteThreadMembersRequest(w, r)
	} else {
		tc.RouteThreadRequest(w, r)
	}
}

func (tc ThreadController) RouteThreadRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tc.GetThread(rid(r), w, r)
	case "POST":
		tc.PostThread(w, r)
	case "PUT":
		tc.PutThread(w, r)
	case "DELETE":
		tc.DeleteThread(w, r)
	default:
		tc.HandleUnknown(w, r)
	}
}

func (tc ThreadController) RouteThreadMembersRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tc.GetThreadMembers(rid(r), w, r)
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

func (tc ThreadController) PostThread(w http.ResponseWriter, r *http.Request) {
	var thread datastore.Thread
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&thread); err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, "invalid request JSON")
		return
	}

	if err := thread.Insert(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "error saving thread")
		return
	}

	tc.GetThread(thread.Id, w, r)
}

func (tc ThreadController) PutThread(w http.ResponseWriter, r *http.Request) {
	var thread datastore.Thread
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&thread); err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, "invalid request JSON")
		return
	}

	if thread.Id == "" {
		thread.Id = rid(r)
	}

	dbThread, err := datastore.GetThread(thread.Id)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "thread not found")
		return
	}

	if thread.Subject == "" {
		thread.Subject = dbThread.Subject
	}

	if err := thread.Update(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "error updating thread")
		return
	}

	tc.GetThread(thread.Id, w, r)
}

func (tc ThreadController) DeleteThread(w http.ResponseWriter, r *http.Request) {
	thread := datastore.Thread{Id: rid(r)}
	if err := thread.Delete(); err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "thread not found")
		return
	}

	fmt.Fprintln(w, "thread deleted")
}

func (tc ThreadController) GetThreadMembers(tid string, w http.ResponseWriter, r *http.Request) {
	thread, err := datastore.GetThread(tid)
	if err != nil {
		http.Error(w, "thread not found", 404)
		return
	}

	comps := pathComponents(r)
	var outputValue interface{}
	if len(comps) > 2 { // Requesting specific member
		mailboxId := comps[2]
		outputValue, err = thread.GetMember(mailboxId)

	} else {
		outputValue, err = thread.GetAllMembers()
	}
	if err != nil {
		http.Error(w, "error getting thread members", 500)
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	if err := encoder.Encode(outputValue); err != nil {
		http.Error(w, "error marshaling response json", 500)
		return
	}
}

func (tc ThreadController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
