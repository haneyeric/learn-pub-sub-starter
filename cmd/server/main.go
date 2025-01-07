package main

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

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
	ch, err := conn.Channel()
	if err != nil {
		fmt.Printf("Connection error: %s", err)
		return
	}
	err = pubsub.PublishJSON(ch, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
	if err != nil {
		log.Printf("could not publish: %v", err)
		return
	}
}
