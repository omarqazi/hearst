package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20151213032410(txn *sql.Tx) {
	sql := `
	create table messages (
		id uuid not null,
		thread_id uuid not null,
		sender_mailbox_id uuid not null,
		createdat timestamp with time zone not null,
		updatedat timestamp with time zone not null,
		expiresat timestamp with time zone not null,
		topic text,
		body text,
		labels jsonb,
		payload jsonb,
		constraint messages_pk primary key (id)
	)
	with (
		OIDS=FALSE
	);
	create index messages_thread on messages(thread_id);
	create index messages_sender on messages(sender_mailbox_id);
	create index on messages using gin (labels);
	create index on messages using gin (payload);
	`

	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error creating messages table:", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20151213032410(txn *sql.Tx) {
	if _, err := txn.Exec("drop table messages;"); err != nil {
		fmt.Println("Error dropping messages table:", err)
	}
}
