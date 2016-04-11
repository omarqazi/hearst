package datastore

import (
	"errors"
	"fmt"
	"strings"
)

type Thread struct {
	Record
	Identifier string // a human readable name for the thread
	Subject    string
	Domain     string // the domain of the server that owns this thread
}

type ThreadMember struct {
	ThreadId          string `db:"thread_id"`
	MailboxId         string `db:"mailbox_id"`
	AllowRead         bool   `db:"allow_read"`
	AllowWrite        bool   `db:"allow_write"`
	AllowNotification bool   `db:"allow_notification"`
}

func GetThread(uuid string) (t Thread, err error) {
	t.Record = Rec(uuid)
	t.Identifier = uuid
	err = t.Load()
	return
}

func (t *Thread) Insert() error {
	t.FillMissing()

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		insert into threads (id, createdat, updatedat, subject, identifier, domain)
		VALUES (:id, now(), now(), :subject, :identifier, :domain);
	`, t)
	tx.Exec(fmt.Sprintf("create sequence %s;", t.SequenceName()))
	err := tx.Commit()
	Stream.AnnounceEvent("thread-insert-"+t.Id, t)
	return err
}

func (t *Thread) Load() error {
	tx := []Thread{}
	db := PostgresDb.Unsafe()
	err := db.Select(&tx, "select * from threads where id = $1 limit 1", t.Record.Id)
	if err != nil {
		err = db.Select(&tx, "select * from threads where identifier = $1 limit 1", t.Identifier)
	}

	if len(tx) == 0 {
		return errors.New("No thread found with that UUID")
	}

	tdb := tx[0]
	t.Record = tdb.Record
	t.Domain = tdb.Domain
	t.Identifier = tdb.Identifier
	t.Subject = tdb.Subject
	return nil
}

func (t *Thread) SequenceName() (sname string) {
	sname = strings.Replace("log-counter-"+t.Id, "-", "_", -1)
	return
}

func (t *Thread) Update() error {
	if t.Id == "" {
		return t.Insert()
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		update threads set updatedat = now(), subject = :subject, identifier = :identifier, domain = :domain where id = :id
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
	tx.Exec(fmt.Sprintf("drop sequence %s;", t.SequenceName()))
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

func (m *ThreadMember) Insert() (err error) {
	if m.ThreadId == "" {
		return errors.New("No thread ID in new thread member")
	}

	t := Thread{Record: Rec(m.ThreadId)}
	if err = t.AddMember(m); err != nil {
		return
	}

	return
}

func (m *ThreadMember) Load() (err error) {
	members := []ThreadMember{}
	err = PostgresDb.Select(&members, "select * from thread_members where mailbox_id = $1 and thread_id = $2", m.MailboxId, m.ThreadId)
	if err != nil {
		return
	} else if len(members) > 0 {
		dbm := members[0]
		m.AllowRead = dbm.AllowRead
		m.AllowWrite = dbm.AllowWrite
		m.AllowNotification = dbm.AllowNotification
		return nil
	}

	return errors.New("No member found with that mailbox id")
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

func (m *ThreadMember) Update() (err error) {
	err = m.UpdatePermissions()
	return
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

func (t Thread) PermissionThreadId() string {
	return t.Id
}

func (m ThreadMember) PermissionThreadId() string {
	return m.ThreadId
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
	t.Domain = "chat.smick.co"
	return id, identifier

}
