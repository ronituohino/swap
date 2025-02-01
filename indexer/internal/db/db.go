package db

import (
	"context"
	"fmt"
	"os"

	"github.com/go-pg/pg/v10"
)

func Initialize() *pg.DB {
	fmt.Println("Connecting to Postgres database")
	db := pg.Connect(&pg.Options{
		Addr:     os.Getenv("PG_ADDR"),
		User:     os.Getenv("PG_USER"),
		Password: os.Getenv("PG_PASSWORD"),
		Database: os.Getenv("PG_DB"),
	})

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		panic(err)
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
