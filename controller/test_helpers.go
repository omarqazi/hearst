package controller

import (
	"crypto/rsa"
	"github.com/omarqazi/hearst/auth"
	"github.com/omarqazi/hearst/datastore"
	"io"
	"net/http"
	"testing"
	"time"
)

func testRequest(method, url string, body io.Reader, t *testing.T, clientKey *rsa.PrivateKey, mb *datastore.Mailbox) *http.Request {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatal("Error creating HTTP request", err)
	}

	req.Header.Add("X-Hearst-Mailbox", mb.Id)
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

	return req
}
