package mq

import (
	"fmt"
	"time"

	"github.com/kvizdos/mq-lib/internal/health"
	"github.com/rabbitmq/amqp091-go"
)

type MQVariables struct {
	Connection   *amqp091.Connection
	Channel      *amqp091.Channel
	Queues       map[string]*amqp091.Queue
	Reconnection chan bool
	CreateQueues DeclareMQQueuesFunc
	Health       health.HealthCheck
	connURI      string
	qos          *QualityOfService
}

func (mq *MQVariables) Listen(options ListenMQOptions) {
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
			if mq.Health != nil {
				mq.Health.SetRabbitStatus(false)
			}
			for {
				// Attempt to reconnect with a delay
				time.Sleep(reconnectDelay)
				rmq, err := InitializeMQ(mq.connURI, mq.CreateQueues, mq.qos, mq.Health)
				if err == nil {
					fmt.Println("Successfully reconnected to RabbitMQ")
					mq.Connection = rmq.Connection
					mq.Channel = rmq.Channel
					mq.Queues = rmq.Queues
					fmt.Println("Sending reconnection notice to MQ consumers..")
					mq.Reconnection <- true
					// health.RabbitUp gets set to TRUE after reconnecting the consumers.
					break
				}
				fmt.Printf("Failed to reconnect to RabbitMQ: %v\n", err)
			}
		}
	}
}
