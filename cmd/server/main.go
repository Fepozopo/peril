package main

import (
	"fmt"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
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

	// Declare and bind the queue
	ch, _, err := pubsub.DeclareAndBind(
		conn,                       // conn
		routing.ExchangePerilTopic, // exchange
		"game_logs",                // queueName
		routing.GameLogSlug+".*",   // key
		pubsub.Durable,             // simpleQueueType
	)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ch.Close()

	// Print server help
	gamelogic.PrintServerHelp()

	// Start an infinite loop
	for {

		// Get user input
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		// Check if the user wants to pause the game
		if words[0] == "pause" {
			fmt.Println("Sending pause message...")
			ps := routing.PlayingState{
				IsPaused: true,
			}
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, ps); err != nil {
				panic(err)
			}
		}

		// Check if the user wants to resume the game
		if words[0] == "resume" {
			fmt.Println("Sending resume message...")
			ps := routing.PlayingState{
				IsPaused: false,
			}
			if err := pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, ps); err != nil {
				panic(err)
			}
		}

		// Check if the user wants to quit the game
		if words[0] == "quit" {
			fmt.Println("Exiting...")
			break
		}

		// Check if the user gave an invalid command
		if words[0] != "pause" && words[0] != "resume" && words[0] != "quit" {
			fmt.Printf("Don't understand command: %v\n", words[0])
		}
	}
}
