package main

import (
	"adv-sse/svc"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TODO: use custom data store to inspect event service behaviour.
type mockEventService struct {
	eventService svc.EventService
}

func (m *mockEventService) Subscribe(topic string) (*svc.Subscription, error) {
	return m.eventService.Subscribe(topic)
}

func (m *mockEventService) Publish(topic, message string) (uint64, error) {
	return m.eventService.Publish(topic, message)
}

func TestPubSubResponses(t *testing.T) {
	eventService := mockEventService{
		eventService: svc.NewService(),
	}

	app, err := setup(&eventService, 10*time.Millisecond)
	assert.NoError(t, err)

	req := httptest.NewRequest("GET", "/testtopic", nil)

	res, err:=app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)

	req = httptest.NewRequest("POST", "/testtopic", strings.NewReader("test"))
	res, err = app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 204, res.StatusCode)
}

