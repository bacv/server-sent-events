package svc

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// EventChanBound is used to create a non blokcing channel for a subscirber.
	EventChanBound = 10

	// SubscriptionTimeout is a helper constant that can be used when creating a new event service.
	SubscriptionTimeout = 30 * time.Second
)

// Event represents a newly created event when data is being published to the service.
type Event struct {
	Id  uint64
	Msg string
}

// Subscription holds all required data for informing the subscirber about new events.
type Subscription struct {
	id      uint64
	eventsC chan Event
	onClose func()
}

// Listen returns a readonly channel for getting new events.
func (s *Subscription) Listen() <-chan Event {
	return s.eventsC
}

// Publish should be used by the topic to inform a subscriber about a new event.
func (s *Subscription) Publish(id uint64, msg string) {
	s.eventsC <- Event{Id: id, Msg: msg}
}

// Close calls a callback that was set during the subscription initialization and closes the event channel.
func (s *Subscription) Close() {
	s.onClose()
	close(s.eventsC)
}

// Topic holds all the data about the topic and it's subscribers.
type Topic struct {
	subscriptions map[uint64]*Subscription
	mu            sync.RWMutex
	sub_idx       uint64
}

// Subscribe increments the internal count of subscribers for a given topic and return a 
// Subscription data structure for the caller.
func (t *Topic) Subscribe() *Subscription {
	t.mu.Lock()
	defer t.mu.Unlock()

	atomic.AddUint64(&t.sub_idx, 1)
	id := atomic.LoadUint64(&t.sub_idx)

	sub := &Subscription{
		id:      id,
		eventsC: make(chan Event, EventChanBound),
		onClose: t.onClose(id),
	}

	t.subscriptions[id] = sub
	return sub
}

// Publish creates a new event and informs all the subscribers for a given topic about it.
func (t *Topic) Publish(id uint64, msg string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, sub := range t.subscriptions {
		sub.Publish(id, msg)
	}
}

// onClose is a helper function that creates a closure for subscription data structures to ease the 
// cleanup when subscription is discarded.
func (t *Topic) onClose(id uint64) func() {
	return func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		delete(t.subscriptions, id)
	}
}

// EventService defines methods for subscribing and publishing to a topic.
type EventService interface {
	Subscribe(topic string) (*Subscription, error)
	Publish(topic, message string) (uint64, error)
}

// service implements EventService interface and holds data about topics and it subscribers.
type service struct {
	subscriptions map[string]*Topic
	event_idx     uint64
	mu            sync.RWMutex
}

// NewService ...
func NewService() EventService {
	return newService()
}

// newService is a helper method for creating an event service, it should only be used in this package.
func newService() *service {
	return &service{
		subscriptions: make(map[string]*Topic),
	}
}

// Subscribe method creates a new subscription for a given topic.
func (s *service) Subscribe(topic string) (*Subscription, error) {
	s.provision(topic)

	sub := s.subscriptions[topic].Subscribe()
	return sub, nil
}

// Publish creates a new event for a given topic and informs all the subscribers about it.
func (s *service) Publish(topic, message string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// If nobody's subscribed to the channel, ignore the message.
	if _, ok := s.subscriptions[topic]; !ok {
		return 0, errors.New("non existent topic")
	}

	atomic.AddUint64(&s.event_idx, 1)
	id := atomic.LoadUint64(&s.event_idx)

	s.subscriptions[topic].Publish(id, message)

	return id, nil
}

// provision is a helper method to create a map entry for a topic if it doesn't exist.
func (s *service) provision(topic string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscriptions[topic]; !ok {
		s.subscriptions[topic] = &Topic{
			subscriptions: make(map[uint64]*Subscription),
		}
	}
}

