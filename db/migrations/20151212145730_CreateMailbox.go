package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20151212145730(txn *sql.Tx) {
	sql := `
	create extension postgis;
	create table mailboxes (
		id uuid not null,
		createdat timestamp with time zone not null,
		updatedat timestamp with time zone not null,
		connectedat timestamp with time zone not null,
		public_key text not null,
		device_id text,
		constraint mailboxes_pk primary key (id)
	)
	with (
		OIDS=FALSE
	);
	alter table mailboxes add column location geometry(Point,4326);
	create index mailboxes_gist on mailboxes using GIST(location);
	create index mailboxes_updated on mailboxes(updatedat);
	`
	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error creating mailboxes table:", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20151212145730(txn *sql.Tx) {
	if _, err := txn.Exec("drop table mailboxes;"); err != nil {
		fmt.Println("Error dropping mailboxes table:", err)
	}
}
