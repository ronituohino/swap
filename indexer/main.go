package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"indexer/internal/db"
	"indexer/internal/rmq"

	"github.com/gin-gonic/gin"
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

	go func() {
		for d := range rmqClient.Messages {
			var scraped db.ScrapedMessage
			if err := json.Unmarshal([]byte(d.Body), &scraped); err != nil {
				log.Printf("Error parsing message: %v, message: %+v", err, scraped)
				d.Nack(false, false) // Don't requeue due to invalid format
				continue
			}
			log.Printf("Added data to buffer from: %+v", scraped.URL)
			message_buffer = append(message_buffer, scraped)
			if len(message_buffer) < message_buffer_max {
				// Buffer messages in local store
				continue
			}

			// Once buffer full, send buffered messaged to db
			if err := db.InsertScrapedData(database, message_buffer); err != nil {
				log.Printf("Error inserting data: %v", err)
				// Negatively ack all messages in this buffer
				// Request to send to another consumer
				d.Nack(true, true)
			} else {
				log.Printf("Successfully processed message buffer!")
				// Ack all messages in this buffer
				d.Ack(true)
			}

			// Empty buffer
			message_buffer = nil
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
