package controller

import (
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

	/* if err := thread.Delete(); err != nil {
		t.Error("Error deleting thread after test get request:", err)
	} */
}
