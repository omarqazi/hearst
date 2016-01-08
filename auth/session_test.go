package auth

import (
	"testing"
	"time"
)

func TestSession(t *testing.T) {
	newKey, err := GeneratePrivateKey(512)
	if err != nil {
		t.Fatal(err)
	}

	token, err := NewToken(newKey)
	if err != nil {
		t.Fatal(err)
	}

	clientKey, err := GeneratePrivateKey(512)
	if err != nil {
		t.Fatal(err)
	}

	s := &Session{
		Token:    token,
		Duration: 1000 * time.Millisecond,
	}

	sig, err := s.SignatureFor(clientKey)
	if err != nil {
		t.Fatal(err)
	}

	s.Signature = sig
	err = s.Valid(&clientKey.PublicKey, &newKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(s.Duration)
	err = s.Valid(&clientKey.PublicKey, &newKey.PublicKey)
	if err == nil {
		t.Fatal("session was supposed to expire but did not")
	}
}
