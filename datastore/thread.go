package datastore

import (
	"errors"
	"time"
)

type Thread struct {
	Id         string // a unique UUID
	Identifier string // a human readable name for the thread
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Subject    string
}

type ThreadMember struct {
	ThreadId          string `db:"thread_id" json:"-"`
	MailboxId         string `db:"mailbox_id"`
	AllowRead         bool   `db:"allow_read"`
	AllowWrite        bool   `db:"allow_write"`
	AllowNotification bool   `db:"allow_notification"`
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
	t.FillMissing()

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		insert into threads (id, createdat, updatedat, subject, identifier)
		VALUES (:id, now(), now(), :subject, :identifier)
	`, t)
	err := tx.Commit()
	Stream.AnnounceEvent("thread-insert-"+t.Id, t)
	return err
}

func (t *Thread) Update() error {
	if t.Id == "" {
		return t.Insert()
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		update threads set updatedat = now(), subject = :subject, identifier = :identifier where id = :id
	`, t)
	err := tx.Commit()
	Stream.AnnounceEvent("thread-update-"+t.Id, t)
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
	Stream.AnnounceEvent("thread-delete-"+t.Id, t)
	return err
}

func (t *Thread) GetMember(mailboxId string) (ThreadMember, error) {
	members := []ThreadMember{}
	err := PostgresDb.Select(&members, "select * from thread_members where mailbox_id = $1 and thread_id = $2", mailboxId, t.Id)
	if err != nil {
		return ThreadMember{}, err
	} else if len(members) > 0 {
		return members[0], nil
	}

	return ThreadMember{}, errors.New("No member found with that mailbox id")
}

func (t *Thread) GetAllMembers() ([]ThreadMember, error) {
	members := []ThreadMember{}
	err := PostgresDb.Select(&members, "select * from thread_members where thread_id = $1", t.Id)
	return members, err
}

func (t *Thread) AddMember(m *ThreadMember) error {
	m.ThreadId = t.Id
	if m.MailboxId == "" {
		return errors.New("Invalid mailbox ID for new member")
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		insert into thread_members 
		(thread_id, mailbox_id, allow_read, allow_write, allow_notification)
		VALUES (:thread_id, :mailbox_id, :allow_read, :allow_write, :allow_notification);
	`, m)
	err := tx.Commit()
	Stream.AnnounceEvent("threadmember-insert-"+t.Id, m)
	return err
}

func (m *ThreadMember) UpdatePermissions() error {
	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		update thread_members set allow_read = :allow_read, allow_write = :allow_write,
		allow_notification = :allow_notification 
		where thread_id = :thread_id and mailbox_id = :mailbox_id;
	`, m)
	err := tx.Commit()
	Stream.AnnounceEvent("threadmember-update-"+m.ThreadId, m)
	return err
}

func (m *ThreadMember) Remove() error {
	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		delete from thread_members
		where thread_id = :thread_id and mailbox_id = :mailbox_id;
	`, m)
	err := tx.Commit()
	Stream.AnnounceEvent("threadmember-delete-"+m.ThreadId, m)
	return err
}

// Function RequireId generates a UUID if the Id field is blank
// It returns the Id field, guaranteed to not be blank
func (t *Thread) RequireId() string {
	if t.Id == "" {
		t.GenerateUUID()
	}

	return t.Id
}

// Funciton RequireIdentifier sets the identifier of the thread
// to it's UUID if the identifier is blank.
// It returns the threads identifier, guaranteed to not be blank
func (t *Thread) RequireIdentifier() string {
	if t.Identifier == "" {
		t.Identifier = t.RequireId()
	}

	return t.Identifier
}

// Function FillMissing fills all missing fields required to insert
// the thread into the database.
// It returns the Id and Identifier fields of the thread
func (t *Thread) FillMissing() (string, string) {
	id := t.RequireId()
	identifier := t.RequireIdentifier()
	return id, identifier

}

func (t *Thread) GenerateUUID() string {
	newId := NewUUID()
	t.Id = newId
	return newId
}
