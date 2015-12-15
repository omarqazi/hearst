package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var wsc = http.StripPrefix("/socket/", WebSocketController{})

func TestWebSocket(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/socket/", nil)
	if err != nil {
		t.Error("Error building GET request:", err)
		return
	}

	w := httptest.NewRecorder()
	wsc.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Error("Expected 400 response code but got", w.Code)
		return
	}
}
