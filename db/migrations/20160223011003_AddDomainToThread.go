package main

import (
	"database/sql"
	"fmt"
)

// Up is executed when this migration is applied
func Up_20160223011003(txn *sql.Tx) {
	sql := `
	alter table threads add column domain text;
	`
	if _, err := txn.Exec(sql); err != nil {
		fmt.Println("Error adding domain column to thread table", err)
	}
}

// Down is executed when this migration is rolled back
func Down_20160223011003(txn *sql.Tx) {
	if _, err := txn.Exec("alter table threads drop column domain;"); err != nil {
		fmt.Println("Error dropping domain column from threads table:", err)
	}
}
