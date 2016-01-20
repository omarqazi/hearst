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
	mb, err := authorizedMailbox(r)
	if err != nil {
		http.Error(w, "session token invalid", 403)
		return
	}

	if subcat := urlSubcategory(r); subcat == "members" {
		tc.RouteThreadMembersRequest(w, r, &mb)
	} else {
		tc.RouteThreadRequest(w, r, &mb)
	}
}

func (tc ThreadController) RouteThreadRequest(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	switch r.Method {
	case "GET":
		tc.GetThread(rid(r), w, r, mb)
	case "POST":
		tc.PostThread(w, r, mb)
	case "PUT":
		tc.PutThread(w, r, mb)
	case "DELETE":
		tc.DeleteThread(w, r, mb)
	default:
		tc.HandleUnknown(w, r)
	}
}

func (tc ThreadController) RouteThreadMembersRequest(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	switch r.Method {
	case "GET":
		tc.GetThreadMembers(rid(r), w, r, mb)
	case "POST":
		tc.PostThreadMember(w, r, mb)
	case "PUT":
		tc.PutThreadMember(w, r, mb)
	case "DELETE":
		tc.DeleteThreadMember(w, r, mb)
	default:
		tc.HandleUnknown(w, r)
	}
}

func (tc ThreadController) GetThread(tid string, w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	thread, err := datastore.GetThread(tid)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "thread not found")
		return
	}

	member, err := thread.GetMember(mb.Id)
	if err != nil || !member.AllowRead {
		http.Error(w, "access denied", 403)
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

func (tc ThreadController) PostThread(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
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

	adminMember := &datastore.ThreadMember{
		ThreadId:          thread.Id,
		MailboxId:         mb.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: true,
	}

	if err := thread.AddMember(adminMember); err != nil {
		http.Error(w, "error adding thread member", 500)
		return
	}

	tc.GetThread(thread.Id, w, r, mb)
}

func (tc ThreadController) PutThread(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
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

	member, err := dbThread.GetMember(mb.Id)
	if err != nil || !member.AllowWrite {
		http.Error(w, "access denied: not thread member", 403)
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

	tc.GetThread(thread.Id, w, r, mb)
}

func (tc ThreadController) DeleteThread(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	thread := datastore.Thread{Id: rid(r)}

	member, err := thread.GetMember(mb.Id)
	if err != nil || !member.AllowWrite {
		http.Error(w, "access denied: not thread member", 403)
	}

	if err := thread.Delete(); err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "thread not found", 403)
		return
	}

	fmt.Fprintln(w, "thread deleted")
}

func (tc ThreadController) GetThreadMembers(tid string, w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
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

func (tc ThreadController) PostThreadMember(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	thread, err := datastore.GetThread(rid(r))
	if err != nil {
		http.Error(w, "thread not found", 404)
		return
	}

	tmember, err := thread.GetMember(mb.Id)
	if err != nil || !tmember.AllowWrite {
		http.Error(w, "access denied", 403)
		return
	}

	var member datastore.ThreadMember
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&member); err != nil {
		http.Error(w, "invalid JSON POST body", 400)
		return
	}

	comps := pathComponents(r)
	if len(comps) > 2 && comps[2] != "" {
		member.MailboxId = comps[2]
	}

	member.ThreadId = thread.Id
	if err := thread.AddMember(&member); err != nil {
		http.Error(w, "error adding member to thread", 500)
		fmt.Println(err)
		return
	}

	tc.GetThreadMembers(thread.Id, w, r, mb)
}

func (tc ThreadController) PutThreadMember(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	comps := pathComponents(r)
	if len(comps) < 3 {
		http.Error(w, "invalid mailbox id for thread member", 400)
		return
	}

	mailboxId := comps[2]
	thread, err := datastore.GetThread(rid(r))
	if err != nil {
		http.Error(w, "thread not found", 404)
		return
	}

	tmember, err := thread.GetMember(mb.Id)
	if err != nil || !tmember.AllowWrite {
		http.Error(w, "access denied", 403)
		return
	}

	var member datastore.ThreadMember
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&member); err != nil {
		http.Error(w, "invalid JSON PUT request body", 400)
		return
	}

	dbMember, err := thread.GetMember(mailboxId)
	if err != nil {
		http.Error(w, "thread member not found", 404)
		return
	}

	dbMember.AllowRead = member.AllowRead
	dbMember.AllowWrite = member.AllowWrite
	dbMember.AllowNotification = member.AllowNotification

	if err := dbMember.UpdatePermissions(); err != nil {
		http.Error(w, "error updating member permissions", 500)
		return
	}

	tc.GetThreadMembers(thread.Id, w, r, mb)
}

func (tc ThreadController) DeleteThreadMember(w http.ResponseWriter, r *http.Request, mb *datastore.Mailbox) {
	comps := pathComponents(r)
	if len(comps) < 3 {
		http.Error(w, "invalid mailbox id for thread member", 400)
		return
	}

	mailboxId := comps[2]
	thread, err := datastore.GetThread(rid(r))
	if err != nil {
		http.Error(w, "thread not found", 404)
		return
	}

	tmember, err := thread.GetMember(mb.Id)
	if err != nil || !tmember.AllowWrite {
		http.Error(w, "access denied", 403)
	}

	member, err := thread.GetMember(mailboxId)
	if err != nil {
		http.Error(w, "thread member not found", 404)
		return
	}

	if err := member.Remove(); err != nil {
		http.Error(w, "Error removing thread member", 500)
		return
	}

	fmt.Fprintln(w, "thread member removed")
}

func (tc ThreadController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
