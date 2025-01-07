package pubsub

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T),
) error {
	qChan, q, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
	if err != nil {
		log.Printf("couldn't DeclareAndBind: %v", err)
		return err
	}

	delChan, err := qChan.Consume(q.Name, "", false, false, false, false, amqp.Table{})
	if err != nil {
		log.Printf("couldn't create delivery channel: %v", err)
		return err
	}

	go func() {
		defer qChan.Close()
		for del := range delChan {
			var msg T
			err = json.Unmarshal(del.Body, &msg)
			if err != nil {
				log.Printf("couldn't unmarshal: %v", err)
				continue
			}
			handler(msg)
			del.Ack(false)
		}
	}()

	return nil
}
