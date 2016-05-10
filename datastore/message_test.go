package datastore

import (
	"github.com/jmoiron/sqlx/types"
	"testing"
)

var testMessageId = ""
var testMessageTopic = "chat-message"

const testMessageBody = "hey man"

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

func prepareToSetupTestMessages(t *testing.T) (Mailbox, Thread) {
	TestMailboxInsert(t)
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Error("Error getting mailbox for message insert:", err)
		return mb, Thread{}
	}

	TestInsertThread(t)
	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for message insert:", err)
		return mb, tr
	}

	return mb, tr
}

func createSampleMessages(tr Thread, mb Mailbox, topic string, body string, totalMessages int, t *testing.T) {
	for i := 0; i < totalMessages; i++ {
		m := Message{
			ThreadId:        tr.Id,
			SenderMailboxId: mb.Id,
			Topic:           topic,
			Body:            body,
			Labels:          types.JSONText("{}"),
			Payload:         types.JSONText("{}"),
		}

		m.Labels.Scan("{}")
		m.Payload.Scan("{}")

		if err := m.Insert(); err != nil {
			t.Error("Error inserting message:", err)
			return
		}
	}
}
func setupTestMessages(totalMessages int, t *testing.T) {
	mb, tr := prepareToSetupTestMessages(t)
	createSampleMessages(tr, mb, testMessageTopic, testMessageBody, totalMessages, t)
}

func setupDualTopicTestMessages(t *testing.T) {
	mb, tr := prepareToSetupTestMessages(t)
	createSampleMessages(tr, mb, "chat-message", testMessageBody, 60, t)
	createSampleMessages(tr, mb, "location-update", testMessageBody, 40, t)
}

func TestRecentMessages(t *testing.T) {
	setupTestMessages(100, t)

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for message insert:", err)
		return
	}

	messages, err := tr.RecentMessages(1000)
	if err != nil {
		t.Fatal("Error getting recent messages:", err)
	}

	if len(messages) != 100 {
		t.Fatal("Expected 100 recent messages but got", len(messages))
	}

	CleanUpMessages(t)
}

func TestRecentMessagesWithTopic(t *testing.T) {
	setupDualTopicTestMessages(t)
	originalTopic := "chat-message"
	newTopic := "location-update"

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for message insert:", err)
		return
	}

	messages, err := tr.RecentMessagesWithTopic(originalTopic, 1000)
	if err != nil {
		t.Fatal("Error getting recent messages with topic:", err)
	}

	if len(messages) != 60 {
		t.Fatal("Expected 60 messages with topic", originalTopic, "but got", len(messages))
	}

	messages, err = tr.RecentMessagesWithTopic(newTopic, 1000)
	if err != nil {
		t.Fatal("Error getting recent messages with topic:", err)
	}

	if len(messages) != 40 {
		t.Fatal("Expected 40 messages with topic", originalTopic, "but got", len(messages))
	}

	CleanUpMessages(t)
}

func TestMessagesSince(t *testing.T) {
	setupDualTopicTestMessages(t)
	originalTopic := "chat-message"
	newTopic := "location-update"

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for message insert:", err)
		return
	}

	messages, err := tr.MessagesSince(0, 1000, originalTopic)
	if err != nil {
		t.Fatal("Error getting recent messages with topic:", err)
	}

	if len(messages) != 60 {
		t.Fatal("Expected 60 messages with topic", originalTopic, "but got", len(messages))
	}

	messages, err = tr.MessagesSince(0, 1000, newTopic)
	if err != nil {
		t.Fatal("Error getting recent messages with topic:", err)
	}

	if len(messages) != 40 {
		t.Fatal("Expected 40 messages with topic", originalTopic, "but got", len(messages))
	}

	messages, err = tr.MessagesSince(70, 1000, originalTopic)
	if err != nil {
		t.Fatal("Error getting recent messages with topic:", err)
	}

	if len(messages) >= 60 {
		t.Fatal("Expected less than 60 messages after providing sequence number but found", len(messages))
	}

	messages, err = tr.MessagesSince(70, 1000, newTopic)
	if err != nil {
		t.Fatal("Error getting messages since with topic:", err)
	}

	if len(messages) >= 40 {
		t.Fatal("Expected less than 40 messages after providing sequence number but found", len(messages))
	}

	messages, err = tr.MessagesSince(70, 1000, "")
	if err != nil {
		t.Fatal("Error getting messages since without topic:", err)
	}

	if len(messages) != 30 {
		t.Fatal("Error: Expected exactly 30 messages but found", len(messages))
	}

	CleanUpMessages(t)
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
