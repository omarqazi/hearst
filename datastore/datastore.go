package datastore

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/redis.v3"
	"log"
)

const connectionString = "dbname=hearst user=postgres password=postgres sslmode=disable"
const redisHost = "localhost:6379"

var PostgresDb *sqlx.DB = nil
var RedisDb *redis.Client
var Stream EventStream

func init() {
	if err := ConnectPostgres(); err != nil {
		log.Fatalln("Error connecting to postgres:", err)
	}

	if err := ConnectRedis(); err != nil {
		log.Fatalln("Error connecting to redis:", err)
	}

	Stream = EventStream{
		Client: RedisDb,
	}
}

func ConnectPostgres() (err error) {
	PostgresDb, err = sqlx.Open("postgres", connectionString)
	return
}

func ConnectRedis() (err error) {
	RedisDb = redis.NewClient(&redis.Options{
		Addr: redisHost,
	})

	_, err = RedisDb.Ping().Result()
	return

}

func NewUUID() string {
	newId := uuid.New()
	return newId
}
