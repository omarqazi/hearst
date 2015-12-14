package datastore

import (
	"errors"
	"time"
)

type Thread struct {
	Id        string
	CreatedAt time.Time
	UpdatedAt time.Time
	Subject   string
}

type ThreadMember struct {
	ThreadId          string
	MailboxId         string
	AllowRead         bool
	AllowWrite        bool
	AllowNotification bool
}

func GetThread(uuid string) (Thread, error) {
	tx := []Thread{}
	db := PostgresDb.Unsafe()
	err := db.Select(&tx, "select * from threads where id = $1", uuid)
	if err != nil {
		return Thread{}, err
	} else if len(tx) > 0 {
		return tx[0], nil
	}

	return Thread{}, errors.New("No thread found with that UUID")
}

func (t *Thread) Insert() error {
	t.RequireId()

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		insert into threads (id, createdat, updatedat, subject)
		VALUES (:id, now(), now(), :subject)
	`, t)
	err := tx.Commit()
	return err
}

func (t *Thread) Update() error {
	if t.Id == "" {
		return t.Insert()
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		update threads set updatedat = now(), subject = :subject where id = :id
	`, t)
	err := tx.Commit()
	return err
}

func (t *Thread) Delete() error {
	if t.Id == "" {
		return errors.New("Cant delete thread with no UUID")
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		delete from threads where id = :id
	`, t)
	err := tx.Commit()
	return err
}

func (t *Thread) RequireId() string {
	if t.Id == "" {
		t.GenerateUUID()
	}

	return t.Id
}

func (t *Thread) GenerateUUID() string {
	newId := NewUUID()
	t.Id = newId
	return newId
}
