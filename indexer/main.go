package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"indexer/internal/db"
	"indexer/internal/rmq"

	"github.com/gin-gonic/gin"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	database := db.Initialize()
	defer database.Close()

	rmqClient := rmq.Initialize()
	defer rmqClient.Close()

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	go func() {
		if err := r.Run(fmt.Sprintf(":%v", port)); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	forever := make(chan struct{})
	message_buffer_max, err := strconv.Atoi(os.Getenv("MESSAGE_BUFFER_MAX"))
	if err != nil {
		panic(err)
	}
	message_buffer := []db.ScrapedMessage{}
	flushTimeout := time.NewTimer(30 * time.Second)
	var lastDelivery *amqp.Delivery

	go func() {
		for {
			select {
			case d, ok := <-rmqClient.Messages:
				if !ok {
					return
				}
				lastDelivery = &d
				var scraped db.ScrapedMessage
				if err := json.Unmarshal([]byte(d.Body), &scraped); err != nil {
					log.Printf("Error parsing message: %v, message: %+v", err, scraped)
					d.Nack(false, false) // Don't requeue due to invalid format
					continue
				}
				message_buffer = append(message_buffer, scraped)
				if len(message_buffer) < message_buffer_max {
					flushTimeout.Reset(5 * time.Second)
					continue
				}

				if err := db.InsertScrapedData(database, message_buffer); err != nil {
					log.Printf("Error inserting data: %v", err)
					d.Nack(true, true)
				} else {
					log.Printf("Successfully processed message buffer!")
					d.Ack(true)
				}
				message_buffer = nil

			case <-flushTimeout.C:
				if len(message_buffer) > 0 && lastDelivery != nil {
					if err := db.InsertScrapedData(database, message_buffer); err != nil {
						log.Printf("Error inserting data on timeout: %v", err)
						lastDelivery.Nack(true, true)
					} else {
						log.Printf("Successfully processed message buffer on timeout!")
						lastDelivery.Ack(true)
					}
					message_buffer = nil
				}
				flushTimeout.Reset(5 * time.Second)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
