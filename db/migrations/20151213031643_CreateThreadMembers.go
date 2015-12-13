package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20151213031643(txn *sql.Tx) {
	sql := `
	create table thread_members (
		thread_id uuid not null,
		mailbox_id uuid not null,
		allow_read boolean default false,
		allow_write boolean default false,
		allow_notification boolean default false,
		constraint thread_members_pk primary key (thread_id, mailbox_id)
	)
	with (
		OIDS=FALSE
	);
	`
	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error creating thread members table:", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20151213031643(txn *sql.Tx) {
	if _, err := txn.Exec("drop table thread_members;"); err != nil {
		fmt.Println("Error dropping thread_members table:", err)
	}
}
