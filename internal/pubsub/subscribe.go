package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int

const (
	Ack Acktype = iota
	NackRequeue
	NackDiscard
)

func SubscribeJSON[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) Acktype,
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
			switch handler(msg) {
			case Ack:
				del.Ack(false)
			case NackDiscard:
				del.Nack(false, false)
			case NackRequeue:
				del.Nack(false, true)
			}
		}
	}()

	return nil
}

func SubscribeGob[T any](
	conn *amqp.Connection,
	exchange,
	queueName,
	key string,
	simpleQueueType int,
	handler func(T) Acktype,
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
			buffer := bytes.NewBuffer(del.Body)
			decoder := gob.NewDecoder(buffer)
			var msg T
			err := decoder.Decode(&msg)

			if err != nil {
				log.Printf("couldn't decode: %v", err)
				continue
			}
			switch handler(msg) {
			case Ack:
				del.Ack(false)
			case NackDiscard:
				del.Nack(false, false)
			case NackRequeue:
				del.Nack(false, true)
			}
		}
	}()

	return nil
}
