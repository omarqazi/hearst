package controller

// this file implements helper functions
// for controllers in the server to use

import (
	"github.com/omarqazi/hearst/auth"
	"github.com/omarqazi/hearst/datastore"
	"net/http"
	"strings"
)

type Controller interface {
	GetUUID(uuid string) (output interface{}, err error)
}

// function path components returns the path components
// for the url in a given request
func pathComponents(r *http.Request) (comps []string) {
	comps = strings.Split(r.URL.Path, "/")
	return
}

// function rid returns the id part of the request
func rid(r *http.Request) string {
	comps := pathComponents(r)
	return comps[0]
}

func urlSubcategory(r *http.Request) string {
	comps := pathComponents(r)
	if len(comps) > 1 {
		return comps[1]
	}
	return ""
}

func authorizedMailbox(r *http.Request) (mb datastore.Mailbox, err error) {
	mailboxId := r.Header.Get("X-Hearst-Mailbox")
	if mailboxId == "" {
		mailboxId = r.URL.Query().Get("mailbox")
	}

	sessionToken := r.Header.Get("X-Hearst-Session")
	if sessionToken == "" {
		sessionToken = r.URL.Query().Get("session")
	}

	mb, err = datastore.GetMailbox(mailboxId)
	if err != nil {
		return
	}

	pubKey, er := auth.PublicKeyFromString(mb.PublicKey)
	if er != nil {
		return mb, er
	}

	session, erx := auth.ParseSession(sessionToken)
	if erx != nil {
		return mb, erx
	}

	err = session.Valid(pubKey, &serverSessionKey.PublicKey)
	return
}
