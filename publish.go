package q

import (
	"errors"
	"time"

	"github.com/streadway/amqp"
)

// PushOpt options for publishing to a queue
type PushOpt struct {
	Exchange   string
	RoutingKey string
	Mandatory  bool
	Immediate  bool
}

// Push will push data onto the queue, and wait for a confirm.
// If no confirms are recieved until within the resendTimeout,
// it continuously resends messages until a confirm is recieved.
// This will block until the server sends a confirm. Errors are
// only returned if the push action itself fails, see UnsafePush.
func (queue *Queue) Push(data []byte, pushOpt ...PushOpt) error {
	if !queue.isConnected {
		return errors.New("failed to push push: not connected")
	}
	for {
		err := queue.UnsafePush(data, pushOpt...)
		if err != nil {
			queue.logger.Println("Push failed. Retrying...")
			continue
		}
		select {
		case confirm := <-queue.notifyConfirm:
			if confirm.Ack {
				queue.logger.Println("Push confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}
		queue.logger.Println("Push didn't confirm. Retrying...")
	}
}

// UnsafePush will push to the queue without checking for
// confirmation. It returns an error if it fails to connect.
// No guarantees are provided for whether the server will
// recieve the message.
func (queue *Queue) UnsafePush(data []byte, pushOpt ...PushOpt) error {

	if !queue.isConnected {
		return errNotConnected
	}
	po := PushOpt{
		RoutingKey: queue.config.Name,
	}
	if len(pushOpt) == 1 {
		po = pushOpt[0]
	}

	return queue.channel.Publish(
		po.Exchange,   // Exchange
		po.RoutingKey, // Routing key
		po.Mandatory,  // Mandatory
		po.Immediate,  // Immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)
}
