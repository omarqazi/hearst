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

var mbc = http.StripPrefix("/mailbox/", MailboxController{})

func TestMailboxGetRequest(t *testing.T) {
	mb := datastore.Mailbox{
		PublicKey: "BeginRsaKey",
		DeviceId:  "some-device-id",
	}

	if err := mb.Insert(); err != nil {
		t.Error("Error adding mailbox to test mailbox controller GET:", err)
		return
	}

	testRequestUrl := fmt.Sprintf("http://localhost:8080/mailbox/%s", mb.Id)
	req, err := http.NewRequest("GET", testRequestUrl, nil)
	if err != nil {
		t.Error("Failed to create test request:", err)
		return
	}

	w := httptest.NewRecorder()
	mbc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Error("Expected content type 'application/json' but got", ct)
		return
	}

	decoder := json.NewDecoder(w.Body)
	var responseBox datastore.Mailbox
	if err := decoder.Decode(&responseBox); err != nil {
		t.Error("Error decoding json response", err)
		return
	}

	if responseBox.Id != mb.Id {
		t.Error("Expected response to have ID", mb.Id, "but got", responseBox.Id)
		return
	}

	if err := mb.Delete(); err != nil {
		t.Error("Error deleting mailbox:", err)
		return
	}
}

func TestMailboxPostRequest(t *testing.T) {
	mailbox := datastore.Mailbox{
		DeviceId:  "iphone-id",
		PublicKey: "RSAKey",
	}

	mailboxBytes, err := json.Marshal(mailbox)
	if err != nil {
		t.Error("Error marshaling post body JSON for mailbox:", err)
		return
	}

	requestUrl := "http://localhost:8080/mailbox/"
	postBody := bytes.NewBuffer(mailboxBytes)
	req, err := http.NewRequest("POST", requestUrl, postBody)
	if err != nil {
		t.Error("Error building POST request:", err)
		return
	}

	w := httptest.NewRecorder()
	mbc.ServeHTTP(w, req)
	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	var responseMailbox datastore.Mailbox
	decoder := json.NewDecoder(w.Body)
	if err := decoder.Decode(&responseMailbox); err != nil {
		t.Error("Error decoding response body", err)
		return
	}

	if responseMailbox.PublicKey != mailbox.PublicKey || responseMailbox.DeviceId != mailbox.DeviceId {
		t.Error("Expected mailbox", mailbox, "but got", responseMailbox)
		return
	}

	mbx, erx := datastore.GetMailbox(responseMailbox.Id)
	if erx != nil {
		t.Error("Error getting mailbox from database:", erx)
		return
	}

	if mbx.PublicKey != mailbox.PublicKey || mbx.DeviceId != mailbox.DeviceId {
		t.Error("Expected mailbox", mailbox, "but got", responseMailbox)
		return
	}

	mbx.Delete()
}

func TestMailboxPutRequest(t *testing.T) {
	clientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		t.Fatal("Error generating private key", err)
	}

	pubKey, err := auth.StringForPublicKey(&clientKey.PublicKey)
	if err != nil {
		t.Fatal("Error generating string for public key", err)
	}

	mailbox := datastore.Mailbox{
		DeviceId:  "something",
		PublicKey: pubKey,
	}

	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox", err)
		return
	}
	defer mailbox.Delete()

	newDeviceId := "else"
	mailbox.DeviceId = newDeviceId
	mailboxBytes, err := json.Marshal(mailbox)
	if err != nil {
		t.Error("Error marshaling put body JSON for mailbox:", err)
		return
	}

	requestUrl := "http://localhost:8080/mailbox/"
	postBody := bytes.NewBuffer(mailboxBytes)
	req, err := http.NewRequest("PUT", requestUrl, postBody)
	if err != nil {
		t.Error("Error building PUT request:", err)
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
	mbc.ServeHTTP(w, req)
	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	mbx, erx := datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		t.Error("Error getting mailbox from database:", erx)
		return
	}

	if mbx.DeviceId != newDeviceId {
		t.Error("Expected PUT request to update public key to", newDeviceId, "but found", mbx.DeviceId)
		return
	}

	anotherClientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		t.Fatal("Error generating private key", err)
	}

	pubKey, err = auth.StringForPublicKey(&anotherClientKey.PublicKey)
	if err != nil {
		t.Fatal("Error generating string for public key", err)
	}

	anotherMailbox := datastore.Mailbox{
		DeviceId:  "something",
		PublicKey: pubKey,
	}

	if err := anotherMailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox", err)
		return
	}
	defer anotherMailbox.Delete()

	postBody = bytes.NewBuffer(mailboxBytes)
	req, err = http.NewRequest("PUT", requestUrl, postBody)
	if err != nil {
		t.Error("Error building PUT request:", err)
		return
	}

	req.Header.Add("X-Hearst-Mailbox", anotherMailbox.Id)

	token, err = auth.NewToken(serverSessionKey)
	if err != nil {
		t.Fatal("Error generating token", err)
	}

	session = auth.Session{
		Token:    token,
		Duration: 300 * time.Second,
	}
	sig, err = session.SignatureFor(anotherClientKey)
	if err != nil {
		t.Fatal("Error signing session:", err)
	}

	session.Signature = sig
	req.Header.Add("X-Hearst-Session", session.String())

	w = httptest.NewRecorder()
	mbc.ServeHTTP(w, req)

	if w.Code != 403 {
		t.Error("Expected 403 response when authorized as another user but got", w.Code)
		return
	}
}

func TestMailboxDeleteRequest(t *testing.T) {
	clientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		t.Fatal("Error generating private key", err)
	}

	pubKey, err := auth.StringForPublicKey(&clientKey.PublicKey)
	if err != nil {
		t.Fatal("Error generating string for public key", err)
	}

	mailbox := datastore.Mailbox{
		DeviceId:  "short",
		PublicKey: pubKey,
	}

	if err := mailbox.Insert(); err != nil {
		t.Error("Error inserting mailbox")
		return
	}

	requestUrl := fmt.Sprintf("http://localhost:8080/mailbox/%s", mailbox.Id)
	req, err := http.NewRequest("DELETE", requestUrl, nil)
	if err != nil {
		t.Error("Error building delete request", err)
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
	mbc.ServeHTTP(w, req)

	if w.Code > 299 || w.Code < 200 {
		t.Error("Expected 200 response but got", w.Code)
		return
	}

	mrx, erx := datastore.GetMailbox(mailbox.Id)
	if erx == nil {
		t.Error("Able to retrieve mailbox after DELETE request but got", mrx)
	}
}
