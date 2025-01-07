package pubsub

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
)

const (
	TRANSIENT = iota
	DURABLE
)

func DeclareAndBind(
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int, // an enum to represent "durable" or "transient"
) (*amqp.Channel, amqp.Queue, error) {
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("couldn't create channel: %v", err)
		return nil, amqp.Queue{}, err
	}
	q, err := ch.QueueDeclare(queueName, simpleQueueType == DURABLE, simpleQueueType == TRANSIENT, simpleQueueType == TRANSIENT, false, amqp.Table{"x-dead-letter-exchange": routing.ExchangePerilDeadLetter})
	if err != nil {
		log.Printf("couldn't create queue: %v", err)
		return nil, amqp.Queue{}, err
	}
	err = ch.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		log.Printf("couldn't bind queue: %v", err)
		return nil, amqp.Queue{}, err
	}

	return ch, q, nil
}
