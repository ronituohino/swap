package main

import (
	"fmt"
	"net/http"
	"os"

	"search-api/internal/db"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	database := db.Initialize()
	defer database.Close()

	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/search", func(c *gin.Context) {
		results := db.Search(database)
		c.JSON(http.StatusOK, results)
	})

	fmt.Printf("Server started in port %v\n", port)
	r.Run(fmt.Sprintf(":%v", port))
}
