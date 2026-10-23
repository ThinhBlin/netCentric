package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

type Movie struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	ReleaseDate string  `json:"release_date"`
	Rating      float64 `json:"vote_average"`
}

type SearchResponse struct {
	Page    int     `json:"page"`
	Results []Movie `json:"results"`
}

func main() {
	endpoint := "http://127.0.0.1:18084/search/movie?query=" + url.QueryEscape("inception")
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(endpoint)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("HTTP status:", resp.StatusCode)
		return
	}
	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("JSON error:", err)
		return
	}
	for _, m := range result.Results {
		fmt.Printf("%d: %s (%.1f)\n", m.ID, m.Title, m.Rating)
	}
	data, err := json.MarshalIndent(result.Results, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile("example_movies.json", data, 0644); err != nil {
		fmt.Println("Save error:", err)
		return
	}
	fmt.Println("Saved example_movies.json")
}
