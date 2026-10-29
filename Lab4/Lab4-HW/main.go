package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Updated struct with the Homework requirement: SourceURL
type CatalogItem struct {
	Source    string  `json:"source"`
	Kind      string  `json:"kind"`
	Title     string  `json:"title"`
	Rating    float64 `json:"rating,omitempty"`
	SourceURL string  `json:"source_url"`
}

type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	Rating      float64 `json:"vote_average"`
}

type SearchEnvelope struct {
	Page    int     `json:"page"`
	Results []Movie `json:"results"`
}

// searchTMDB hits the live public API using an environment variable for authentication
func searchTMDB(query string, client *http.Client) ([]Movie, string, error) {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return nil, "", fmt.Errorf("missing TMDB_API_KEY environment variable. Please set it before running")
	}

	// Construct the real TMDB endpoint
	endpointURL := "https://api.themoviedb.org/3/search/movie?query=" + url.QueryEscape(query)

	req, err := http.NewRequest("GET", endpointURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("request creation error: %w", err)
	}

	// TMDB uses Bearer token authentication
	req.Header.Add("Authorization", "Bearer "+apiKey)
	req.Header.Add("accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, endpointURL, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, endpointURL, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	var envelope SearchEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, endpointURL, fmt.Errorf("JSON decode error: %w", err)
	}

	// Homework requirement: Max 10 records
	results := envelope.Results
	if len(results) > 10 {
		results = results[:10]
	}

	return results, endpointURL, nil
}

func main() {
	client := &http.Client{Timeout: 5 * time.Second}
	var catalog []CatalogItem

	// Search for a movie
	searchQuery := "inception"
	fmt.Printf("Searching live TMDB API for: %q\n", searchQuery)

	movies, sourceEndpoint, err := searchTMDB(searchQuery, client)
	if err != nil {
		fmt.Printf("TMDB extraction failed: %v\n", err)
		return
	}

	// Normalize movies and inject the SourceURL
	for _, m := range movies {
		catalog = append(catalog, CatalogItem{
			Source:    "TMDB API",
			Kind:      "movie",
			Title:     m.Title,
			Rating:    m.Rating,
			SourceURL: sourceEndpoint, // Assigning provenance
		})
	}

	// Save to JSON
	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		fmt.Println("Failed to format JSON:", err)
		return
	}

	if err := os.WriteFile("homework_catalog.json", data, 0644); err != nil {
		fmt.Println("Failed to write file:", err)
		return
	}

	fmt.Printf("Successfully collected and saved %d records to homework_catalog.json\n", len(catalog))
}
