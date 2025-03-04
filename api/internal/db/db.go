package db

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
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

	// Load CA cert if given
	ca_env, ca_env_defined := os.LookupEnv("CA_CERT")
	ca_path, ca_path_defined := os.LookupEnv("CA_PATH")

	var tlsConfig *tls.Config

	if ca_path_defined || ca_env_defined {
		var caCert []byte
		if ca_path_defined {
			var err error
			caCert, err = os.ReadFile(ca_path)
			if err != nil {
				panic(fmt.Sprintf("failed to read CA certificate: %v", err))
			}
		}
		if ca_env_defined {
			data, err := base64.StdEncoding.DecodeString(ca_env)
			if err != nil {
				panic(fmt.Sprintf("failed to decode CA certificate: %v", err))
			}
			caCert = data
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig =
			&tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: false,
				ServerName:         os.Getenv("POSTGRES_HOST"),
			}
	}

	db := pg.Connect(&pg.Options{
		Addr:      fmt.Sprintf("%s:%s", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT")),
		User:      os.Getenv("POSTGRES_USER"),
		Password:  os.Getenv("POSTGRES_PASSWORD"),
		Database:  os.Getenv("POSTGRES_DB"),
		TLSConfig: tlsConfig,
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

func Search(db *pg.DB, query string, lemmatize map[string]string, transforms map[string]string) (*SearchResponse, error) {
	start := time.Now()

	keywords := strings.Fields(strings.ToLower(query))
	processed_keywords := []string{}

	for _, k := range keywords {
		k = strings.Trim(k, "\n\t ")
		k = strings.TrimRight(k, "'s")

		// Returns characters (runes) between a-z and 0-9
		clean := func(r rune) rune {
			switch {
			case r >= 'a' && r <= 'z':
				return r
			case r >= '0' && r <= '9':
				return r
			}
			return -1
		}
		k = strings.Map(clean, k)

		lem, ok := lemmatize[k]
		if ok {
			k = lem
		}

		tra, ok := transforms[k]
		if ok {
			k = tra
		}

		processed_keywords = append(processed_keywords, k)
	}

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
	`, pg.Array(processed_keywords))

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
