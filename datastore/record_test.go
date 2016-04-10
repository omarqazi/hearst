package datastore

import "testing"

func TestRecord(t *testing.T) {
	record := Record{}
	record.RequireId()
	if record.Id == "" {
		t.Fatal("Expected record to have id but field was blank")
	}
}
