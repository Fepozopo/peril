package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AckType int

const (
	Ack AckType = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
	handler func(T) AckType,
) error {
	// Declare and bind the queue
	ch, queue, err := DeclareAndBind(
		conn,            // connection
		exchange,        // exchange
		queueName,       // queue name
		key,             // routing key
		simpleQueueType, // queue type
	)
	if err != nil {
		return fmt.Errorf("failed to declare and bind queue: %w", err)
	}

	// Limit the prefetch count to 10
	if err := ch.Qos(10, 0, false); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	// Get a new chan of amqp.Delivery structs
	msgs, err := ch.Consume(
		queue.Name, // queue name
		"",         // consumer name - auto-generated
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to consume from queue: %w", err)
	}

	// Start a goroutine to handle messages
	go func() {
		for msg := range msgs {
			var t T
			if err := json.Unmarshal(msg.Body, &t); err != nil {
				fmt.Printf("failed to unmarshal message: %v\n", err)
				msg.Nack(false, true)
				fmt.Println("NackRequeue due to unmarshal error")
				continue
			}
			switch handler(t) {
			case Ack:
				msg.Ack(false)
			case NackRequeue:
				msg.Nack(false, true)
			case NackDiscard:
				msg.Nack(false, false)
			}
		}
	}()

	return nil
}
