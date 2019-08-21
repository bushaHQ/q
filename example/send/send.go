package main

import (
	"log"
	"time"

	"github.com/bushaHQ/q"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {

	queue := q.New(q.QueueConfig{
		Name:      "",
		Addr:      "amqp://guest:guest@localhost:5672/",
		Exclusive: true,
		NilQueue:  true,
		Exchanges: []q.ExchangeConfig{
			q.ExchangeConfig{
				Name:    "logging",
				Type:    "fanout",
				Durable: true,
			},
			q.ExchangeConfig{
				Name:    "logs",
				Type:    "fanout",
				Durable: true,
			},
		},
	})
	time.Sleep(time.Second)

	log.Println("here")

	body := "Hello, World!"
	err := queue.Push([]byte(body), q.PushOpt{
		Exchange: "logs",
	})
	failOnError(err, "Failed to publish a message")

	// err = queue.Push([]byte(body+" logging"), q.PushOpt{
	// 	Exchange: "logging",
	// })

	log.Printf(" [x] Sent %s", body)
	failOnError(err, "Failed to publish a message")
}
