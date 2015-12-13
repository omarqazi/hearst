package datastore

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

const connectionString = "dbname=hearst user=postgres password=postgres sslmode=disable"

var PostgresDatabase *sql.DB = nil

func init() {
	ConnectPostgres()
}

func ConnectPostgres() {
	var err error
	PostgresDatabase, err = sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatalln("Error connecting to postgres:", err)
	}
}
