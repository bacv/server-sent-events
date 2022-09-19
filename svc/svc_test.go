package svc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSvcSubscription(t *testing.T) {
	service := newService()

	sub, err := service.Subscribe("test")
	assert.NoError(t, err)

	select {
	case event := <-sub.Listen():
		fmt.Printf("msg %s", event.msg)
	case <-time.After(1 * time.Second):
		sub.Close()
	}

	assert.Equal(t, len(service.subscriptions), 1)
}

func TestSvcPublishEvent(t *testing.T) {
	service := newService()

	_, err := service.Publish("topic", "data")
	assert.Error(t, err) // non existent topic
	assert.Equal(t, len(service.subscriptions), 0)

	_, err = service.Subscribe("topic")
	assert.NoError(t, err)

	_, err = service.Publish("topic", "data")
	assert.NoError(t, err)

	assert.Equal(t, len(service.subscriptions), 1)
}

