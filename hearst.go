package main

import (
	"log"
	"net/http"
)

// ye

func main() {
	log.Println("Starting hearst on :8080...")
	err := http.ListenAndServe(":8080", nil)
	log.Fatalln("Error starting server:", err)
}
