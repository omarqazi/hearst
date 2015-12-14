package datastore

import (
	"testing"
)

var testMailboxId string

const devicePushNotificationId = "push-notification-id"
const devicePublicKey = "BeginRSAKey"

func TestMailboxInsert(t *testing.T) {
	mb := Mailbox{
		DeviceId:  devicePushNotificationId,
		PublicKey: devicePublicKey,
	}

	if err := mb.Insert(); err != nil {
		t.Error("Error inserting mailbox", err)
	}

	CleanupMailbox(t)
	testMailboxId = mb.Id
}

func TestMailboxSelect(t *testing.T) {
	TestMailboxInsert(t)
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Error("Error getting mailbox", err)
		return
	}

	if mb.Id == "" {
		t.Error("Expected ID got got blank string on select")
		return
	}

	if mb.PublicKey != devicePublicKey {
		t.Error("Error: expected public key", devicePublicKey, "but got", mb.PublicKey)
		return
	}

	if mb.DeviceId != devicePushNotificationId {
		t.Error("Error expected device id", devicePushNotificationId, "but got", mb.DeviceId)
		return
	}
}

func TestMailboxUpdate(t *testing.T) {
	TestMailboxInsert(t)
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Error("Error getting mailbox for update:", err)
		return
	}

	newPublicKey := "NewPublicKey"
	mb.PublicKey = newPublicKey
	if err := mb.Update(); err != nil {
		t.Error("Error updating mailbox:", err)
		return
	}

	mbx, erx := GetMailbox(testMailboxId)
	if erx != nil {
		t.Error("Error getting mailbox after update:", err)
		return
	}

	if mbx.PublicKey != newPublicKey {
		t.Error("Error: Expected public key", newPublicKey, "but got", mbx.PublicKey)
		return
	}
}

func TestMailboxDelete(t *testing.T) {
	TestMailboxInsert(t)
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Error("Error getting mailbox for delete:", err)
		return
	}

	if err := mb.Delete(); err != nil {
		t.Error("Error deleting mailbox:", err)
		return
	}

	mbx, erx := GetMailbox(testMailboxId)
	if erx == nil {
		t.Error("Error: expected error getting deleted mailbox but it was still found", mbx)
		return
	}
}

func CleanupMailbox(t *testing.T) {
	if testMailboxId != "" {
		mb, err := GetMailbox(testMailboxId)
		if err == nil {
			erx := mb.Delete()
			if erx != nil {
				t.Error("Error cleaning up mailbox:", erx)
			} else {
				testMailboxId = ""
			}
		}
	}
}
