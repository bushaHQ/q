package q_test

import (
	"log"

	"github.com/bushaHQ/q"
)

func ExamplePush() {
	queue := q.New(q.QueueConfig{
		Name:       "test",
		Addr:       "amqp://guest:guest@localhost:5672/",
		AutoDelete: true,
		NilQueue:   true,
	})
	err := queue.Push([]byte("busha test"))
	if err != nil {
		log.Println(err)
	}
}
