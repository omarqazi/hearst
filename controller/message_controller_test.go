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

var mc = http.StripPrefix("/messages/", MessageController{})

func TestMessageGetRequest(t *testing.T) {
	thread := datastore.Thread{Subject: "test message get request"}
	if err := thread.Insert(); err != nil {
		t.Error("Error inserting thread for test message GET request:", err)
		return
	}

	mailbox := datastore.Mailbox{
		PublicKey: "some-public-key",
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for test message GET request:", err)
		return
	}

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

	mailbox := datastore.Mailbox{
		PublicKey: "some-public-key",
		DeviceId:  "some-device-id",
	}
	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox for test message GET request:", err)
		return
	}

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

	w := httptest.NewRecorder()
	mc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response code but got", w.Code)
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
