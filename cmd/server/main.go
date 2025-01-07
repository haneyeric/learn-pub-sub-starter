package main

import (
	"fmt"
	"log"

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
	fmt.Println("Connection successful. Starting Peril server...")
	gamelogic.PrintServerHelp()

	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Connection error: %s", err)
		return
	}

	_, q, err := pubsub.DeclareAndBind(conn, routing.ExchangePerilTopic, routing.GameLogSlug, routing.GameLogSlug+".*", pubsub.DURABLE)
	if err != nil {
		log.Printf("couldn't DeclareAndBind: %v", err)
		return
	}
	fmt.Printf("Queue %v declared and bound.\n", q.Name)

	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}
		switch words[0] {

		case "pause":
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			if err != nil {
				log.Printf("could not publish: %v", err)
				return
			}
			fmt.Println("Pause message sent")

		case "resume":
			err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			if err != nil {
				log.Printf("could not publish: %v", err)
				return
			}
			fmt.Println("Resume message sent")

		case "quit":
			fmt.Println("Exiting")
			return

		default:
			fmt.Printf("I do not understand command: %v\n", words[0])
		}
	}
}
