package main

import (
	"log"
	"net/http"
)

const startMessage = "Starting Hearst on port "
const errorMessage = "Error starting server:"
const bindAddress = ":8080"

func init() {
	log.Println(startMessage, bindAddress)
}

func main() {
	log.Fatalln(errorMessage, http.ListenAndServe(bindAddress, nil))
}
