package svc

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

const (
	EventChanBound      = 10
	SubscriptionTimeout = 30 * time.Second
)

type Event struct {
	Id  uint64
	Msg string
}

type Subscription struct {
	id      uint64
	eventsC chan Event
	onClose func()
}

func (s *Subscription) Listen() <-chan Event {
	return s.eventsC
}

func (s *Subscription) Publish(id uint64, msg string) {
	s.eventsC <- Event{Id: id, Msg: msg}
}

func (s *Subscription) Close() {
	s.onClose()
	close(s.eventsC)
}

type Topic struct {
	subscriptions map[uint64]*Subscription
	mu            sync.RWMutex
	sub_idx       uint64
}

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

func (t *Topic) Publish(id uint64, msg string) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	for _, sub := range t.subscriptions {
		sub.Publish(id, msg)
	}
}

func (t *Topic) onClose(id uint64) func() {
	return func() {
		t.mu.Lock()
		defer t.mu.Unlock()

		delete(t.subscriptions, id)
	}
}

type EventService interface {
	Subscribe(topic string) (*Subscription, error)
	Publish(topic, message string) (uint64, error)
}

type service struct {
	subscriptions map[string]*Topic
	event_idx     uint64
	mu            sync.RWMutex
}

func NewService() EventService {
	return newService()
}

func newService() *service {
	return &service{
		subscriptions: make(map[string]*Topic),
	}
}

func (s *service) Subscribe(topic string) (*Subscription, error) {
	s.provision(topic)

	sub := s.subscriptions[topic].Subscribe()
	return sub, nil
}

func (s *service) provision(topic string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.subscriptions[topic]; !ok {
		s.subscriptions[topic] = &Topic{
			subscriptions: make(map[uint64]*Subscription),
		}
	}
}

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
