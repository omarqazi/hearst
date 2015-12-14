package main

import (
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
	log.Fatalln(errorMessage, http.ListenAndServe(bindAddress, nil))
}
