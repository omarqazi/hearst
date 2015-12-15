package datastore

import (
	"encoding/json"
	"errors"
	"gopkg.in/redis.v3"
)

type EventStream struct {
	Client *redis.Client
	Pubsub *redis.PubSub
}

func (es *EventStream) AnnounceEvent(eventId string, payload interface{}) error {
	if es.Client == nil {
		return errors.New("EventStream redis client = nil")
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	rtopic := "notification-" + eventId
	err = es.Client.Publish(rtopic, string(jsonPayload)).Err()
	return err
}
