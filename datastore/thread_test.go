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

func TestSelectThread(t *testing.T) {
	TestInsertThread(t)
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

func TestUpdateThread(t *testing.T) {
	TestInsertThread(t)
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
