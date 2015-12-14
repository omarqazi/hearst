package controller

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

var tc = http.StripPrefix("/thread/", ThreadController{})

func TestThreadGetRequest(t *testing.T) {
	testRequestUrl := "http://localhost:8080/thread/some-thread"
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
}
