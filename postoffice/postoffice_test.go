package postoffice

import (
	"testing"
)

func TestSocket(t *testing.T) {
	client, err := NewClient("chat.smick.co")
	if err != nil {
		t.Fatal("client failed to init", err)
	}

	if err := client.ConnectSocket(); err != nil {
		t.Fatal("Error connecting socket:", err)
	}
}
