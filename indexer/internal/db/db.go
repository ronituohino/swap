package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type Website struct {
	ID    int    `pg:",pk"`
	URL   string `pg:",notnull"`
	Title string
}

type Keyword struct {
	ID   int    `pg:",pk"`
	Word string `pg:",notnull,unique"`
}

type Relation struct {
	ID        int      `pg:",pk"`
	WebsiteID int      `pg:",notnull,fk:website_id"`
	KeywordID int      `pg:",notnull,fk:keyword_id"`
	Relevance float32  `pg:",notnull"`
	TF        float32  `pg:",notnull"`
	TFIDF     float32  `pg:",notnull"`
	Website   *Website `pg:"fk:website_id,rel:has-one"`
	Keyword   *Keyword `pg:"fk:keyword_id,rel:has-one"`
}

type KeywordProperties struct {
	TermFrequency float32 `json:"term_frequency"`
	Relevance     float32 `json:"relevance"`
}

type ScrapedMessage struct {
	URL      string                       `json:"url"`
	Title    string                       `json:"title"`
	Keywords map[string]KeywordProperties `json:"keywords"`
}

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

	models := []interface{}{
		(*Website)(nil),
		(*Keyword)(nil),
		(*Relation)(nil),
	}

	for _, model := range models {
		err := db.Model(model).CreateTable(&orm.CreateTableOptions{
			IfNotExists:   true,
			FKConstraints: true,
		})
		if err != nil {
			panic(fmt.Sprintf("Error creating table for %T: %v", model, err))
		}
	}

	return db
}

func InsertScrapedData(db *pg.DB, message ScrapedMessage) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting transaction: %v", err)
	}
	defer tx.Rollback()

	website := &Website{
		URL:   message.URL,
		Title: message.Title,
	}
	_, err = tx.Model(website).Insert()
	if err != nil {
		return fmt.Errorf("error inserting website: %v", err)
	}

	for word, props := range message.Keywords {
		keyword := &Keyword{Word: word}
		_, err = tx.Model(keyword).
			Where("word = ?", word).
			SelectOrInsert()
		if err != nil {
			return fmt.Errorf("error inserting keyword: %v", err)
		}

		relation := &Relation{
			WebsiteID: website.ID,
			KeywordID: keyword.ID,
			Relevance: props.Relevance,
			TF:        props.TermFrequency,
			TFIDF:     0.01,
		}
		_, err = tx.Model(relation).Insert()
		if err != nil {
			return fmt.Errorf("error inserting relation: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
