package main

import (
	"adv-sse/svc"
	"bufio"
	"errors"
	"fmt"
	"log"
	"time"

	fiber "github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

func main() {
	eventService := svc.NewService()

	app := fiber.New()

	app.Post("/:topic", func(c *fiber.Ctx) error {
		params := c.AllParams()
		topic := params["topic"]

		if topic == "" {
			return errors.New("missing topic")
		}

		eventService.Publish(topic, string(c.Body()))
		return c.Status(fiber.StatusNoContent).Send(nil)
	})

	app.Get("/:topic", func(c *fiber.Ctx) error {
		params := c.AllParams() 
		topic := params["topic"]

		if topic == "" {
			return errors.New("missing topic")
		}

		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			sub, err := eventService.Subscribe(topic)
			if err != nil {
				return 
			}
			defer sub.Close()

			var isTimeout bool
			for {
				select {
				case event := <- sub.Listen():
					fmt.Fprintf(w, "id: %d\n", event.Id)
					fmt.Fprintf(w, "event: msg\n")
					fmt.Fprintf(w, "data: %s\n", event.Msg)
				case <-time.After(svc.SubscriptionTimeoutSecs * time.Second):
					fmt.Fprintf(w, "event: timeout\n")
					fmt.Fprintf(w, "data: %ds\n", svc.SubscriptionTimeoutSecs)
					isTimeout = true
				}

				err := w.Flush()
				if err != nil || isTimeout {
					break
				}
			}
		}))

		return nil
	})

	log.Fatal(app.Listen(":8090"))
}
