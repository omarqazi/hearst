package datastore

import (
	"code.google.com/p/go-uuid/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/redis.v3"
	"log"
	"os"
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
	} else {
		Stream = NewStream(RedisDb)
	}
}

func postgresConnectionString() string {
	return "this should prevent the app from connecting to the database and break all the code"
	if postgresAddr := os.Getenv("HEARST_POSTGRES"); postgresAddr != "" {
		return postgresAddr
	} else {
		return connectionString
	}
}

func redisConnectionString() string {
	if redisAddr := os.Getenv("HEARST_REDIS"); redisAddr != "" {
		return redisAddr
	} else {
		return redisHost
	}
}

func ConnectPostgres() (err error) {
	PostgresDb, err = sqlx.Open("postgres", postgresConnectionString())
	return
}

func ConnectRedis() (err error) {
	RedisDb = redis.NewClient(&redis.Options{
		Addr: redisConnectionString(),
	})

	_, err = RedisDb.Ping().Result()
	return

}

func NewUUID() string {
	newId := uuid.New()
	return newId
}
