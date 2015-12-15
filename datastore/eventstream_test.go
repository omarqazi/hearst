package datastore

import (
	"encoding/json"
	"testing"
	"time"
)

// Test to see if we can publish an event and then receive it correctly
func TestPublishEvent(t *testing.T) {
	es := NewStream(RedisDb)
	defer es.Close()

	eid := "custom-test-notification-id"
	ch := es.EventChannel("custom-test")
	es.AnnounceEvent(eid, map[string]string{"hello": "world"})

	timeout := time.After(100 * time.Millisecond)
	select {
	case evt := <-ch:
		if evt.ModelClass != "custom" || evt.Action != "test" || evt.ObjectId != "notification-id" {
			t.Error("Expected message insert but got:", evt)
			return
		}

		var decodedObject map[string]string
		if err := json.Unmarshal(evt.Payload, &decodedObject); err != nil {
			t.Error("Error unmarshaling event payload:", err)
			return
		}

		if decodedObject["hello"] != "world" {
			t.Error("Expected payload[hello] == world but got", decodedObject["hello"])
			return
		}
	case <-timeout:
		t.Error("Event channel did not return event within 100ms")
		return
	}
}

//
