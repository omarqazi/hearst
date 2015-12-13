package datastore

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

const connectionString = "dbname=hearst user=postgres password=postgres sslmode=disable"

var PostgresDb *sqlx.DB = nil

func init() {
	ConnectPostgres()
}

func ConnectPostgres() {
	var err error
	PostgresDb, err = sqlx.Open("postgres", connectionString)
	if err != nil {
		log.Fatalln("Error connecting to postgres:", err)
	}
}
