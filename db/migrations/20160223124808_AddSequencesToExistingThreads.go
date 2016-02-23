package main

import (
	"database/sql"
	"fmt"
	"github.com/omarqazi/hearst/datastore"
	"log"
)

func Up_20160223124808(txn *sql.Tx) {
	thread := datastore.Thread{}
	rows, err := datastore.PostgresDb.Queryx("select * from threads")
	if err != nil {
		log.Fatalln("Error getting threads", err)
	}

	for rows.Next() {
		err := rows.StructScan(&thread)
		if err != nil {
			log.Println("Error adding sequence for thread:", err)
			continue
		}

		// make sure there is a sequence
		tx := datastore.PostgresDb.MustBegin()
		tx.Exec("create sequence " + thread.SequenceName())
		err = tx.Commit()

		if err == nil {
			enumerateThread(thread)
		}
	}
}

func enumerateThread(t datastore.Thread) error {
	rows, err := datastore.PostgresDb.Queryx("select * from messages where thread_id = $1 order by createdat asc")
	if err != nil {
		return err
	}

	for message := (datastore.Message{}); rows.Next(); {
		err = rows.StructScan(&message)
		if err != nil {
			continue
		}

		tx := datastore.PostgresDb.MustBegin()
		tx.Exec(fmt.Sprintf("update messages set index = nextval('%s') where id = '%s'", t.SequenceName(), message.Id))
		tx.Commit()
	}

	return nil
}

func Down_20160223124808(txn *sql.Tx) {
}
