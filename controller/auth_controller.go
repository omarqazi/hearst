package controller

import (
	"crypto/rsa"
	"fmt"
	"github.com/omarqazi/hearst/auth"
	"github.com/omarqazi/hearst/datastore"
	"log"
	"net/http"
)

var serverSessionKey *rsa.PrivateKey

const keySize = 2048
const sessionKeyCacheLocation = "hearst-env-server-session-key"

func init() {
	privateKeyString, err := datastore.RedisDb.Get(sessionKeyCacheLocation).Result()
	if err != nil {
		serverSessionKey, err = auth.GeneratePrivateKey(keySize)
		if err != nil {
			log.Fatalln("Could not generate private key", err)
		}

		privateKeyString = auth.StringForPrivateKey(serverSessionKey)
		if err = datastore.RedisDb.Set(sessionKeyCacheLocation, privateKeyString, 0).Err(); err != nil {
			log.Println("Could not save server session key to redis:", err)
		}
	}

	serverSessionKey, err = auth.PrivateKeyFromString(privateKeyString)
	if err != nil {
		log.Fatalln("error parsing private key from redis:", err)
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
