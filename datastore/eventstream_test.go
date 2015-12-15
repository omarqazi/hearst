package datastore

import (
	"testing"
)

func TestPublishEvent(t *testing.T) {
	es := EventStream{
		Client: RedisDb,
	}
	eid := "notification-message-insert-10E815D5-6526-4944-ABE5-F666CA2DC037"
	es.AnnounceEvent(eid, map[string]string{"hello": "world"})
}
