package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril server...")

	// Declare a connection string
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a new channel
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	fmt.Println("Connected to RabbitMQ!")

	// Publish a message to the exchange
	ps := routing.PlayingState{
		IsPaused: true,
	}
	if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, ps); err != nil {
		panic(err)
	}

	// Wait for a signal to exit and if received, close the connection
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel

	fmt.Println("\nGoodbye!")
}
