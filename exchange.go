package q

import (
	"log"

	"github.com/streadway/amqp"
)

// ExchangeConfig configuration settings when decraring an exchange
type ExchangeConfig struct {
	Name        string
	Type        string
	Durable     bool
	AutoDeleted bool
	Internal    bool
	NoWait      bool
	Args        amqp.Table
	BindRoutes  []string
}

// declareExchanges  declares exchanges connected to a queue
func declareExchanges(ch *amqp.Channel, queueName string, ee []ExchangeConfig) error {
	for _, ec := range ee {
		err := ch.ExchangeDeclare(
			ec.Name,
			ec.Type,
			ec.Durable,
			ec.AutoDeleted,
			ec.Internal,
			ec.NoWait,
			ec.Args,
		)
		if err != nil {
			return err
		}

		err = bindQueTopics(ch, queueName, ec)
		if err != nil {
			return err
		}
	}
	return nil
}

// bindQueTopics binds a list of bind route on exchangeConfig to a queue named queueName
func bindQueTopics(ch *amqp.Channel, queueName string, ec ExchangeConfig) error {

	for _, br := range ec.BindRoutes {
		log.Println("bind", queueName, ec.Name)
		err := ch.QueueBind(
			queueName,
			br,
			ec.Name,
			false,
			nil,
		)

		if err != nil {
			return err
		}
	}

	return nil
}
