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

// Test to make sure we don't get any messages we didn't ask for
func TestEventQualifier(t *testing.T) {
	es := NewStream(RedisDb)
	defer es.Close()

	ch := es.EventChannel("event-im-not-going-to-get")
	es.AnnounceEvent("some-other-event", map[string]string{"hello": "world"})

	select {
	case impossible := <-ch:
		t.Error("Got event that pattern should not have matched:", impossible)
		return
	default:
	}
}

// Test to make sure two different patterns on one connection works
func TestMultiplePatterns(t *testing.T) {
	es := NewStream(RedisDb)
	defer es.Close()

	if es.Pubsub != nil {
		t.Error("Expected nil Pubsub but found", es.Pubsub)
		return
	}

	cha := es.EventChannel("channel-a")
	chb := es.EventChannel("channel-b")
	chAll := es.EventChannel("channel")

	if es.Pubsub == nil {
		t.Error("Expected Pubsub but found nil")
		return
	}

	es.AnnounceEvent("channel-a-bob", map[string]string{})
	es.AnnounceEvent("channel-b-steve", map[string]string{})
	es.AnnounceEvent("channel-a-mark", map[string]string{})

	timeout := time.After(100 * time.Millisecond)
	select {
	case <-cha:
	case <-timeout:
		t.Error("no notification received for channel-a")
		return
	}

	timeout = time.After(100 * time.Millisecond)
	select {
	case <-chb:
	case <-timeout:
		t.Error("no notification received for channel-b")
		return
	}

	timeout = time.After(100 * time.Millisecond)
	select {
	case <-cha:
	case <-timeout:
		t.Error("no second notification received for channel-a")
		return
	}

	timeout = time.After(100 * time.Millisecond)
	for i := 0; i < 3; i++ {
		select {
		case <-chAll:
		case <-timeout:
			t.Error("not enough notifications received on all channel")
			return
		}
	}
}

// Test to make sure we automatically unsubscribe / disconnect
func TestAutoRemove(t *testing.T) {
	es := NewStream(RedisDb)
	defer es.Close()

	if es.Pubsub != nil {
		t.Error("Expected nil Pubsub but found", es.Pubsub)
		return
	}

	cha := es.EventChannel("auto-remove-test")

	if es.Pubsub == nil {
		t.Error("Expected Pubsub but found nil")
		return
	}

	es.AnnounceEvent("auto-remove-test-event", map[string]string{})
	timeout := time.After(100 * time.Millisecond)
	select {
	case evt := <-cha:
		if evt.ModelClass != "auto" || evt.Action != "remove" {
			t.Error("expected auto remove event but got", evt)
			return
		}
	case <-timeout:
		t.Error("never received auto remove event")
		return
	}

	cha = nil

	for i := 0; i < 100; i++ {
		es.AnnounceEvent("auto-remove-test-event", map[string]string{})
	}

	if es.Pubsub != nil {
		t.Error("Expected Pubsub to be nil but found", es.Pubsub)
		return
	}

}
