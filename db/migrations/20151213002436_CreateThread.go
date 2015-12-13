package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20151213002436(txn *sql.Tx) {
	sql := `
	create table threads (
		id uuid not null,
		createdat timestamp with time zone not null,
		updatedat timestamp with time zone not null,
		subject text,
		constraint threads_pk primary key (id)
	)
	with (
		OIDS=FALSE
	);
	create index threads_updated on threads(updatedat);
	create index threads_uuid on threads(id);
	`
	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error creating threads table:", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20151213002436(txn *sql.Tx) {
	if _, err := txn.Exec("drop table threads;"); err != nil {
		fmt.Println("Error dropping threads table:", err)
	}
}
