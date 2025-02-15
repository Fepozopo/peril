package pubsub

import (
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
	handler func(T),
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
				continue
			}
			handler(t)
			msg.Ack(false)
		}
	}()

	return nil
}
