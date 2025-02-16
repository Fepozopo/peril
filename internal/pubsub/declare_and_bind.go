package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Define queue types.
const (
	Durable = iota
	Transient
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	// Create a new channel
	ch, err := conn.Channel()
	if err != nil {
		return nil, amqp.Queue{}, err
	}

	// Set queue declare parameters based on simpleQueueType.
	// For a durable queue: durable = true. For a transient queue:
	// autoDelete and exclusive should be true.
	var durable, autoDelete, exclusive bool
	if simpleQueueType == Durable {
		durable = true
		autoDelete = false
		exclusive = false
	} else if simpleQueueType == Transient {
		durable = false
		autoDelete = true
		exclusive = true
	} else {
		ch.Close()
		return nil, amqp.Queue{}, fmt.Errorf("unknown simpleQueueType: %d", simpleQueueType)
	}

	// Declare the queue.
	// noWait is false so we wait for a confirmation.
	queueArgs := amqp.Table{"x-dead-letter-exchange": "peril_dlx"}
	queue, err := ch.QueueDeclare(
		queueName,  // queue name
		durable,    // durable - stays available after broker restart if true
		autoDelete, // autoDelete - deleted when last consumer disconnects if true
		exclusive,  // exclusive - only accessible by the current connection if true
		false,      // noWait - false, wait for a server reply
		queueArgs,  // arguments
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, fmt.Errorf("declaring queue: %v", err)
	}

	// Bind the queue to the exchange.
	err = ch.QueueBind(
		queue.Name, // queue name
		key,        // routing key
		exchange,   // exchange
		false,      // noWait - false, wait for a server reply
		nil,        // arguments
	)
	if err != nil {
		ch.Close()
		return nil, amqp.Queue{}, fmt.Errorf("binding queue: %v", err)
	}

	return ch, queue, nil
}
