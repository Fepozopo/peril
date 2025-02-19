package main

import (
	"fmt"
	"time"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.AckType {
	return func(ps routing.PlayingState) pubsub.AckType {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.AckType {
	return func(move gamelogic.ArmyMove) pubsub.AckType {
		defer fmt.Print("> ")
		outcome := gs.HandleMove(move)
		switch outcome {
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			defender := gs.GetPlayerSnap() // The current player is being attacked
			attacker := move.Player        // The player from the move is attacking

			// Only publish war if attacker and defender are different players
			if attacker.Username != defender.Username {
				warRec := gamelogic.RecognitionOfWar{
					Attacker: attacker,
					Defender: defender,
				}
				key := fmt.Sprintf("%s.%s", routing.WarRecognitionsPrefix, attacker.Username)
				err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, key, warRec)
				if err != nil {
					return pubsub.NackRequeue
				}
			}
			return pubsub.Ack
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.NackDiscard
		default:
			return pubsub.NackDiscard
		}
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.RecognitionOfWar) pubsub.AckType {
	return func(rw gamelogic.RecognitionOfWar) pubsub.AckType {
		defer fmt.Print("> ")

		outcome, winner, loser := gs.HandleWar(rw)
		var message string

		switch outcome {
		case gamelogic.WarOutcomeOpponentWon, gamelogic.WarOutcomeYouWon:
			message = fmt.Sprintf("%s won a war against %s", winner, loser)
		case gamelogic.WarOutcomeDraw:
			message = fmt.Sprintf("A war between %s and %s resulted in a draw", winner, loser)
		default:
			return pubsub.NackDiscard
		}

		gameLog := routing.GameLog{
			CurrentTime: time.Now(),
			Message:     message,
			Username:    rw.Attacker.Username,
		}
		err := pubsub.PublishJSON(ch, routing.ExchangePerilTopic, fmt.Sprintf("%s.%s", routing.GameLogSlug, gameLog.Username), &gameLog)
		if err != nil {
			return pubsub.NackRequeue
		}
		return pubsub.Ack
	}
}
