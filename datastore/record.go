package datastore

import (
	"time"
)

type Record struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Recordable interface {
	Load() error
	Insert() error
	Update() error
	Delete() error
	PermissionThreadId() string // A thread id used to control permissions for the object
}

func Rec(uuid string) (r Record) {
	r.Id = uuid
	return
}

// Function RequireId generates a UUID if the
// Id of the Mailbox is a blank string
func (r *Record) RequireId() string {
	if r.Id == "" {
		r.GenerateUUID()
	}

	return r.Id
}

// Function GenerateUUID generates a new UUID,
// sets it as the id of the calling struct,
// and returns the newly generated UUID
func (r *Record) GenerateUUID() string {
	newId := NewUUID()
	r.Id = newId
	return newId
}
