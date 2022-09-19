package svc

type Subscription struct {
	id uint64
	eventsC chan string
}

func (s *Subscription) Listen() <-chan string {
	return s.eventsC
}

func (s *Subscription) Close() {
	// TODO
	close(s.eventsC)
}

type EventService interface {
	Subscribe(topic string) (*Subscription, error)
	Publish(topic, message string) uint64
	Spawn()
}

type service struct {
	subscriptions map[string]map[uint64]*Subscription
}

func NewService() EventService {
	return &service{
		subscriptions: make(map[string]map[uint64]*Subscription),
	}
}

func (s *service) Subscribe(topic string) (*Subscription, error) {
	s.provision(topic)

	id := uint64(0)
	sub := &Subscription{
		id: id,
		eventsC: make(chan string),
	}

	s.subscriptions[topic][id] = sub
	return sub, nil
}

func (s *service) provision(topic string) {
	if _, ok := s.subscriptions[topic]; !ok {
		s.subscriptions[topic] = make(map[uint64]*Subscription)
	}
}

func (s *service) Publish(topic, message string) uint64 {
	return 0
}

func (s *service) Spawn() {}
