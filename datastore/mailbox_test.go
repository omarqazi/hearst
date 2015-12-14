package datastore

import (
	"github.com/omarqazi/hearst/auth"
	"testing"
)

var testMailboxId string

const devicePushNotificationId = "push-notification-id"

func TestMailboxInsert(t *testing.T) {
	privateKey, err := auth.GeneratePrivateKey(1024)
	if err != nil {
		t.Error("Error generating private key:", err)
		return
	}

	publicKeyString, err := auth.StringForPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Error("Error generating string for public key:", err)
		return
	}

	mb := &Mailbox{
		DeviceId:  devicePushNotificationId,
		PublicKey: publicKeyString,
	}

	if err := mb.Insert(); err != nil {
		t.Error("Error inserting mailbox", err)
	}

	CleanUpMailbox(t)
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

	newDeviceId := "NewDeviceId"
	mb.DeviceId = newDeviceId
	if err := mb.Update(); err != nil {
		t.Error("Error updating mailbox:", err)
		return
	}

	mbx, erx := GetMailbox(testMailboxId)
	if erx != nil {
		t.Error("Error getting mailbox after update:", err)
		return
	}

	if mbx.DeviceId != newDeviceId {
		t.Error("Error: Expected device id", newDeviceId, "but got", mbx.DeviceId)
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

func CleanUpMailbox(t *testing.T) {
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
