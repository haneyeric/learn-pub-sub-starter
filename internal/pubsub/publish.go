package pubsub

import (
	"bytes"
	"context"
	"encoding/gob"
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

func PublishGob[T any](ch *amqp.Channel, exchange, key string, val T) error {

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(val)
	if err != nil {
		log.Fatalf("Error encoding to gob: %s", err)
		return err
	}

	msg := amqp.Publishing{ContentType: "application/gob", Body: buf.Bytes()}

	ch.PublishWithContext(context.Background(), exchange, key, false, false, msg)
	return nil
}
