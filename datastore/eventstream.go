package datastore

import (
	"encoding/json"
	"errors"
	"gopkg.in/redis.v3"
	"log"
	"strings"
)

// EventStream maintains a redis pubsub connection
// It can be used to subscribe to database change events
// use NewStream(redisClient) to get an initialized EventStream
type EventStream struct {
	Client      *redis.Client
	Pubsub      *redis.PubSub
	subscribers map[string][]chan Event
}

// Event describes a model change even in the database
type Event struct {
	ModelClass string // what kind of object is it?
	Action     string // either insert, update, or delete
	ObjectId   string // the id of the object being modified
	Payload    []byte // JSON data for object
}

func ParseEventId(eventId string, payload string) (ev Event) {
	comps := strings.Split(eventId, "-")
	if len(comps) > 1 {
		ev.ModelClass = comps[1]
	}

	if len(comps) > 2 {
		ev.Action = comps[2]
		idComps := comps[3:]
		ev.ObjectId = strings.Join(idComps, "-")
	}

	ev.Payload = []byte(payload)
	return
}

// NewStream returns an initialized stream that is ready
// to use. All event streams should be created through NewStream.
// The argument should be a verified working redis client.
func NewStream(client *redis.Client) (ev EventStream) {
	ev.Client = client
	ev.subscribers = make(map[string][]chan Event)
	return
}

// AnounceEvent posts a new data store event including the serialized payload.
// "es" must have a working redis client, and it must be possible to serialize
// "payload" using json.Marshal. Will return error if redis publish fails
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

// Function EventChannel returns a channel of events for a given
// subscription query. A wildcard is added to the end of the query
// so you can specify a subscription to only get a certain subcategory
// of events, or leave the argument blank to get all events
func (es *EventStream) EventChannel(subscription string) chan Event {
	ec := make(chan Event, 5)
	rtopic := "notification-" + subscription + "*"

	originalLength := len(es.subscribers[rtopic])
	es.subscribers[rtopic] = append(es.subscribers[rtopic], ec)
	if originalLength == 0 {
		es.FollowPattern(rtopic)
	}
	return ec
}

// FollowPattern subscribes to a given pattern in redis
// If there is no redis connection yet, we create one and
// subscribe to the provided pattern. If a connection is
// already open, we pass the pattern to the background goroutine
// using a channel to have it added to the connection
func (es *EventStream) FollowPattern(pattern string) (err error) {
	if es.Pubsub == nil { // If there's no pubsub connection
		cb := make(chan error)           // let us know if it works
		go es.ListenToRedis(pattern, cb) // and go listen to redis
		if err = <-cb; err != nil {
			return err
		}
	} else {
		// send request to subscribe to the background routine
	}
	return
}

// Function ListenToRedis is intended to run in a goroutine
// and encapsulate all access to the PubSub object, which
// is not thread safe.
func (es *EventStream) ListenToRedis(pattern string, callback chan error) {
	var err error
	es.Pubsub, err = es.Client.PSubscribe(pattern)
	callback <- err
	if err != nil {
		return
	}
	defer es.Close()

	for {
		msg, err := es.Pubsub.ReceiveMessage()
		if err != nil {
			log.Println("Error receiving message:", err)
			continue
		}

		patternSubscribers := es.subscribers[msg.Pattern]
		if len(patternSubscribers) == 0 {
			// There are no subscribers to notify about this event so
			// skip to the next message. Maybe unsubscribe from pattern?
			log.Println("No subscribers for this pattern")
			continue
		}

		for i := range patternSubscribers {
			subscriberChannel := patternSubscribers[i]
			evt := ParseEventId(msg.Channel, msg.Payload)
			select {
			case subscriberChannel <- evt:
			default:
				log.Println("skipped send to subscriber")
			}
		}
	}

	return
}

func (es *EventStream) Close() {
	es.Pubsub.Close()
	es.Pubsub = nil
}
