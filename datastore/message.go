package datastore

import (
	"errors"
	"github.com/jmoiron/sqlx/types"
	"time"
)

type Message struct {
	Id              string
	ThreadId        string `db:"thread_id"`
	SenderMailboxId string `db:"sender_mailbox_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	ExpiresAt       time.Time
	Topic           string
	Body            string
	Labels          types.JsonText
	Payload         types.JsonText
}

func GetMessage(uuid string) (Message, error) {
	mx := []Message{}
	err := PostgresDb.Select(&mx, "select * from messages where id = $1", uuid)
	if err != nil {
		return Message{}, err
	} else if len(mx) > 0 {
		return mx[0], nil
	}

	return Message{}, errors.New("No message found with that UUID")
}

func (m *Message) Insert() error {
	m.RequireId()

	tx := PostgresDb.MustBegin()
	_, err := tx.NamedExec(`
		insert into messages
		(id, thread_id, sender_mailbox_id, createdat, updatedat, expiresat, topic, body, labels, payload)
		VALUES
		(:id, :thread_id, :sender_mailbox_id, now(), now(), :expiresat, :topic, :body, :labels, :payload);
	`, m)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func (m *Message) Update() error {
	if m.Id == "" {
		return m.Insert()
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		update messages set updatedat = now(), expiresat = :expiresat, topic = :topic, body = :body,
		labels = :labels, payload = :payload where id = :id;
	`, m)
	err := tx.Commit()
	return err
}

func (m *Message) Delete() error {
	if m.Id == "" {
		return errors.New("Cannot delete message with no UUID")
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		delete from messages where id = :id;
	`, m)
	err := tx.Commit()
	return err
}

func (m *Message) RequireId() string {
	if m.Id == "" {
		m.GenerateUUID()
	}

	return m.Id
}

func (m *Message) GenerateUUID() string {
	newId := NewUUID()
	m.Id = newId
	return newId
}
