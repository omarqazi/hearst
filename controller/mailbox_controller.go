package controller

import (
	"encoding/json"
	"fmt"
	"github.com/omarqazi/hearst/datastore"
	"log"
	"net/http"
)

type MailboxController struct {
}

func (c MailboxController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		c.GetMailbox(rid(r), w, r)
	case "POST":
		c.PostMailbox(w, r)
	case "PUT":
		c.PutMailbox(w, r)
	case "DELETE":
		c.DeleteMailbox(w, r)
	default:
		c.HandleUnknown(w, r)
	}
}

// Function GetMailbox handles a GET request by retrieving
// a mailbox and rendering it as JSON
func (c MailboxController) GetMailbox(mbid string, w http.ResponseWriter, r *http.Request) {
	mb, err := datastore.GetMailbox(mbid)
	if err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "mailbox not found")
		return
	}

	encoder := json.NewEncoder(w)
	w.Header().Add("Content-Type", "application/json")
	if err := encoder.Encode(mb); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "error marshaling json response :(")
		log.Println("Error encoding mailbox json:", err)
		return
	}
}

// Function PostMailbox handles a HTTP POST request
// By parsing the JSON request body and inserting it
// into the database
func (c MailboxController) PostMailbox(w http.ResponseWriter, r *http.Request) {
	var mailbox datastore.Mailbox
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mailbox); err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, "invalid request JSON")
		return
	}

	if err := mailbox.Insert(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "error saving mailbox:", err.Error())
		return
	}

	c.GetMailbox(mailbox.Id, w, r)
}

// Function PutMailbox handles an HTTP PUT request
// by parsing the JSON request body and updating
// the existing database record
func (c MailboxController) PutMailbox(w http.ResponseWriter, r *http.Request) {
	var mailbox datastore.Mailbox
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&mailbox); err != nil {
		w.WriteHeader(400)
		fmt.Fprintln(w, "invalid request JSON")
		return
	}

	if mailbox.Id == "" {
		mailbox.Id = rid(r)
	}

	dbBox, erx := datastore.GetMailbox(mailbox.Id)
	if erx != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "mailbox not found")
		return
	}

	if mailbox.PublicKey == "" {
		mailbox.PublicKey = dbBox.PublicKey
	}

	if mailbox.DeviceId == "" {
		mailbox.DeviceId = dbBox.DeviceId
	}

	if err := mailbox.Update(); err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "error updating mailbox")
		return
	}

	c.GetMailbox(mailbox.Id, w, r)
}

func (c MailboxController) DeleteMailbox(w http.ResponseWriter, r *http.Request) {
	identifier := rid(r)
	mailbox := datastore.Mailbox{Id: identifier}
	if err := mailbox.Delete(); err != nil {
		w.WriteHeader(404)
		fmt.Fprintln(w, "mailbox not found")
		return
	}
	fmt.Fprintln(w, "Mailbox deleted")
}

func (c MailboxController) HandleUnknown(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	fmt.Fprintln(w, "what the fuck are you talking about?")
}
