package controller

import (
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"net/http/httptest"
	"testing"
)

var wsc = http.StripPrefix("/socket/", WebSocketController{})

func TestWebSocket(t *testing.T) {
	mailbox, clientKey, err := datastore.NewMailboxWithKey()
	if err != nil {
		t.Fatal("Error generating private key:", err)
	}

	if err := mailbox.Insert(); err != nil {
		t.Fatal("Error inserting mailbox:", err)
	}

	req := testRequest("GET", "http://localhost:8080/socket/", nil, t, clientKey, &mailbox)

	w := httptest.NewRecorder()
	wsc.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Error("Expected 400 response code but got", w.Code)
		return
	}
}
