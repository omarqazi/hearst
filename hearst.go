package main

import (
	"github.com/omarqazi/hearst/datastore"
	"log"
	"net/http"
)

const startMessage = "Starting Hearst on port 8080..."
const errorMessage = "Error starting server:"
const bindAddress = ":8080"

func init() {
	log.Println(startMessage)
}

func main() {
	es := datastore.NewStream(datastore.RedisDb)
	ec := es.EventChannel("")

	go func() {
		for evt := range ec {
			log.Println(evt.ModelClass, evt.Action, evt.ObjectId)
		}
	}()
	log.Fatalln(errorMessage, http.ListenAndServe(bindAddress, nil))
}
