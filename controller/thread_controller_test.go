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
