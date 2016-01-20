package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"net/http/httptest"
	"testing"
)

var tc = http.StripPrefix("/thread/", ThreadController{})

func TestThreadGetRequest(t *testing.T) {
	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating mailbox with key:", err, mailbox, clientKey)
	}

	if err := mailbox.Insert(); err != nil {
		t.Fatal("Error inserting mailbox:", err)
	}

	thread := datastore.Thread{
		Subject: "whats up man",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread:", err)
		return
	}

	err = thread.AddMember(&datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		ThreadId:          thread.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: false,
	})
	if err != nil {
		t.Fatal("Error adding thread member:", err)
	}

	testRequestUrl := fmt.Sprintf("http://localhost:8080/thread/%s", thread.Id)
	req := testRequest("GET", testRequestUrl, nil, t, clientKey, &mailbox)

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response code but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	if err := thread.Delete(); err != nil {
		t.Error("Error deleting thread after test get request:", err)
	}
}

func TestThreadPostRequest(t *testing.T) {
	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating mailbox with key:", err)
	}

	if err = mailbox.Insert(); err != nil {
		t.Fatal("Error inserting mailbox:", err)
	}

	thread := datastore.Thread{
		Subject: "I posted this from the API",
	}
	threadBytes, err := json.Marshal(thread)
	if err != nil {
		t.Error("Error marshaling post body JSON for thread:", err)
		return
	}

	postBody := bytes.NewBuffer(threadBytes)
	req := testRequest("POST", "http://localhost:8080/thread/", postBody, t, clientKey, &mailbox)
	if err != nil {
		t.Error("Error building POST request:", err)
		return
	}

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	var responseThread datastore.Thread
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&responseThread); err != nil {
		t.Error("Error decoding response json:", err)
		return
	}

	if responseThread.Subject != thread.Subject {
		t.Error("Posted subject", thread.Subject, "but got", responseThread.Subject)
		return
	}

	trx, erx := datastore.GetThread(responseThread.Id)
	if erx != nil {
		t.Error("Error getting thread after POST:", erx)
		return
	}

	if trx.Subject != thread.Subject {
		t.Error("Posted subject", thread.Subject, "but database value was", trx.Subject)
		return
	}

	members, err := trx.GetAllMembers()
	if err != nil {
		t.Fatal("Error getting members for posted thread:", err)
	}

	if len(members) != 1 {
		t.Fatal("Expected 1 member for posted thread but found", len(members))
	}

	if members[0].MailboxId != mailbox.Id {
		t.Fatal("Expected posted thread to have admin member", mailbox.Id, "but found", members[0].MailboxId)
	}

	if err := trx.Delete(); err != nil {
		t.Error("Error deleting thread:", err)
		return
	}
}

func TestThreadPutRequest(t *testing.T) {
	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	if err := mailbox.Insert(); err != nil {
		t.Fatal("error inserting mailbox:", err)
	}

	thread := datastore.Thread{
		Subject: "this will be updated later",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for put request:", err)
		return
	}

	thread.AddMember(&datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		ThreadId:          thread.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: false,
	})

	updatedText := "this has now been updated"
	thread.Subject = updatedText

	threadBytes, err := json.Marshal(thread)
	if err != nil {
		t.Error("Error marshaling PUT body JSON for thread:", err)
		return
	}

	requestUrl := fmt.Sprintf("http://localhost:8080/thread/%s", thread.Id)
	putBody := bytes.NewBuffer(threadBytes)
	req := testRequest("PUT", requestUrl, putBody, t, clientKey, &mailbox)
	if err != nil {
		t.Error("Error building PUT request:", err)
		return
	}

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	var responseThread datastore.Thread
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&responseThread); err != nil {
		t.Error("Error decoding response json:", err)
		return
	}

	if responseThread.Subject != thread.Subject {
		t.Error("Posted subject", thread.Subject, "but got", responseThread.Subject)
		return
	}

	trx, erx := datastore.GetThread(responseThread.Id)
	if erx != nil {
		t.Error("Error getting thread after PUT:", erx)
		return
	}

	if trx.Subject != thread.Subject {
		t.Error("Posted subject", thread.Subject, "but database value was", trx.Subject)
		return
	}

	if err := trx.Delete(); err != nil {
		t.Error("Error deleting thread:", err)
		return
	}
}

func TestThreadDeleteRequest(t *testing.T) {
	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("error generating private key:", err)
	}

	if err := mailbox.Insert(); err != nil {
		t.Fatal("Error inserting mailbox:", err)
	}

	thread := datastore.Thread{
		Subject: "my days are numbered (in ms)",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error saving thread:", err)
		return
	}

	member := &datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		ThreadId:          thread.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: false,
	}
	if err := thread.AddMember(member); err != nil {
		t.Fatal("Error adding thread member:", err)
	}

	requestUrl := fmt.Sprintf("http://localhost:8080/thread/%s", thread.Id)
	req := testRequest("DELETE", requestUrl, nil, t, clientKey, &mailbox)

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	trx, erx := datastore.GetThread(thread.Id)
	if erx == nil {
		t.Error("Expected thread to be deleted but found", trx)
		return
	}
}

