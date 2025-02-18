package main

import (
	"fmt"

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
	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}
	defer ch.Close()

	fmt.Println("Connected to RabbitMQ!")

	// Welcome the user
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Printf("Hello, %s! Let's start the game!\n", username)

	// Create a new game state
	gameState := gamelogic.NewGameState(username)

	// Subscribe to pauses
	err = pubsub.SubscribeJSON(
		conn,                          // conn
		routing.ExchangePerilDirect,   // exchange
		routing.PauseKey+"."+username, // queueName
		routing.PauseKey,              // key
		pubsub.Transient,              // simpleQueueType
		handlerPause(gameState),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Subscribe to moves from other players
	err = pubsub.SubscribeJSON(
		conn,                                 // conn
		routing.ExchangePerilTopic,           // exchange
		routing.ArmyMovesPrefix+"."+username, // queueName
		routing.ArmyMovesPrefix+".*",         // key
		pubsub.Transient,                     // simpleQueueType
		handlerMove(gameState, ch),
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Subscribe to wars
	err = pubsub.SubscribeJSON(
		conn,                               // conn
		routing.ExchangePerilTopic,         // exchange
		"war",                              // queueName
		routing.WarRecognitionsPrefix+".*", // key
		pubsub.Durable,                     // simpleQueueType
		handlerWar(gameState, ch),          // handler
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Start an infinite loop
	for {

		// Get user input
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		// Handle spawn
		if words[0] == "spawn" {
			if err := gameState.CommandSpawn(words); err != nil {
				fmt.Println(err)
			}
		}

		// Handle move
		if words[0] == "move" {
			move, err := gameState.CommandMove(words)
			if err != nil {
				fmt.Println(err)
			} else {
				err = pubsub.PublishJSON(
					ch,                                   // channel
					routing.ExchangePerilTopic,           // exchange
					routing.ArmyMovesPrefix+"."+username, // routing key
					move,
				)
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println("Move published successfully.")
			}
		}

		// Handle status
		if words[0] == "status" {
			gameState.CommandStatus()
		}

		// Handle help
		if words[0] == "help" {
			gamelogic.PrintClientHelp()
		}

		// Hndle spam
		if words[0] == "spam" {
			fmt.Println("Spamming not allowed yet!")
		}

		// Handle quit
		if words[0] == "quit" {
			gamelogic.PrintQuit()
			return
		}

		// Handle unknown
		if words[0] != "spawn" && words[0] != "move" && words[0] != "status" && words[0] != "help" && words[0] != "spam" && words[0] != "quit" {
			fmt.Printf("Don't understand command: %v\n", words[0])
		}
	}
}
