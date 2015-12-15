package datastore

import (
	"testing"
)

func TestPublishEvent(t *testing.T) {
	es := NewStream(RedisDb)
	eid := "message-insert-10E815D5-6526-4944-ABE5-F666CA2DC037"
	es.AnnounceEvent(eid, map[string]string{"hello": "world"})
}
