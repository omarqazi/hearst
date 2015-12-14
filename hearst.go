package main

import (
	"github.com/omarqazi/hearst/auth"
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
	privateKey, err := auth.GeneratePrivateKey(1024)
	if err != nil {
		log.Println("Error generating private key:", err)
		return
	}

	publicKeyString, err := auth.StringForPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Println("Error generating string for public key:", err)
		return
	}

	mb := &datastore.Mailbox{
		DeviceId:  "push-notification-id",
		PublicKey: publicKeyString,
	}

	if token, err := auth.NewToken(privateKey); err != nil {
		log.Println("Error genearing auth token:", err)
		return
	} else {
		log.Println("Generated token:", token)
	}

	if err := mb.Insert(); err != nil {
		log.Println("Error inserting mailbox", err)
	}

	log.Println(mb.Id)

	log.Fatalln(errorMessage, http.ListenAndServe(bindAddress, nil))
}
