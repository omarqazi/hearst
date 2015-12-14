package datastore

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

const connectionString = "dbname=hearst user=postgres password=postgres sslmode=disable"

var PostgresDb *sqlx.DB = nil

func init() {
	if err := ConnectPostgres(); err != nil {
		log.Fatalln("Error connecting to postgres:", err)
	}
}

func ConnectPostgres() (err error) {
	PostgresDb, err = sqlx.Open("postgres", connectionString)
	return
}

func NewUUID() string {
	newId := uuid.New()
	return newId
}
