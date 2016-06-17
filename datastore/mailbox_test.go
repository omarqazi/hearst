package datastore

import (
	"github.com/omarqazi/hearst/auth"
	"testing"
	"time"
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

// mb.StillConnected should update the UpdatedAt and ConnectedAt fields,
// but not any other fields.
func TestMailboxStillConnected(t *testing.T) {
	TestMailboxInsert(t) // create a new mailbox
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Fatal("Error getting mailbox for connected update", err)
	}

	// this test could be better implemented by setting the connectedat time
	// manually to something else, rather than sleeping. but i'm lazy
	time.Sleep(1 * time.Second)

	mb.DeviceId = ""
	mb.PublicKey = ""
	if err := mb.StillConnected(); err != nil {
		t.Fatal("Error telling database that we are still connected", err)
	}

	mbx, err := GetMailbox(testMailboxId)
	if mbx.ConnectedAt == mb.ConnectedAt {
		t.Fatal("Expected ConnectedAt to change but new time was", mbx.ConnectedAt, "and old time was", mb.ConnectedAt)
	}
	if mbx.UpdatedAt == mb.UpdatedAt {
		t.Fatal("Expected UpdatedAt to change but new time was", mbx.UpdatedAt, "and old time was", mb.UpdatedAt)
	}

	if mbx.DeviceId == "" || mbx.PublicKey == "" {
		t.Fatal("Error: DeviceId or PublicKey changed when calling still connected", mbx)
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

func TestRecentThreads(t *testing.T) {
	mb := Mailbox{DeviceId: "testbox"}
	if err := mb.Insert(); err != nil {
		t.Fatal("Error inserting mailbox to test recent threads")
	}

	threads := []Thread{
		{Subject: "Game of Thrones"},
		{Subject: "Mad Men"},
		{Subject: "The Wicker Man"},
	}

	for i := range threads {
		threads[i].Insert()
		err := threads[i].AddMember(&ThreadMember{
			MailboxId:         mb.Id,
			AllowRead:         true,
			AllowWrite:        true,
			AllowNotification: true,
		})
		if err != nil {
			t.Fatal("Error adding thread member:", err)
		}
	}

	cutoffTime := time.Now()
	recentThreads, err := mb.RecentThreads(time.Unix(0, 0), 100, 0)
	if err != nil {
		t.Fatal("Error calling mailbox.RecentThreads:", err)
	}

	if len(recentThreads) != len(threads) {
		t.Fatal("Expected to find", len(threads), "thread but found", len(recentThreads))
	}

	for i, tr := range recentThreads {
		inversePosition := len(threads) - i - 1
		if tr.Subject != threads[inversePosition].Subject {
			t.Fatal("Expected threads to be returned in order updated")
		}
	}

	cutoffThreads, err := mb.RecentThreads(cutoffTime, 100, 0)
	if err != nil {
		t.Fatal("Error calling recent threads again:", err)
	}

	if len(cutoffThreads) > 0 {
		t.Fatal("Expected no threads to be past cutoff, but found", len(cutoffThreads))
	}

	threadToUpdate := threads[0]
	message := Message{
		ThreadId:        threadToUpdate.Id,
		SenderMailboxId: mb.Id,
		Topic:           "test-message",
		Body:            "Bump me to the top",
		Labels:          []byte("{}"),
		Payload:         []byte("{}"),
	}

	if err := message.Insert(); err != nil {
		t.Fatal("Error inserting message:", err)
	}

	cutoffThreads, err = mb.RecentThreads(cutoffTime, 100, 0)
	if err != nil {
		t.Fatal("Error calling recent threads after message save:", err)
	}

	if len(cutoffThreads) != 1 {
		t.Fatal("Expected 1 updated thread but found", len(cutoffThreads))
	}

	if cutoffThreads[0].Subject != threadToUpdate.Subject {
		t.Fatal("First thread returned was not most recently updated")
	}

	recentThreads, err = mb.RecentThreads(time.Unix(0, 0), 1, 0)
	if len(recentThreads) != 1 {
		t.Fatal("Expected 1 thread after limit but found", len(recentThreads))
	}

	if recentThreads[0].Subject != threadToUpdate.Subject {
		t.Fatal("First thread returned when limiting was not most recently updated")
	}

	recentThreads, err = mb.RecentThreads(time.Unix(0, 0), 1, 1)
	if len(recentThreads) != 1 {
		t.Fatal("Limit broken after applying offset")
	}

	if recentThreads[0].Subject == threadToUpdate.Subject {
		t.Fatal("Found same thread after offset")
	}

	for _, tr := range threads {
		tr.Delete()
	}

	if err := mb.Delete(); err != nil {
		t.Fatal("Error deleting mailbox:", err)
	}
}

func TestCanMethods(t *testing.T) {
	mb := Mailbox{}
	if err := mb.Insert(); err != nil {
		t.Fatal("Could not insert mailbox when test Can methods", err)
	}

	thread := Thread{}
	if err := thread.Insert(); err != nil {
		t.Fatal("Could not insert thread when testing Can methods", err)
	}

	member := &ThreadMember{
		MailboxId:         mb.Id,
		ThreadId:          thread.Id,
		AllowRead:         false,
		AllowWrite:        false,
		AllowNotification: false,
	}

	if mb.CanRead(thread.Id) || mb.CanWrite(thread.Id) || mb.CanFollow(thread.Id) {
		t.Fatal("Error: mailbox with no thread member can read write or follow")
	}

	if err := thread.AddMember(member); err != nil {
		t.Fatal("Error adding thread member", err)
	}

	if mb.CanRead(thread.Id) || mb.CanWrite(thread.Id) || mb.CanFollow(thread.Id) {
		t.Fatal("Error: mailbox with no permissions can read write or follow")
	}

	member.AllowRead = true
	if err := member.UpdatePermissions(); err != nil {
		t.Fatal("Error updating thread member permisions:", err)
	}

	if !mb.CanRead(thread.Id) {
		t.Fatal("Error: expected mailbox to be able to read thread, but CanRead returned false")
	}

	if mb.CanWrite(thread.Id) {
		t.Fatal("Error: expected mailbox to not be able to write to thread, but CanRead returned true")
	}

	if mb.CanFollow(thread.Id) {
		t.Fatal("Error: expected mailbox to not be able to follow thread, but CanFollow returned true")
	}

	member.AllowRead = false
	member.AllowWrite = true

	if err := member.UpdatePermissions(); err != nil {
		t.Fatal("Error updating member permissions when trying Can methods")
	}

	if mb.CanRead(thread.Id) {
		t.Fatal("Error: expected mailbox to not be able to read thread but CanRead returned true")
	}

	if !mb.CanWrite(thread.Id) {
		t.Fatal("Error: expected mailbox to be able to write to thread but CanWrite returned false")
	}

	if mb.CanFollow(thread.Id) {
		t.Fatal("Error: expected mailbox to not be able to follow thread but CanFollow returned true")
	}

	member.AllowWrite = false
	member.AllowNotification = true

	if err := member.UpdatePermissions(); err != nil {
		t.Fatal("Error updating member permissions when testing Can methods", err)
	}

	if mb.CanRead(thread.Id) {
		t.Fatal("Error: expected mailbox to not be able to read thread but CanRead returned true")
	}

	if mb.CanWrite(thread.Id) {
		t.Fatal("Error: expected mailbox to not be able to write thread but CanWrite returned true")
	}

	if !mb.CanFollow(thread.Id) {
		t.Fatal("Error: expected mailbox to be able to follow thread but CanFollow returned false")
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
