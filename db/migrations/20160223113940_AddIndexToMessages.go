package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20160223113940(txn *sql.Tx) {
	sql := `
	alter table messages add column index integer not null default 0;
	`
	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error adding index column to messages table", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20160223113940(txn *sql.Tx) {
	if _, err := txn.Exec("alter table threads drop column index;"); err != nil {
		fmt.Println("Error dropping index column from messages table:", err)
	}
}
