package mq

import (
	"fmt"
	"time"

	"github.com/Vizdos-Enterprises/mq-lib/health"
	"github.com/rabbitmq/amqp091-go"
)

type MQVariables struct {
	Connection   *amqp091.Connection
	Channel      *amqp091.Channel
	Queues       map[string]*amqp091.Queue
	Reconnection chan bool
	CreateQueues DeclareMQQueuesFunc
	health       health.HealthCheck
	connURI      string
	qos          *QualityOfService
	isConsuming  bool
}

func (mq *MQVariables) Listen(options ListenMQOptions) {
	mq.isConsuming = true
	listenOnMQ(mq, options)
}

func (mq *MQVariables) MonitorConnection() {
	reconnectDelay := 5 * time.Second
	for {
		// Listen to the NotifyClose channel
		closeCh := make(chan *amqp091.Error)
		mq.Connection.NotifyClose(closeCh)

		// Wait for the connection to close
		err := <-closeCh
		if err != nil {
			fmt.Printf("Connection to RabbitMQ lost: %v. Reconnecting...\n", err)
			if mq.health != nil {
				mq.health.SetRabbitStatus(false)
			}
			for {
				// Attempt to reconnect with a delay
				time.Sleep(reconnectDelay)
				rmq, err := InitializeMQ(mq.connURI, mq.CreateQueues, mq.qos, mq.health)
				if err == nil {
					fmt.Println("Successfully reconnected to RabbitMQ")
					mq.Connection = rmq.Connection
					mq.Channel = rmq.Channel
					mq.Queues = rmq.Queues
					if mq.isConsuming {
						fmt.Println("Sending reconnection notice to MQ consumers..")
						mq.Reconnection <- true
					} else {
						mq.health.SetRabbitStatus(true)
					}
					// health.RabbitUp gets set to TRUE after reconnecting the consumers.
					break
				}
				fmt.Printf("Failed to reconnect to RabbitMQ: %v\n", err)
			}
		}
	}
}
