package rmq

import (
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RMQClient struct {
	Connection *amqp.Connection
	Channel    *amqp.Channel
	Queue      amqp.Queue
	Messages   <-chan amqp.Delivery
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func Initialize() *RMQClient {
	rmqHost := os.Getenv("RMQ_HOST")
	rmqPort := os.Getenv("RMQ_PORT")
	rmqUser := os.Getenv("RMQ_USER")
	rmqPassword := os.Getenv("RMQ_PASSWORD")
	maxRetries := 10

	var conn *amqp.Connection
	var err error

	for i := 0; i < maxRetries; i++ {
		conn, err = amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", rmqUser, rmqPassword, rmqHost, rmqPort))
		if err == nil {
			break
		}
		log.Printf("Failed to connect to RabbitMQ, attempt %d/%d: %v\n", i+1, maxRetries, err)
		if i < maxRetries-1 {
			time.Sleep(2 * time.Second)
		}
	}
	failOnError(err, "Failed to connect to RabbitMQ after multiple attempts")

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")

	q, err := ch.QueueDeclare(
		"scraped_items", // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	return &RMQClient{
		Connection: conn,
		Channel:    ch,
		Queue:      q,
		Messages:   msgs,
	}
}

func (c *RMQClient) Close() {
	if c.Channel != nil {
		c.Channel.Close()
	}
	if c.Connection != nil {
		c.Connection.Close()
	}
}