func TestThreadMembersGetRequest(t *testing.T) {
	thread := datastore.Thread{
		Subject: "a thread with members",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error saving thread:", err)
		return
	}

	mb, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	if err := mb.Insert(); err != nil {
		t.Error("Error saving mailbox:", err)
		return
	}

	tm := datastore.ThreadMember{
		ThreadId:          thread.Id,
		MailboxId:         mb.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: true,
	}

	if err := thread.AddMember(&tm); err != nil {
		t.Error("Error adding member to thread:", err)
		return
	}

	requestUrl := fmt.Sprintf("http://localhost:8080/thread/%s/members", thread.Id)
	req := testRequest("GET", requestUrl, nil, t, clientKey, &mb)

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	var threadMembers []datastore.ThreadMember
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&threadMembers); err != nil {
		t.Error("Error decoding response json:", err)
		return
	}

	if len(threadMembers) != 1 {
		t.Error("Expected 1 thread member but found", len(threadMembers))
		return
	}

	tmx := threadMembers[0]
	if tmx.AllowRead != tm.AllowRead || tmx.AllowWrite != tm.AllowWrite || tmx.AllowNotification != tm.AllowNotification {
		t.Error("Expected thread member", tm, "but response was", tmx)
		return
	}

	dbm, err := thread.GetMember(mb.Id)
	if err != nil {
		t.Error("Error getting thread member from database:", err)
		return
	}

	if dbm.AllowRead != tm.AllowRead || dbm.AllowWrite != tm.AllowWrite || dbm.AllowNotification != tm.AllowNotification {
		t.Error("Expected thread member", tm, "but response was", dbm)
		return
	}

	if err := dbm.Remove(); err != nil {
		t.Error("Error removing database member:", err)
	}

	if err := thread.Delete(); err != nil {
		t.Error("Error deleting thread:", err)
	}

	if err := mb.Delete(); err != nil {
		t.Error("Error deleting mailbox:", err)
	}
}

func TestThreadMembersPostRequest(t *testing.T) {
	thread := datastore.Thread{
		Subject: "member added with post request",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for member add post request:", err)
		return
	}

	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	if err := mailbox.Insert(); err != nil {
		t.Error("Error insert mailbox for member add post request", err)
		return
	}

	member := &datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: false,
	}

	if err := thread.AddMember(member); err != nil {
		t.Fatal("Error adding thread member:", err)
	}

	otherUser := datastore.NewMailbox()
	if err := otherUser.Insert(); err != nil {
		t.Fatal("Error inserting other user:", err)
	}

	member = &datastore.ThreadMember{
		MailboxId:         otherUser.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: true,
	}

	threadBytes, err := json.Marshal(member)
	if err != nil {
		t.Error("Error marshaling post body JSON for thread members:", err)
		return
	}

	postBody := bytes.NewBuffer(threadBytes)
	rurl := fmt.Sprintf("http://localhost:8080/thread/%s/members", thread.Id)
	req := testRequest("POST", rurl, postBody, t, clientKey, &mailbox)

	allMembers, err := thread.GetAllMembers()
	if err != nil {
		t.Error("Error getting all members of thread for member POST:", err)
		return
	}

	if len(allMembers) != 1 {
		t.Error("Expected 1 thread members but found", len(allMembers))
		return
	}

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	allMembers, err = thread.GetAllMembers()
	if err != nil {
		t.Error("Error getting all members of thread for member after POST:", err)
		return
	}

	if len(allMembers) != 2 {
		t.Error("Expected to find 2 members in database but found", len(allMembers))
		return
	}

	am := allMembers[1]
	if am.AllowWrite {
		t.Error("Expected AllowWrite to be false but found true")
		return
	}

	am.Remove()
	thread.Delete()
	mailbox.Delete()
}

func TestThreadMembersPutRequest(t *testing.T) {
	thread := datastore.Thread{
		Subject: "member that will be updated",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for member add put request:", err)
		return
	}

	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for member put request", err)
		return
	}

	member := datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: false,
	}

	if err := thread.AddMember(&member); err != nil {
		t.Error("Error adding member to thread:", err)
		return
	}

	member.AllowNotification = true

	threadBytes, err := json.Marshal(member)
	if err != nil {
		t.Error("Error marshaling json for thread member")
		return
	}
	putBody := bytes.NewBuffer(threadBytes)

	rurl := fmt.Sprintf("http://localhost:8080/thread/%s/members/%s", thread.Id, mailbox.Id)
	req := testRequest("PUT", rurl, putBody, t, clientKey, &mailbox)

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	dbm, err := thread.GetMember(mailbox.Id)
	if err != nil {
		t.Error("Error getting member from database:", err)
		return
	}

	if dbm.AllowNotification != true {
		t.Error("AllowWrite is false when it was updated to true")
		return
	}
	dbm.Remove()
	thread.Delete()
	mailbox.Delete()
}

func TestThreadMembersDeleteRequest(t *testing.T) {
	thread := datastore.Thread{
		Subject: "delete one of my members",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread:", err)
		return
	}

	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox:", err)
		return
	}

	err = thread.AddMember(&datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: true,
	})
	if err != nil {
		t.Error("Error adding member to thread:", err)
		return
	}

	rurl := fmt.Sprintf("http://localhost:8080/thread/%s/members/%s", thread.Id, mailbox.Id)
	req := testRequest("DELETE", rurl, nil, t, clientKey, &mailbox)

	w := httptest.NewRecorder()
	tc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	dbm, err := thread.GetMember(mailbox.Id)
	if err == nil {
		t.Error("Deleted thread member but still found", dbm)
		dbm.Remove()
	}

	thread.Delete()
	mailbox.Delete()
}
