# Q

Q is a package used internally to abstract Rabbitmq queues to simple stream and push op.


## Configs



### QueueConfig

```golang

type QueueConfig struct {
	Name       string // name of the queue
	Addr       string // address of the queue
	Durable    bool 
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
	Exchanges  []ExchangeConfig // list of exchanges
    NilQueue   bool // if true, no queue is created. Only the exchange is created if set.\
    // setting the queue binding will lead to an error if no queue exist
}

```

### ExchangeConfig

```golang

type ExchangeConfig struct {
	Name        string // name of exchange
	Type        string // type of config
	Durable     bool
	AutoDeleted bool
	Internal    bool
	NoWait      bool
	Args        amqp.Table
    BindRoutes  []string // the bind routes for the exchange. this is the routingKey when binding
    // an exchange to a queue
}

```

