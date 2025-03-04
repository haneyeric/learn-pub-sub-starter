package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

func main() {
	connAddr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connAddr)
	if err != nil {
		fmt.Printf("Connection error: %s", err)
		return
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Connection error: %s", err)
		return
	}
	defer ch.Close()
	fmt.Println("Starting Peril client...")
	username, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Printf("couldn't get username: %v", err)
		return
	}

	gamestate := gamelogic.NewGameState(username)

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilDirect, routing.PauseKey+"."+gamestate.GetUsername(), routing.PauseKey, pubsub.TRANSIENT, handlerPause(gamestate))
	if err != nil {
		log.Printf("couldn't subscribe to pause: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(conn, routing.ExchangePerilTopic, routing.WarRecognitionsPrefix, routing.WarRecognitionsPrefix+".*", pubsub.DURABLE, handlerWar(gamestate, ch))
	if err != nil {
		log.Printf("couldn't subscribe to declarewar: %v", err)
		return
	}

	err = pubsub.SubscribeJSON(conn, string(routing.ExchangePerilTopic), routing.ArmyMovesPrefix+"."+gamestate.GetUsername(), routing.ArmyMovesPrefix+".*", pubsub.TRANSIENT, handlerMove(gamestate, ch))
	if err != nil {
		log.Printf("couldn't subscribe to armymoves: %v", err)
		return
	}

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {

		case "spawn":
			err = gamestate.CommandSpawn(words)
			if err != nil {
				log.Printf("could not spawn: %v", err)
			}

		case "move":
			mv, err := gamestate.CommandMove(words)
			if err != nil {
				fmt.Println(err)
				continue
			}
			err = pubsub.PublishJSON(ch, routing.ExchangePerilTopic, routing.ArmyMovesPrefix+"."+mv.Player.Username, mv)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Printf("Moved %v units to %s\n", len(mv.Units), mv.ToLocation)

			gamestate.CommandStatus()

		case "help":
			gamelogic.PrintClientHelp()

		case "spam":
			if len(words) < 2 {
				fmt.Println("Spamming command incomplete")
				break
			}
			n, err := strconv.Atoi(words[1])
			if err != nil {
				fmt.Println("Spamming command incorrect")
				break
			}
			for i := 0; i < n; i++ {
				msg := gamelogic.GetMaliciousLog()
				log := routing.GameLog{CurrentTime: time.Now(), Username: gamestate.Player.Username, Message: msg}
				err = pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+log.Username, log)
				if err != nil {
					fmt.Printf("Spamming error: %s\n", err)
				}
			}

		case "quit":
			gamelogic.PrintQuit()
			return

		default:
			fmt.Println("unknown command")
		}
	}
}

func handlerPause(gs *gamelogic.GameState) func(routing.PlayingState) pubsub.Acktype {

	return func(ps routing.PlayingState) pubsub.Acktype {
		defer fmt.Print("> ")
		gs.HandlePause(ps)
		return pubsub.Ack
	}
}

func handlerMove(gs *gamelogic.GameState, ch *amqp.Channel) func(gamelogic.ArmyMove) pubsub.Acktype {

	return func(mv gamelogic.ArmyMove) pubsub.Acktype {
		defer fmt.Print("> ")
		mo := gs.HandleMove(mv)
		switch mo {
		case gamelogic.MoveOutcomeSamePlayer:
			return pubsub.Ack
		case gamelogic.MoveOutComeSafe:
			return pubsub.Ack
		case gamelogic.MoveOutcomeMakeWar:
			err := pubsub.PublishJSON(
				ch,
				routing.ExchangePerilTopic,
				routing.WarRecognitionsPrefix+"."+gs.GetUsername(),
				gamelogic.RecognitionOfWar{
					Attacker: mv.Player,
					Defender: gs.GetPlayerSnap(),
				},
			)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				return pubsub.NackRequeue
			}
			return pubsub.Ack
		}
		fmt.Println("bad move outcome")
		return pubsub.NackDiscard
	}
}

func handlerWar(gs *gamelogic.GameState, ch *amqp.Channel) func(dw gamelogic.RecognitionOfWar) pubsub.Acktype {
	return func(dw gamelogic.RecognitionOfWar) pubsub.Acktype {
		defer fmt.Print("> ")
		warOutcome, w, l := gs.HandleWar(dw)

		gl := routing.GameLog{CurrentTime: time.Now(), Username: gs.Player.Username}

		switch warOutcome {
		case gamelogic.WarOutcomeNotInvolved:
			return pubsub.NackRequeue
		case gamelogic.WarOutcomeNoUnits:
			return pubsub.NackDiscard
		case gamelogic.WarOutcomeOpponentWon:
			gl.Message = fmt.Sprintf("%s won a war against %s", w, l)
		case gamelogic.WarOutcomeYouWon:
			gl.Message = fmt.Sprintf("%s won a war against %s", w, l)
		case gamelogic.WarOutcomeDraw:
			gl.Message = fmt.Sprintf("A war between %s and %s resulted in a draw", w, l)
		default:
			fmt.Println("error: unknown war outcome")
			return pubsub.NackDiscard
		}

		err := pubsub.PublishGob(ch, routing.ExchangePerilTopic, routing.GameLogSlug+"."+dw.Attacker.Username, gl)
		if err != nil {
			fmt.Printf("error: couldn't publish war result: %s\n", err)
			return pubsub.NackRequeue
		}

		return pubsub.Ack

	}
}
