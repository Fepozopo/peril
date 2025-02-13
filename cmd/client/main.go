package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")

	// Declare a connection string
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create a new channel
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Hello, %s! Let's start the game!\n", username)

	// Call DeclareAndBind
	_, _, err = pubsub.DeclareAndBind(
		conn,                          // conn
		routing.ExchangePerilDirect,   // exchange
		routing.PauseKey+"."+username, // queueName
		routing.PauseKey,              // key
		pubsub.Transient,              // simpleQueueType
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Wait for a signal to exit and if received, close the connection
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	<-signalChannel

	fmt.Println("\nGoodbye!")
}
