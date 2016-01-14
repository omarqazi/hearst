package controller

import (
	"crypto/rsa"
	"fmt"
	"github.com/omarqazi/hearst/auth"
	"log"
	"net/http"
)

var serverSessionKey *rsa.PrivateKey

const keySize = 2048

func init() {
	var err error
	serverSessionKey, err = auth.GeneratePrivateKey(keySize)
	if err != nil {
		log.Fatalln("Could not generate private key", err)
	}
}

type AuthController struct {
}

func (ac AuthController) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token, err := auth.NewToken(serverSessionKey)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintln(w, token)
}
