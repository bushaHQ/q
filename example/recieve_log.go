package main

import (
	"fmt"
	"log"
	"time"

	"github.com/bushaHQ/q"
)

func main() {

	queue := q.New(q.QueueConfig{
		Name: "test_busha_q",
		Addr: "amqp://guest:guest@localhost:5672/",
		// Exclusive: true,
		AutoDelete: true,
		Exchanges: []q.ExchangeConfig{
			q.ExchangeConfig{
				Name:       "logging",
				Type:       "fanout",
				Durable:    true,
				BindRoutes: []string{""},
			},
			q.ExchangeConfig{
				Name:       "logs",
				Type:       "fanout",
				Durable:    true,
				BindRoutes: []string{""},
			},
		},
	})
	time.Sleep(time.Second)
	msgs, err := queue.Stream()
	failOnError(err, "Failed to stream")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			fmt.Printf(" [x] %s\n", d.Body)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
