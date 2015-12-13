package main

import (
	_ "github.com/omarqazi/hearst/datastore"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting hearst on :8080...")
	err := http.ListenAndServe(":8080", nil)
	log.Fatalln("Error starting server:", err)
}
