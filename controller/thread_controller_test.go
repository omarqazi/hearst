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
	thread := datastore.Thread{
		Subject: "whats up man",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread:", err)
		return
	}

	testRequestUrl := fmt.Sprintf("http://localhost:8080/thread/%s", thread.Id)
	req, err := http.NewRequest("GET", testRequestUrl, nil)
	if err != nil {
		t.Error("Error building GET request:", err)
		return
	}

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
	thread := datastore.Thread{
		Subject: "I posted this from the API",
	}

	threadBytes, err := json.Marshal(thread)
	if err != nil {
		t.Error("Error marshaling post body JSON for thread:", err)
		return
	}

	postBody := bytes.NewBuffer(threadBytes)
	req, err := http.NewRequest("POST", "http://localhost:8080/thread/", postBody)
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

	if err := trx.Delete(); err != nil {
		t.Error("Error deleting thread:", err)
		return
	}
}

func TestThreadPutRequest(t *testing.T) {
	thread := datastore.Thread{
		Subject: "this will be updated later",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for put request:", err)
		return
	}

	updatedText := "this has now been updated"
	thread.Subject = updatedText

	threadBytes, err := json.Marshal(thread)
	if err != nil {
		t.Error("Error marshaling PUT body JSON for thread:", err)
		return
	}

	requestUrl := fmt.Sprintf("http://localhost:8080/thread/%s", thread.Id)
	putBody := bytes.NewBuffer(threadBytes)
	req, err := http.NewRequest("PUT", requestUrl, putBody)
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
	thread := datastore.Thread{
		Subject: "my days are numbered (in ms)",
	}
	if err := thread.Insert(); err != nil {
		t.Error("Error saving thread:", err)
		return
	}

	requestUrl := fmt.Sprintf("http://localhost:8080/thread/%s", thread.Id)
	req, err := http.NewRequest("DELETE", requestUrl, nil)
	if err != nil {
		t.Error("Error building DELETE request:", err)
		return
	}

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

	mb := datastore.Mailbox{
		PublicKey: "some-public-key",
		DeviceId:  "some-device-id",
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
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		t.Error("Error building GET request:", err)
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

	mailbox := datastore.Mailbox{
		PublicKey: "some-public-key",
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error insert mailbox for member add post request", err)
		return
	}

	member := datastore.ThreadMember{
		MailboxId:         mailbox.Id,
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
	req, err := http.NewRequest("POST", rurl, postBody)
	if err != nil {
		t.Error("Error building POST request:", err)
		return
	}

	allMembers, err := thread.GetAllMembers()
	if err != nil {
		t.Error("Error getting all members of thread for member POST:", err)
		return
	}

	if len(allMembers) > 0 {
		t.Error("Expected 0 thread members but found", len(allMembers))
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

	if len(allMembers) < 1 {
		t.Error("Expected to find some members in database but found nothing")
		return
	}

	am := allMembers[0]
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

	mailbox := datastore.Mailbox{
		PublicKey: "some-public-key",
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for member put request", err)
		return
	}

	member := datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: true,
	}

	if err := thread.AddMember(&member); err != nil {
		t.Error("Error adding member to thread:", err)
		return
	}

	member.AllowWrite = true

	threadBytes, err := json.Marshal(member)
	if err != nil {
		t.Error("Error marshaling json for thread member")
		return
	}
	putBody := bytes.NewBuffer(threadBytes)

	rurl := fmt.Sprintf("http://localhost:8080/thread/%s/members/%s", thread.Id, mailbox.Id)
	req, err := http.NewRequest("PUT", rurl, putBody)
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

	dbm, err := thread.GetMember(mailbox.Id)
	if err != nil {
		t.Error("Error getting member from database:", err)
		return
	}

	if dbm.AllowWrite != true {
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

	mailbox := datastore.Mailbox{
		PublicKey: "some-public-key",
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox:", err)
		return
	}

	err := thread.AddMember(&datastore.ThreadMember{
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: true,
	})
	if err != nil {
		t.Error("Error adding member to thread:", err)
		return
	}

	rurl := fmt.Sprintf("http://localhost:8080/thread/%s/members/%s", thread.Id, mailbox.Id)
	req, err := http.NewRequest("DELETE", rurl, nil)
	if err != nil {
		t.Error("Error building DELETE request:", err)
		return
	}

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
