package pubsub

import (
	"context"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func PublishJSON[T any](ch *amqp.Channel, exchange, key string, val T) error {
	j, err := json.Marshal(val)
	if err != nil {
		log.Fatalf("Error marshalling: %s", err)
		return err
	}

	msg := amqp.Publishing{ContentType: "application/json", Body: j}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	return nil
}
