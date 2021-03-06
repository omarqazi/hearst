package auth

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Session struct {
	Token     string        // A token from the server
	Duration  time.Duration // the duration of the session
	Signature []byte        // Client signature of session message
}

func ParseSession(sessionKey string) (Session, error) {
	s := Session{}
	comps := strings.Split(sessionKey, "!@@!")
	if len(comps) < 3 {
		return s, errors.New("error parsing session key: len") // not a valid sessionKey
	}

	s.Token = comps[0]
	durationInt, err := strconv.ParseInt(comps[1], 10, 64)
	if err != nil {
		return s, errors.New("error parsing session key: int") // invalid duration
	}
	s.Duration = time.Duration(durationInt) * time.Second

	s.Signature, err = base64.URLEncoding.DecodeString(comps[2])
	if err != nil {
		return s, errors.New("Error parsing session key: bsf")
	}

	return s, nil
}

func (s Session) Message() string {
	secs := s.Duration / time.Second
	return fmt.Sprintf("%s!@@!%d", s.Token, secs)
}

func (s Session) SignatureFor(priv *rsa.PrivateKey) ([]byte, error) {
	sig, err := SignMessageWithKey(priv, s.Message())
	return sig, err
}

func (s Session) SignatureString() string {
	sig := base64.URLEncoding.EncodeToString(s.Signature)
	return sig
}

func (s Session) String() string {
	return fmt.Sprintf("%s!@@!%s", s.Message(), s.SignatureString())
}

func (s Session) Valid(client *rsa.PublicKey, server *rsa.PublicKey) error {
	if err := ValidateSignatureForMessage(s.Message(), s.Signature, client); err != nil {
		return fmt.Errorf("client signature invalid") // the client did not sign off on this
	}

	maxDuration := s.Duration
	if maxDuration > (24 * time.Hour) {
		maxDuration = 24 * time.Hour
	}

	if TokenValid(s.Token, maxDuration, server) == false {
		return fmt.Errorf("token has expired or is invalid")
	}

	return nil
}
