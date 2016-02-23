package datastore

import (
	"testing"
)

var testThreadId = ""

const threadSubject = "Hello, World!"

func TestInsertThread(t *testing.T) {
	tr := Thread{
		Subject: threadSubject,
	}

	if err := tr.Insert(); err != nil {
		t.Error("Error inserting thread:", err)
	}

	CleanUpThread(t)
	testThreadId = tr.Id
}

// The thread has a field called identifier which should
// have a uniqueness constarint across the table enforced
// by the database engine
func TestThreadIdentifierUnique(t *testing.T) {
	appleIdentifier := "apple-identifier"
	orangeIdentifier := "orange-identifier"

	tr := Thread{
		Subject:    threadSubject,
		Identifier: appleIdentifier,
	}

	if err := tr.Insert(); err != nil {
		t.Fatal("Error inserting thread that should be unique:", err)
	}

	trx := Thread{
		Subject:    "some random subject",
		Identifier: appleIdentifier,
	}

	if err := trx.Insert(); err == nil {
		tr.Delete()
		trx.Delete()
		t.Fatal("Expected insert of already assigned identifier to fail, but it did not:", trx)
	}

	dbTr, err := GetThread(tr.Id)
	if err != nil {
		tr.Delete()
		t.Fatal("Error getting thread that was supposedly inserted while testing uniqueness:", err)
	}

	if dbTr.Subject != threadSubject {
		dbTr.Delete()
		t.Fatal("Error: Expected thread to have subject", threadSubject, "but found", dbTr.Subject)
	}

	trx.Identifier = orangeIdentifier
	if err := trx.Insert(); err != nil {
		tr.Delete()
		t.Fatal("Expected thread insert to succeed after changing to unique identifier but it did not")
	}

	tr.Delete()
	trx.Delete()
}

func TestSelectThread(t *testing.T) {
	TestInsertThread(t)
	defer CleanUpThread(t)

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread:", err)
		return
	}

	if tr.Subject != threadSubject {
		t.Error("Error: Expected subject", threadSubject, "but got", tr.Subject)
		return
	}
}

func TestSelectThreadByIdentifier(t *testing.T) {
	TestInsertThread(t)
	defer CleanUpThread(t)

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Fatal("Error getting thread:", err)
	}

	tr.Identifier = "select-thread-identifier"
	if err := tr.Update(); err != nil {
		t.Fatal("Error updating thread when testing seelct by identifier:", err)
	}

	ti, err := GetThread(tr.Identifier)
	if err != nil {
		t.Fatal("Could not load thread by identifier:", err)
	}

	if ti.Id != tr.Id {
		t.Fatal("Expected to find thread", tr.Id, "but found thread", ti.Id)
	}
}

func TestUpdateThread(t *testing.T) {
	TestInsertThread(t)
	defer CleanUpThread(t)

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for update:", err)
		return
	}

	newSubject := "Some New Subject"
	tr.Subject = newSubject
	if err := tr.Update(); err != nil {
		t.Error("Error updating thread:", err)
	}

	trx, erx := GetThread(testThreadId)
	if erx != nil {
		t.Error("Error getting thread after update:", erx)
		return
	}

	if newSubject != trx.Subject {
		t.Error("Error: expected subject", newSubject, "but got", trx.Subject)
		return
	}
}

func TestDeleteThread(t *testing.T) {
	TestInsertThread(t)
	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for delete:", err)
		return
	}

	if err := tr.Delete(); err != nil {
		t.Error("Error deleting thread:", err)
		return
	}

	trx, erx := GetThread(testThreadId)
	if erx == nil {
		t.Error("Expected thread to be deleted but found", trx)
		return
	}
}

func TestAddMember(t *testing.T) {
	TestInsertThread(t)
	defer CleanUpThread(t)

	tr, err := GetThread(testThreadId)
	if err != nil {
		t.Error("Error getting thread for add member:", err)
		return
	}

	TestMailboxInsert(t)
	mb, err := GetMailbox(testMailboxId)
	if err != nil {
		t.Error("Error getting mailbox on member add:", err)
		return
	}

	member := ThreadMember{
		MailboxId:         mb.Id,
		AllowRead:         true,
		AllowWrite:        false,
		AllowNotification: true,
	}

	allMembers, err := tr.GetAllMembers()
	if err != nil {
		t.Error("Error getting thread members:", err)
		return
	}

	originalNumberOfMembers := len(allMembers)
	if originalNumberOfMembers != 0 {
		t.Error("Expected 0 thread members but found", originalNumberOfMembers)
		return
	}

	if err := tr.AddMember(&member); err != nil {
		t.Error("Error adding member:", member)
		return
	}

	allMembersx, err := tr.GetAllMembers()
	if err != nil {
		t.Error("Error getting members after add:", err)
		return
	}

	updatedNumberOfMembers := len(allMembersx)
	if expected := originalNumberOfMembers + 1; updatedNumberOfMembers != expected {
		t.Error("Expected", expected, "members after add but found", updatedNumberOfMembers)
		return
	}

	dbMember := allMembersx[0]
	if dbMember.AllowRead != member.AllowRead || dbMember.AllowWrite != member.AllowWrite || dbMember.AllowNotification != member.AllowNotification {
		t.Error("Expected member", member, "but got", dbMember)
		return
	}

	dbMember.AllowWrite = true
	if err := dbMember.UpdatePermissions(); err != nil {
		t.Error("Error updating permissions:", err)
		return
	}

	directMember, err := tr.GetMember(dbMember.MailboxId)
	if err != nil {
		t.Error("Error getting members after update:", err)
		return
	}

	if directMember.AllowWrite != true {
		t.Error("Expected AllowWrite = true after update")
		return
	}

	if err := directMember.Remove(); err != nil {
		t.Error("Error removing member:", err)
		return
	}

	allMembersy, err := tr.GetAllMembers()
	if err != nil {
		t.Error("Error getting members after remove", err)
		return
	}

	updatedNumberOfMembers = len(allMembersy)
	if updatedNumberOfMembers != originalNumberOfMembers {
		t.Error("Expected", originalNumberOfMembers, "members", "but found", updatedNumberOfMembers)
		return
	}

	missingMember, err := tr.GetMember(mb.Id)
	if err == nil {
		t.Error("Expected error getting member after delete but got", missingMember)
	}

	CleanUpMailbox(t)
}

func CleanUpThread(t *testing.T) {
	if testThreadId != "" {
		tr, err := GetThread(testThreadId)
		if err == nil {
			erx := tr.Delete()
			if erx != nil {
				t.Error("Error cleanig up thread:", erx)
			} else {
				testThreadId = ""
			}
		}
	}
}
