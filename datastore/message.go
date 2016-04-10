package datastore

import (
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx/types"
	"github.com/lib/pq"
	"time"
)

type Message struct {
	Id              string
	ThreadId        string `db:"thread_id"`
	SenderMailboxId string `db:"sender_mailbox_id"`
	CreatedAt       time.Time
	UpdatedAt       time.Time   `json:"-"`
	ExpiresAt       pq.NullTime `json:"-"`
	Topic           string
	Body            string
	Labels          types.JSONText
	Payload         types.JSONText
	Index           int
}

func (t *Thread) RecentMessagesWithTopic(topicFilter string, limit int) (mx []Message, err error) {
	mx = []Message{}
	if topicFilter == "" {
		topicFilter = "%"
	}

	err = PostgresDb.Select(&mx, `
	select * from (select * from messages where thread_id = $1 and topic LIKE $2 order by index desc limit $3) as sub order by index asc;
	`, t.Id, topicFilter, limit)
	return
}

func (t *Thread) RecentMessages(limit int) (mx []Message, err error) {
	mx = []Message{}
	err = PostgresDb.Select(&mx, `
		select * from (select * from messages where thread_id = $1 order by index desc limit $2) as sub order by index asc;
		`, t.Id, limit)
	return
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
	m.CreatedAt = time.Now()
	tx := PostgresDb.MustBegin()
	sequenceName := (&Thread{Record: Rec(m.ThreadId)}).SequenceName()
	query := fmt.Sprintf(`
	insert into messages 
		(id, thread_id, sender_mailbox_id, createdat, updatedat, expiresat, topic, body, labels, payload, index)
	VALUES
		(:id, :thread_id, :sender_mailbox_id, now(), now(), :expiresat, :topic, :body, :labels, :payload, nextval('%s'))
	`, sequenceName)
	_, err := tx.NamedExec(query, m)
	if err != nil {
		return err
	}

	_, err = tx.NamedExec(`
		update threads set updatedat = now() where id = :thread_id;
	`, m)
	if err != nil {
		return err
	}

	err = tx.Commit()
	Stream.AnnounceEvent("message-insert-"+m.ThreadId, m)
	return err
}

func (m *Message) Update() error {
	if m.Id == "" {
		return m.Insert()
	}

	tx := PostgresDb.MustBegin()
	_, err := tx.NamedExec(`
		update messages set updatedat = now(), expiresat = :expiresat, topic = :topic, body = :body,
		labels = :labels, payload = :payload where id = :id;
	`, m)
	if err != nil {
		return err
	}

	err = tx.Commit()
	Stream.AnnounceEvent("message-update-"+m.ThreadId, m)
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
	Stream.AnnounceEvent("message-delete-"+m.ThreadId, m)
	return err
}

func (m Message) PermissionThreadId() string {
	return m.ThreadId
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
