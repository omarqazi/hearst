package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/omarqazi/hearst/auth"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var mc = http.StripPrefix("/messages/", MessageController{})

func TestMessageGetRequest(t *testing.T) {
	thread := datastore.Thread{Subject: "test message get request"}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for test message GET request:", err)
		return
	}

	clientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		t.Fatal("Could not generate private key:", err)
	}

	pubKey, err := auth.StringForPublicKey(&clientKey.PublicKey)
	if err != nil {
		t.Fatal("Error generating public key string", err)
	}

	mailbox := datastore.Mailbox{
		PublicKey: pubKey,
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for test message GET request:", err)
		return
	}

	thread.AddMember(&datastore.ThreadMember{
		ThreadId:          thread.Id,
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: false,
	})

	m := &datastore.Message{
		ThreadId:        thread.Id,
		SenderMailboxId: mailbox.Id,
	}
	m.Payload.Scan("{}")
	m.Labels.Scan("{}")
	if err := m.Insert(); err != nil {
		t.Error("Error inserting message:", err)
		return
	}

	testRequestUrl := fmt.Sprintf("http://localhost:8080/messages/%s", thread.Id)
	req, err := http.NewRequest("GET", testRequestUrl, nil)
	if err != nil {
		t.Error("Error building GET request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", mailbox.Id)
	token, err := auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session := auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err := session.SignatureFor(clientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}
	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w := httptest.NewRecorder()
	mc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response code but got", w.Code)
		return
	}

	var responseMessages []datastore.Message
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&responseMessages); err != nil {
		t.Error("Error decoding message from response body", err)
		return
	}

	if len(responseMessages) == 0 {
		t.Error("Expected 1 message but found 0")
		return
	}

	responseMessage := responseMessages[0]

	if responseMessage.SenderMailboxId != m.SenderMailboxId {
		t.Error("Expected message", m, "but got", responseMessage)
		return
	}

	m.Delete()
	thread.Delete()
	mailbox.Delete()
}

func TestMessagePostRequest(t *testing.T) {
	thread := datastore.Thread{Subject: "test message post request"}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for test message GET request:", err)
		return
	}

	clientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	pubKey, err := auth.StringForPublicKey(&clientKey.PublicKey)
	if err != nil {
		t.Fatal("Error generating string for public key:", err)
	}

	mailbox := datastore.Mailbox{
		PublicKey: pubKey,
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for test message GET request:", err)
		return
	}

	thread.AddMember(&datastore.ThreadMember{
		ThreadId:          thread.Id,
		MailboxId:         mailbox.Id,
		AllowRead:         true,
		AllowWrite:        true,
		AllowNotification: false,
	})

	messageBody := "yo yo yo"
	m := &datastore.Message{
		ThreadId:        thread.Id,
		SenderMailboxId: mailbox.Id,
		Body:            messageBody,
	}
	m.Payload.Scan("{}")
	m.Labels.Scan("{}")
	messageBytes, err := json.Marshal(m)
	if err != nil {
		t.Error("Error marshaling message for message post request:", err)
		return
	}

	postBody := bytes.NewBuffer(messageBytes)
	testRequestUrl := fmt.Sprintf("http://localhost:8080/messages/%s", thread.Id)
	req, err := http.NewRequest("POST", testRequestUrl, postBody)
	if err != nil {
		t.Error("Error building POST request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", mailbox.Id)
	token, err := auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session := auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err := session.SignatureFor(clientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}
	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w := httptest.NewRecorder()
	mc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response code but got", w.Code, w.Body.String())
		return
	}

	messages, err := thread.RecentMessages(10)
	if err != nil {
		t.Error("Error getting recent messages", err)
		return
	}

	if len(messages) != 1 {
		t.Error("Error: expected 1 message but got", len(messages))
		return
	}

	message := messages[0]

	if message.Body != messageBody {
		t.Error("Expected body", messageBody, "but got", message.Body)
	}

	message.Delete()
	thread.Delete()
	mailbox.Delete()
}

func TestPermissions(t *testing.T) {
	clientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	pubKey, err := auth.StringForPublicKey(&clientKey.PublicKey)
	if err != nil {
		t.Fatal("Error generating string for public key:", err)
	}

	mailbox := datastore.Mailbox{
		PublicKey: pubKey,
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for test message GET request:", err)
		return
	}

	thread := datastore.Thread{Subject: "test message permissions"}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for test message permissions request:", err)
		return
	}

	tm := &datastore.ThreadMember{
		ThreadId:          thread.Id,
		MailboxId:         mailbox.Id,
		AllowRead:         false,
		AllowWrite:        false,
		AllowNotification: false,
	}

	thread.AddMember(tm)

	m := &datastore.Message{
		ThreadId:        thread.Id,
		SenderMailboxId: mailbox.Id,
	}
	m.Payload.Scan("{}")
	m.Labels.Scan("{}")
	if err := m.Insert(); err != nil {
		t.Error("Error inserting message:", err)
		return
	}

	testRequestUrl := fmt.Sprintf("http://localhost:8080/messages/%s", thread.Id)
	req, err := http.NewRequest("GET", testRequestUrl, nil)
	if err != nil {
		t.Error("Error building GET request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", mailbox.Id)
	token, err := auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session := auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err := session.SignatureFor(clientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}
	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w := httptest.NewRecorder()
	mc.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatal("Expected server to deny access to reading message but got", w.Code)
	}

	tm.AllowRead = true
	if err := tm.UpdatePermissions(); err != nil {
		t.Fatal("Error updating permissions:", err)
	}

	req, err = http.NewRequest("GET", testRequestUrl, nil)
	if err != nil {
		t.Error("Error building GET request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", mailbox.Id)
	token, err = auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session = auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err = session.SignatureFor(clientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}
	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w = httptest.NewRecorder()
	mc.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatal("Expected to be able to read message after updating permissions but got", w.Code)
	}

	messageBody := "yo yo yo"
	m = &datastore.Message{
		ThreadId:        thread.Id,
		SenderMailboxId: mailbox.Id,
		Body:            messageBody,
	}
	m.Payload.Scan("{}")
	m.Labels.Scan("{}")
	messageBytes, err := json.Marshal(m)
	if err != nil {
		t.Error("Error marshaling message for message post request:", err)
		return
	}

	postBody := bytes.NewBuffer(messageBytes)
	req, err = http.NewRequest("POST", testRequestUrl, postBody)
	if err != nil {
		t.Error("Error building POST request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", mailbox.Id)
	token, err = auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session = auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err = session.SignatureFor(clientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}
	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w = httptest.NewRecorder()
	mc.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Fatal("Expected controller to deny access to message write but got", w.Code)
	}

	tm.AllowWrite = true
	if err := tm.UpdatePermissions(); err != nil {
		t.Fatal("Error updating permissions:", err)
	}

	postBody = bytes.NewBuffer(messageBytes)
	req, err = http.NewRequest("POST", testRequestUrl, postBody)
	if err != nil {
		t.Error("Error building POST request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", mailbox.Id)
	token, err = auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session = auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err = session.SignatureFor(clientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}
	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w = httptest.NewRecorder()
	mc.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatal("Expected controller to allow access to message write but got", w.Code)
	}

}
