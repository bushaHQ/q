package q

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

// QueueConfig configuration setting for declaring a queue
type QueueConfig struct {
	Name       string
	Addr       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
	Exchanges  []ExchangeConfig
	NilQueue   bool
}

// Queue represents a connection to a specific queue.
type Queue struct {
	config        QueueConfig
	logger        *log.Logger
	connection    *amqp.Connection
	channel       *amqp.Channel
	done          chan bool
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	isConnected   bool
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second

	// When resending messages the server didn't confirm
	resendDelay = 10 * time.Second
)

var (
	errNotConnected  = errors.New("not connected to the queue")
	errNotConfirmed  = errors.New("message not confirmed")
	errAlreadyClosed = errors.New("already closed: not connected to the queue")
)

// New creates a new queue instance, and automatically
// attempts to connect to the server.
func New(qc QueueConfig) *Queue {
	queue := Queue{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		config: qc,
		done:   make(chan bool),
	}
	go queue.handleReconnect(qc.Addr)
	return &queue
}

// handleReconnect will wait for a connection error on
// notifyClose, and then continously attempt to reconnect.
func (queue *Queue) handleReconnect(addr string) {
	for {
		queue.isConnected = false
		queue.logger.Println("Attempting to connect")
		for !queue.connect(addr) {
			queue.logger.Println("Failed to connect. Retrying...")
			time.Sleep(reconnectDelay)
		}
		select {
		case <-queue.done:
			return
		case <-queue.notifyClose:
		}
	}
}

// connect will make a single attempt to connect to
// RabbitMQ. It returns the success of the attempt.
func (queue *Queue) connect(addr string) bool {
	conn, err := amqp.Dial(addr)
	if err != nil {
		return false
	}
	ch, err := conn.Channel()
	if err != nil {
		return false
	}

	ch.Confirm(false)

	if !queue.config.NilQueue {
		_, err = ch.QueueDeclare(
			queue.config.Name,
			queue.config.Durable,    // Durable
			queue.config.AutoDelete, // Delete when unused
			queue.config.Exclusive,  // Exclusive
			queue.config.NoWait,     // No-wait
			queue.config.Args,       // Arguments
		)
		if err != nil {
			queue.logger.Println(err)
			return false
		}
	}

	// declare exchanegs and bind the queue to them with their bindRoutes value
	err = declareExchanges(ch, queue.config.Name, queue.config.Exchanges)
	if err != nil {
		queue.logger.Println(err)
		return false
	}

	queue.changeConnection(conn, ch)
	queue.isConnected = true
	queue.logger.Println("Connected!")
	return true
}

// changeConnection takes a new connection to the queue,
// and updates the channel listeners to reflect this.
func (queue *Queue) changeConnection(connection *amqp.Connection, channel *amqp.Channel) {
	queue.connection = connection
	queue.channel = channel
	queue.notifyClose = make(chan *amqp.Error)
	queue.notifyConfirm = make(chan amqp.Confirmation)
	queue.channel.NotifyClose(queue.notifyClose)
	queue.channel.NotifyPublish(queue.notifyConfirm)
}

// Connected will return true when if aqueue
// is connected or false if not
func (queue *Queue) Connected() bool {
	return queue.isConnected
}

// Stream will continuously put queue items on the channel.
// It is required to call delivery.Ack when it has been
// successfully processed, or delivery.Nack when it fails.
// Ignoring this will cause data to build up on the server.
func (queue *Queue) Stream() (<-chan amqp.Delivery, error) {
	if !queue.isConnected {
		return nil, errNotConnected
	}
	queue.logger.Println(queue.config.Name)
	return queue.channel.Consume(
		queue.config.Name,
		"",    // Consumer
		false, // Auto-Ack
		false, // Exclusive
		false, // No-local
		false, // No-Wait
		nil,   // Args
	)
}

// Close will cleanly shutdown the channel and connection.
func (queue *Queue) Close() error {
	if !queue.isConnected {
		return errAlreadyClosed
	}
	err := queue.channel.Close()
	if err != nil {
		return err
	}
	err = queue.connection.Close()
	if err != nil {
		return err
	}
	close(queue.done)
	queue.isConnected = false
	return nil
}
