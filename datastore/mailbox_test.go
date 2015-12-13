package datastore

import (
	"testing"
)

var testObjectId string

const devicePushNotificationId = "push-notification-id"
const devicePublicKey = "BeginRSAKey"

func TestInsert(t *testing.T) {
	mb := Mailbox{
		DeviceId:  devicePushNotificationId,
		PublicKey: devicePublicKey,
	}

	if err := mb.Insert(); err != nil {
		t.Error("Error inserting mailbox", err)
	}

	testObjectId = mb.Id
}

func TestSelect(t *testing.T) {
	TestInsert(t)
	mb, err := GetMailbox(testObjectId)
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

func TestUpdate(t *testing.T) {
	TestInsert(t)
	mb, err := GetMailbox(testObjectId)
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

	mbx, erx := GetMailbox(testObjectId)
	if erx != nil {
		t.Error("Error getting mailbox after update:", err)
		return
	}

	if mbx.PublicKey != newPublicKey {
		t.Error("Error: Expected public key", newPublicKey, "but got", mbx.PublicKey)
		return
	}
}
