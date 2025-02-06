package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
)

func Initialize() *pg.DB {
	fmt.Println("Connecting to Postgres database")
	db := pg.Connect(&pg.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Database: os.Getenv("POSTGRES_DB"),
	})

	ctx := context.Background()
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		if err := db.Ping(ctx); err != nil {
			if i == maxRetries-1 {
				panic(err)
			}
			fmt.Printf("Failed to connect to database, attempt %d/%d: %v\n", i+1, maxRetries, err)
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Println("Successfully connected to database")
		break
	}

	return db
}

type SearchResult struct {
	id string `pg:"type:uuid,pk,unique"`
}

func Search(db *pg.DB) []SearchResult {
	var results []SearchResult
	err := db.Model(&results).Select()

	if err != nil {
		fmt.Println("Error searching for results: ", err)
	}

	return results
}
