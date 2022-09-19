package svc

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSvcSubscription(t *testing.T) {
	service := NewService()
	go service.Spawn()

	sub, err := service.Subscribe("test")
	assert.NoError(t, err)

	select {
	case msg := <-sub.Listen():
		fmt.Printf("msg %s\n", msg)
	case <-time.After(1 * time.Second):
		sub.Close()
	}

	//assert.Equal(t, len(service.topics), 1)
}

func TestSvcPublishEvent(t *testing.T) {
	service := NewService()
	go service.Spawn()

	//id := service.Publish("topic", "data")
	//assert.Equal(t, len(service.topics), 1)
}

