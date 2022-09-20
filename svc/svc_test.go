package svc

import (
	"fmt"
	"sync"
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
		fmt.Printf("msg %s", event.Msg)
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

func TestSvcMultipleSubs(t *testing.T) {
	service := newService()

	subEvents := make(map[uint64]map[int]string)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		sub, err := service.Subscribe("test topic")
		assert.NoError(t, err)

		subEvents[sub.id] = make(map[int]string)
		wg.Add(1)
		go func() {
			defer wg.Done()
			i := 0
			for {
				select {
				case e := <-sub.Listen():
					subEvents[sub.id][i] = e.Msg
				case <-time.After(time.Second):
					return
				}
				i++
			}
		}()
	}

	service.Publish("test topic", "test1")
	service.Publish("test topic", "test2")
	service.Publish("test topic", "test3")
	service.Publish("test topic", "test4")

	wg.Wait()

	for _, events := range subEvents {
		assert.Equal(t, 4, len(events))
		assert.Equal(t, "test4", events[3])
	}
}
