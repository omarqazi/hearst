package datastore

import "testing"

func TestPostgresConnect(t *testing.T) {
	if err := ConnectPostgres(); err != nil {
		t.Error("Error connecting to postgres:", err)
	}
}

func TestRedisConnect(t *testing.T) {
	if err := ConnectRedis(); err != nil {
		t.Error("Error connecting to redis:", err)
	}
}

func TestGenerateUUID(t *testing.T) {
	uuida := NewUUID()
	if uuida == "" {
		t.Error("Error: expected UUID but got blank string")
		return
	}

	uuidb := NewUUID()
	if uuida == uuidb {
		t.Error("Error: expected UUID to be unique but got same string twice")
		return
	}
}
