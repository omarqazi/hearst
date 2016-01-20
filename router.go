package main

import (
	"github.com/omarqazi/hearst/controller"
	"net/http"
)

const staticPath = "www"

var routes = map[string]http.Handler{
	"/":          http.FileServer(http.Dir(staticPath)),
	"/mailbox/":  controller.MailboxController{},
	"/thread/":   controller.ThreadController{},
	"/messages/": controller.MessageController{},
	"/socket/":   controller.WebSocketController{},
	"/auth/":     controller.AuthController{},
}

func init() {
	for rule, handler := range routes {
		http.Handle(rule, http.StripPrefix(rule, handler))
	}
}
