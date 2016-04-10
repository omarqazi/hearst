package datastore

import (
	"crypto/rsa"
	"errors"
	"github.com/omarqazi/hearst/auth"
	"time"
)

type Mailbox struct {
	Record
	ConnectedAt time.Time
	PublicKey   string `db:"public_key"`
	DeviceId    string `db:"device_id"`
}

func NewMailbox() (mb Mailbox) {
	mb.RequireId()
	return
}

func NewMailboxWithKey() (Mailbox, *rsa.PrivateKey, error) {
	mb := NewMailbox()

	clientKey, err := auth.GeneratePrivateKey(2048)
	if err != nil {
		return mb, nil, err
	}

	pubKey, err := auth.StringForPublicKey(&clientKey.PublicKey)
	if err != nil {
		return mb, clientKey, err
	}

	mb.PublicKey = pubKey
	return mb, clientKey, nil
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

	return Mailbox{}, errors.New("No mailbox found with that UUID")
}

// Function GenerateNewKey generates a new private key and sets the mailboxes
// public key to match. It returns the newly generated private key
func (mb *Mailbox) GenerateNewKey() (key *rsa.PrivateKey, err error) {
	if key, err = auth.GeneratePrivateKey(2048); err != nil { // generate a new key
		return
	}

	mb.PublicKey, err = auth.StringForPublicKey(&key.PublicKey)
	return
}

func (mb *Mailbox) SessionToken(duration time.Duration, mailboxKey *rsa.PrivateKey, serverSessionKey *rsa.PrivateKey) (string, error) {
	token, err := auth.NewToken(serverSessionKey)
	if err != nil {
		return "", err
	}

	session := auth.Session{
		Token:    token,
		Duration: duration,
	}
	sigBytes, err := session.SignatureFor(mailboxKey)
	if err != nil {
		return "", err
	}

	session.Signature = sigBytes
	return session.String(), nil
}

// Function Insert executes an SQL insert statement
// to add the mailbox to the database
func (mb *Mailbox) Insert() error {
	mb.RequireId()

	tx := PostgresDb.MustBegin()
	tx.NamedExec("insert into mailboxes (id, createdat, updatedat, connectedat, public_key, device_id) VALUES (:id, now(), now(), now(), :public_key, :device_id)", mb)
	err := tx.Commit()
	Stream.AnnounceEvent("mailbox-insert-"+mb.Id, mb)
	return err
}

func (mb *Mailbox) Update() (err error) {
	err = mb.ExecuteUpdateQuery(`
		update mailboxes set updatedat = now(), connectedat = now(),
		public_key = :public_key, device_id = :device_id where id = :id;
	`)
	return
}

func (mb *Mailbox) StillConnected() (err error) {
	err = mb.ExecuteUpdateQuery("update mailboxes set updatedat = now(), connectedat = now() where id = :id;")
	return
}

func (mb *Mailbox) ExecuteUpdateQuery(query string) error {
	if mb.Id == "" {
		return mb.Insert()
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(query, mb)
	err := tx.Commit()
	Stream.AnnounceEvent("mailbox-update-"+mb.Id, mb)
	return err
}

func (mb *Mailbox) Delete() error {
	if mb.Id == "" {
		return errors.New("Cant delete mailbox with no UUID")
	}

	tx := PostgresDb.MustBegin()
	tx.NamedExec(`
		delete from mailboxes where id = :id;
	`, mb)
	err := tx.Commit()
	Stream.AnnounceEvent("mailbox-delete-"+mb.Id, mb)
	return err
}

func (mb *Mailbox) CanRead(threadId string) bool {
	dbThread := Thread{Id: threadId}
	member, err := dbThread.GetMember(mb.Id)
	if err != nil || !member.AllowRead {
		return false
	}

	return true
}

func (mb *Mailbox) CanWrite(threadId string) bool {
	dbThread := Thread{Id: threadId}
	member, err := dbThread.GetMember(mb.Id)
	if err != nil || !member.AllowWrite {
		return false
	}

	return true
}

func (mb *Mailbox) CanFollow(threadId string) bool {
	dbThread := Thread{Id: threadId}
	member, err := dbThread.GetMember(mb.Id)
	if err != nil || !member.AllowNotification {
		return false
	}

	return true
}
