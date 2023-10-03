# MQ-Lib (an AMQP wrapper)

This is a pretty high-level AMQP091 wrapper. It primarily helps with **getting things setup**, **insights into service connection health**, and **autonomous reconnections**. This package also helps in reducing the chances of creating *infinite loops*, which if not handled properly, are rather easy to do with the AMQP package by-default.

This is still pretty early, and there are still some features I'd eventually like to add (mainly: reconnect to a different server if/when it drops), however in my testing setup (Docker Swarm), it's already "being handled" (with about 10 million asterisks). I'd also like to build in some deserialization and rate limiting, though for now, that is up to you (the developer) to handle (rate limiting has been tested to be possible with `/x/time/rate`).

## Quick Start Guide

### Install

```
$ go get github.com/Vizdos-Enterprises/mq-lib
```

### Define Queues

This function will be used in the "initializeMQ" function to declare specific queues. It is pretty generic to allow for easy customization and hopefully, less breaking changes to this package if/when the amqp091 Channel changes in the future.

```
(import mq from mq-lib/mq)
func declareQueues(ch *amqp091.Channel) (mq.MQQueues, error) {
	// Setup Example queue
	q, err := ch.QueueDeclare(
		"example", // name
		true,         // durable
		false,        // delete when unused
		false,        // exclusive
		false,        // no-wait
		nil,          // arguments
	)

	if err != nil {
		fmt.Println(err)
		return mq.MQQueues{}, err
	}

	...

	// You'll reuse the key names here later on!
	return mq.MQQueues{
		"example": &q,
		...
	}, nil
}
```

### Connecting to RabbitMQ

**It is recommended to follow the instructions below this, "Connecting to RabbitMQ (WITH Health Checks)" for the reasons listed at the bottom of this README.**

Wherever you initialize your services, place the following code:

```
mqServer := loadServiceURI("MQ", "amqp")
mqVars, err := mq.InitializeMQ(mqServer, declareQueues, nil)
if err != nil {
	panic(fmt.Sprintf("error starting RabbitMQ: %s", err))
}
go mqVars.MonitorConnection()
```

*Be sure to remember the "mqVars.MonitorConnection()" piece, as this is what controls reconnections.*

#### Connecting to RabbitMQ (WITH Health Checks)
**If you have a health check system in place, this package can also help.** Make your your health check structure has the following functions:

```
func (h HealthCheck) GetRabbitStatus() bool {
	return h.RabbitMQUp
}

func (h *HealthCheck) SetRabbitStatus(status bool) {
	h.RabbitMQUp = status
}
```

Now, modify your initialization code to be:

```
health := &HealthCheck{}
mqServer := loadServiceURI("MQ", "amqp")
// Make sure to modify "nil" to "health"
mqVars, err := mq.InitializeMQ(mqServer, declareQueues, health)
if err != nil {
	panic(fmt.Sprintf("error starting RabbitMQ: %s", err))
}
// Set it to UP at the start, as long as there have been no errors.
health.SetRabbitStatus(true)
go mqVars.MonitorConnection()
```

And just like that, you have real-time health checks. This will set the health check to FALSE as soon as the connection drops, and then back to TRUE once reconnection has occured *and* the consumers are live.

### Consume Events

#### Define Consumers

Create a function that defines your channel consumers:

```
func createConsumers(vars *mq.MQVariables) (mq.MQConsumers, error) {
    // Use the keys from the beginning!
	exampleEvents, err := vars.Channel.Consume(
		vars.Queues["example"].Name, // queue
		"",                        // consumer
		true,                      // auto-ack
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)

	if err != nil {
		return nil, err
	}

	...

	return mq.MQConsumers{
		"exampleEvents":     exampleEvents,
		...
	}, nil
}
```

#### Define Event Handling

Create a function that handles incoming events:

```
func handleEvents(events mq.MQConsumers) {
	select {
	case d, ok := <-events["exampleEvents"]:
		if !ok {
			return
		}

		fmt.Println("[x] Received:", string(d.Body))
	}
}
```

#### Put it all together

Then, in you're "DO" area, add in the following code to listen on the MQ:

```
mqVars.Listen(mq.ListenMOptions{
	CreateConsumers: createConsumers,
	HandleEvents:    handleEvents,
})
```

### Final Notes / Pitfalls
- It is **highly recommended** to have a health check system in place. A basic health system isn't too hard to implement, and without it, **there is one major indefinite loop: when you quit the program.** When the program gets quit without a health check in place, the `options.HandleEvents()` inside of `Listen()` will trigger indefinitely until the program has quit (which in testing, still causes a few hundred/thousand executions to occur). While no events will be triggered, the surrounding `for {}` will be triggered many many many times. **The easiest fix to this is implementing your health check, and setting RabbitMQ to DOWN before closing channels.**
