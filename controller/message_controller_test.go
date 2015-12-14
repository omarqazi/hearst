package controller

import (
	"encoding/json"
	"fmt"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"net/http/httptest"
	"testing"
)

var mc = http.StripPrefix("/message/", MessageController{})

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

	testRequestUrl := fmt.Sprintf("http://localhost:8080/message/%s", m.Id)
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

	var responseMessage datastore.Message
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&responseMessage); err != nil {
		t.Error("Error decoding message from response body", err)
		return
	}

	if responseMessage.ThreadId != m.ThreadId || responseMessage.SenderMailboxId != m.SenderMailboxId {
		t.Error("Expected message", m, "but got", responseMessage)
		return
	}

	m.Delete()
	thread.Delete()
	mailbox.Delete()
}
