package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-pg/pg/v10"
)

type SearchResult struct {
	URL      string   `json:"url"`
	Title    string   `json:"title"`
	Score    float32  `json:"score"`
	Keywords []string `json:"keywords"`
}

type SearchResponse struct {
	Results   []SearchResult `json:"results"`
	QueryTime string         `json:"query_time"`
	TotalHits int            `json:"total_hits"`
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

	return db
}

func Search(db *pg.DB, query string) (*SearchResponse, error) {
	start := time.Now()

	keywords := strings.Fields(strings.ToLower(query))

	var results []struct {
		WebsiteID int     `pg:"website_id"`
		URL       string  `pg:"url"`
		Title     string  `pg:"title"`
		Score     float32 `pg:"score"`
	}

	_, err := db.Query(&results, `
			WITH matching_keywords AS (
					SELECT DISTINCT website_id, 
								 sum(tf * idf * relevance) as score
					FROM relations r
					JOIN keywords k ON r.keyword_id = k.id
					WHERE k.word = ANY(?::text[])
					GROUP BY website_id
			)
			SELECT w.id as website_id, 
						 w.url, 
						 w.title, 
						 mk.score
			FROM matching_keywords mk
			JOIN websites w ON w.id = mk.website_id
			ORDER BY mk.score DESC
			LIMIT 20
	`, pg.Array(keywords))

	if err != nil {
		return nil, fmt.Errorf("search query failed: %v", err)
	}

	searchResults := make([]SearchResult, 0, len(results))
	for _, r := range results {
		var websiteKeywords []string
		_, err := db.Query(&websiteKeywords, `
					SELECT k.word
					FROM relations r
					JOIN keywords k ON r.keyword_id = k.id
					WHERE r.website_id = ?
					ORDER BY (r.tf * r.idf * r.relevance) DESC
					LIMIT 5
			`, r.WebsiteID)

		if err != nil {
			return nil, fmt.Errorf("failed to fetch keywords: %v", err)
		}

		searchResults = append(searchResults, SearchResult{
			URL:      r.URL,
			Title:    r.Title,
			Score:    r.Score,
			Keywords: websiteKeywords,
		})
	}

	queryTime := time.Since(start)

	return &SearchResponse{
		Results:   searchResults,
		QueryTime: fmt.Sprintf("%.6fs", queryTime.Seconds()),
		TotalHits: len(results),
	}, nil
}
