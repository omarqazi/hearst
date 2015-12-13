package datastore

import (
	"code.google.com/p/go-uuid/uuid"
	"time"
)

type Mailbox struct {
	Id          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	ConnectedAt time.Time
	PublicKey   string `db:"public_key"`
	DeviceId    string `db:"device_id"`
}

// Function GetMailbox retrieves a Mailbox
// given a UUID. Returns mailbox and error
func GetMailbox(uuid string) (Mailbox, error) {
	mbx := []Mailbox{}
	db := PostgresDb.Unsafe()
	err := db.Select(&mbx, "select * from mailboxes where id = $1", uuid)
	if err != nil {
		return Mailbox{}, err
	} else if len(mbx) > 0 {
		return mbx[0], nil
	}

	return Mailbox{}, nil
}

// Function Insert executes an SQL insert statement
// to add the mailbox to the database
func (mb *Mailbox) Insert() error {
	mb.RequireId()

	tx := PostgresDb.MustBegin()
	tx.NamedExec("insert into mailboxes (id, createdat, updatedat, connectedat, public_key, device_id) VALUES (:id, now(), now(), now(), :public_key, :device_id)", mb)
	err := tx.Commit()
	return err
}

func (mb *Mailbox) Update() error {
	if mb.Id == "" {
		return mb.Insert()
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		update mailboxes set updatedat = now(), connectedat = now(),
		public_key = :public_key, device_id = :device_id where id = :id
	`, mb)
	err := tx.Commit()
	return err
}

// Function RequireId generates a UUID if the
// Id of the Mailbox is a blank string
func (mb *Mailbox) RequireId() string {
	if mb.Id == "" {
		mb.GenerateUUID()
	}

	return mb.Id
}

// Function GenerateUUID generates a new UUID,
// sets it as the id of the calling struct,
// and returns the newly generated UUID
func (mb *Mailbox) GenerateUUID() string {
	newId := uuid.New()
	mb.Id = newId
	return newId
}