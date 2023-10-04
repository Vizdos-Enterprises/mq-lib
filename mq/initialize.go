package mq

import (
	"fmt"
	"time"

	"github.com/Vizdos-Enterprises/mq-lib/health"
	"github.com/rabbitmq/amqp091-go"
)

type MQQueues map[string]*amqp091.Queue
type DeclareMQQueuesFunc func(*amqp091.Channel) (MQQueues, error)

func InitializeMQ(connURI string, declareQueues DeclareMQQueuesFunc, qos *QualityOfService, health health.HealthCheck) (*MQVariables, error) {
	conn, err := amqp091.Dial(connURI)
	if err != nil {
		fmt.Println("Failed to connect to MQ. Trying again in 5 seconds..", err)
		time.Sleep(5 * time.Second)
		return InitializeMQ(connURI, declareQueues, qos, health)
	}

	ch, err := conn.Channel()
	if err != nil {
		return &MQVariables{}, fmt.Errorf("failed to open a channel")
	}

	if qos == nil {
		qos = &QualityOfService{
			PrefetchCount: 0,
			PrefetchSize:  0,
			Global:        false,
		}
	}

	err = ch.Qos(qos.PrefetchCount, qos.PrefetchSize, qos.Global)
	if err != nil {
		return &MQVariables{}, fmt.Errorf("failed to configure QoS")
	}

	queues, err := declareQueues(ch)

	if err != nil {
		return &MQVariables{}, err
	}

	return &MQVariables{
		Connection:   conn,
		Channel:      ch,
		Queues:       queues,
		CreateQueues: declareQueues,
		Reconnection: make(chan bool),
		health:       health,
		connURI:      connURI,
		qos:          qos,
	}, nil
}
