package mq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type MQConsumers map[string]<-chan amqp091.Delivery

type MQConsumersFunc func(vars *MQVariables) (MQConsumers, error)
type MQHandleEventsFunc func(events MQConsumers)

type ListenMQOptions struct {
	CreateConsumers MQConsumersFunc    // Required
	HandleEvents    MQHandleEventsFunc // Required
}

func listenOnMQ(vars *MQVariables, options ListenMQOptions) error {
	events, err := options.CreateConsumers(vars)

	if err != nil {
		return err
	}

	updatedEvents := make(chan MQConsumers)
	didReconnect := make(chan bool)

	// Handle reconnections here
	go func() {
		for {
			select {
			case <-vars.Reconnection:
				fmt.Println("Reconnected to RabbitMQ, reconfiguring consumers..")
				newEvents, err := options.CreateConsumers(vars)

				if err != nil {
					fmt.Println("error setting up consumer after reconnection:", err)
					break
				}

				fmt.Printf("RabbitMQ reconnected and cosumers setup.\n")
				if vars.Health != nil {
					vars.Health.SetRabbitStatus(true)
				}

				didReconnect <- true
				updatedEvents <- newEvents
			}
		}
	}()

	// Handle Events
	go func() {
		currentEvents := events
		for {
			select {
			case newEvents := <-updatedEvents:
				// Update the current events with the new ones
				fmt.Println("Updating new points for the MQ..")
				currentEvents = newEvents
			default:
				if vars.Health != nil && !vars.Health.GetRabbitStatus() {
					// don't cause an infinite loop if rmq is down.
					<-didReconnect
				}
				options.HandleEvents(currentEvents)
			}
		}
	}()

	return nil
}
