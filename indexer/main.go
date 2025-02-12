package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

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

	go func() {
		for d := range rmqClient.Messages {
			var message db.ScrapedMessage
			if err := json.Unmarshal([]byte(d.Body), &message); err != nil {
				log.Printf("Error parsing message: %v, message: %+v", err, message)
				d.Nack(false, false) // Don't requeue due to invalid format
				continue
			}

			if err := db.InsertScrapedData(database, message); err != nil {
				log.Printf("Error inserting data: %v, message: %+v", err, message)
				d.Nack(false, true)
				continue
			}

			log.Printf("Successfully processed message for URL: %s", message.URL)
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
