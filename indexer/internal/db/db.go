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
	URL   string `pg:",notnull,unique"`
	Title string
}

type Keyword struct {
	ID   int    `pg:",pk"`
	Word string `pg:",notnull,unique"`
}

type Relation struct {
	ID        int      `pg:",pk"`
	WebsiteID int      `pg:",notnull,fk:website_id,unique:website_id_keyword_id"`
	KeywordID int      `pg:",notnull,fk:keyword_id,unique:website_id_keyword_id"`
	Relevance float32  `pg:",notnull"`
	TF        float32  `pg:",notnull"`
	IDF       float32  `pg:",notnull"`
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

func InsertScrapedData(db *pg.DB, messages []ScrapedMessage) error {
	websiteMap := make(map[string]*Website)
	for _, m := range messages {
		websiteMap[m.URL] = &Website{
			URL:   m.URL,
			Title: m.Title,
		}
	}
	websites := make([]*Website, 0, len(websiteMap))
	for _, website := range websiteMap {
		websites = append(websites, website)
	}

	tx1, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting website transaction: %v", err)
	}
	_, err = tx1.Model(&websites).OnConflict("(url) DO UPDATE").
		Set("title = EXCLUDED.title").
		Returning("id").
		Insert()
	if err != nil {
		tx1.Rollback()
		return fmt.Errorf("error inserting websites: %v", err)
	}
	if err = tx1.Commit(); err != nil {
		return fmt.Errorf("error committing website transaction: %v", err)
	}

	keywordMap := make(map[string]*Keyword)
	for _, m := range messages {
		for word := range m.Keywords {
			if _, exists := keywordMap[word]; !exists {
				keywordMap[word] = &Keyword{Word: word}
			}
		}
	}
	keywords := make([]*Keyword, 0, len(keywordMap))
	for _, keyword := range keywordMap {
		keywords = append(keywords, keyword)
	}

	tx2, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting keyword transaction: %v", err)
	}
	_, err = tx2.Model(&keywords).OnConflict("DO NOTHING").
		Returning("id").
		Insert()
	if err != nil {
		tx2.Rollback()
		return fmt.Errorf("error inserting keywords: %v", err)
	}
	if err = tx2.Commit(); err != nil {
		return fmt.Errorf("error committing keyword transaction: %v", err)
	}

	relationMap := make(map[string]*Relation)
	tx3, err := db.Begin()
	if err != nil {
		return fmt.Errorf("error starting relations transaction: %v", err)
	}

	for _, website := range websites {
		msg := findMessageByURL(messages, website.URL)
		if msg == nil {
			continue
		}

		for word, props := range msg.Keywords {
			keyword := keywordMap[word]
			if keyword.ID == 0 {
				err = tx3.Model(keyword).Where("word = ?", word).Select()
				if err != nil {
					tx3.Rollback()
					return fmt.Errorf("error fetching keyword ID: %v", err)
				}
			}

			compositeKey := fmt.Sprintf("%d-%d", website.ID, keyword.ID)
			relationMap[compositeKey] = &Relation{
				WebsiteID: website.ID,
				KeywordID: keyword.ID,
				Relevance: props.Relevance,
				TF:        props.TermFrequency,
				IDF:       0.01,
			}
		}
	}

	relations := make([]*Relation, 0, len(relationMap))
	for _, relation := range relationMap {
		relations = append(relations, relation)
	}

	_, err = tx3.Model(&relations).OnConflict("(website_id, keyword_id) DO UPDATE").
		Set("relevance = EXCLUDED.relevance").
		Set("tf = EXCLUDED.tf").
		Set("idf = EXCLUDED.idf").
		Insert()
	if err != nil {
		tx3.Rollback()
		return fmt.Errorf("error inserting relations: %v", err)
	}
	if err = tx3.Commit(); err != nil {
		return fmt.Errorf("error committing relations transaction: %v", err)
	}

	return nil
}

func findMessageByURL(messages []ScrapedMessage, url string) *ScrapedMessage {
	for i := range messages {
		if messages[i].URL == url {
			return &messages[i]
		}
	}
	return nil
}
