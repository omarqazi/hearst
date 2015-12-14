package datastore

import (
	"github.com/jmoiron/sqlx/types"
	"testing"
)

func TestInsertMessage(t *testing.T) {
	TestMailboxInsert(t)
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Error("Error getting mailbox for message insert:", err)
		return
	}

	TestInsertThread(t)
	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for message insert:", err)
		return
	}

	m := Message{
		ThreadId:        tr.Id,
		SenderMailboxId: mb.Id,
		Topic:           "chat-message",
		Body:            "hey man",
		Labels:          types.JsonText("{}"),
		Payload:         types.JsonText("{}"),
	}

	m.Labels.Scan("{}")
	m.Payload.Scan("{}")

	if err := m.Insert(); err != nil {
		t.Error("Error inserting message:", err)
		return
	}

	CleanUpThread(t)
	CleanUpMailbox(t)
}
