package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"search-api/internal/db"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	database := db.Initialize()
	defer database.Close()

	// Read in lemmatize.json
	lem_content, err := os.ReadFile("./lemmatize.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var lemmatize map[string]string
	err = json.Unmarshal(lem_content, &lemmatize)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	// Read in transforms.json
	tra_content, err := os.ReadFile("./transforms.json")
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}
	var transforms map[string]string
	err = json.Unmarshal(tra_content, &transforms)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET"},
		AllowHeaders:     []string{"Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "accept", "origin", "Cache-Control", "X-Requested-With"},
		AllowCredentials: true,
	}))

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/search", func(c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
			return
		}

		results, err := db.Search(database, query, lemmatize, transforms)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, results)
	})

	fmt.Printf("Server started in port %v\n", port)
	r.Run(fmt.Sprintf(":%v", port))
}
