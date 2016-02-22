package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20160222130821(txn *sql.Tx) {
	sql := `
	alter table threads add column identifier text unique;
	`
	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error adding identifier column to thread table", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20160222130821(txn *sql.Tx) {
	if _, err := txn.Exec("alter table threads drop column identifier;"); err != nil {
		fmt.Println("Error dropping identifier column from threads table:", err)
	}
}
