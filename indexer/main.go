package main

import (
	"log"
	"time"
	"os"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

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
		log.Printf("Failed to connect to RabbitMQ, attempt %d/%d: %v\n", i + 1, maxRetries, err)
		if i < maxRetries - 1 {
			time.Sleep(2 * time.Second)
		}
	}
	failOnError(err, "Failed to connect to RabbitMQ after multiple attempts")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"scraped_items", // name
		true,   // durable
		false,  // delete when unused
		false,  // exclusive
		false,  // no-wait
		nil,    // arguments
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

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	go func() {
		if err := r.Run(fmt.Sprintf(":%v", port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	var forever chan struct{}

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			time.Sleep(1 * time.Second)
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
