package datastore

import (
	"github.com/jmoiron/sqlx/types"
	"testing"
)

const testMessageTopic = "chat-message"
const testMessageBody = "hey man"

var testMessageId = ""

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
		Topic:           testMessageTopic,
		Body:            testMessageBody,
		Labels:          types.JSONText("{}"),
		Payload:         types.JSONText("{}"),
	}

	m.Labels.Scan("{}")
	m.Payload.Scan("{}")

	if err := m.Insert(); err != nil {
		t.Error("Error inserting message:", err)
		return
	}

	CleanUpMessages(t)
	testMessageId = m.Id
}

func TestSelectMessage(t *testing.T) {
	TestInsertMessage(t)
	m, err := GetMessage(testMessageId)
	if err != nil {
		t.Error("Error getting message:", err)
		return
	}

	if m.Topic != testMessageTopic || m.Body != testMessageBody {
		t.Error("Expected different message contents", m)
		return
	}

	CleanUpMessages(t)
}

func TestUpdateMessage(t *testing.T) {
	TestInsertMessage(t)
	defer CleanUpMessages(t)

	m, err := GetMessage(testMessageId)
	if err != nil {
		t.Error("Error getting message for update:", err)
		return
	}

	newMessageBody := "hey dude"
	m.Body = newMessageBody
	if err := m.Update(); err != nil {
		t.Error("Error updating message:", err)
		return
	}

	mx, erx := GetMessage(testMessageId)
	if erx != nil {
		t.Error("Error getting message after update:", err)
		return
	}

	if mx.Body != newMessageBody {
		t.Error("Error: expected body", newMessageBody, "but got", mx.Body)
		return
	}
}

func TestDeleteMessage(t *testing.T) {
	TestInsertMessage(t)
	defer CleanUpMessages(t)

	m, err := GetMessage(testMessageId)
	if err != nil {
		t.Error("Error getting message for delete:", err)
		return
	}

	if err := m.Delete(); err != nil {
		t.Error("Error deleting message:", err)
		return
	}

	mx, erx := GetMessage(testMessageId)
	if erx == nil {
		t.Error("Error: expected message to be deleted but found", mx)
		return
	}
}

func CleanUpMessages(t *testing.T) {
	if testMessageId == "" {
		return
	}

	m := Message{Id: testMessageId}
	if err := m.Delete(); err != nil {
		t.Error("Error cleaning up messages:", err)
	}

	CleanUpThread(t)
	CleanUpMailbox(t)
}
