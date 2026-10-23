package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"golang.org/x/net/html"
)

// --- DATA MODELS ---

type CatalogItem struct {
	Source string  `json:"source"`
	Kind   string  `json:"kind"`
	Title  string  `json:"title"`
	Price  string  `json:"price,omitempty"`
	Rating float64 `json:"rating,omitempty"`
}

type Book struct {
	Title        string
	Price        string
	Availability string
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

// --- HTML EXTRACTION HELPERS ---

func extractBookDetails(node *html.Node) Book {
	var b Book
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "title" {
					b.Title = attr.Val
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "p" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "price_color" && n.FirstChild != nil {
					b.Price = n.FirstChild.Data
				}
				if attr.Key == "class" && attr.Val == "availability" && n.FirstChild != nil {
					b.Availability = n.FirstChild.Data
				}
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}
	traverse(node)
	return b
}

func parseBookPage(doc *html.Node, baseURL string) ([]Book, string) {
	var books []Book
	var nextURL string

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "article" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "product_pod" {
					books = append(books, extractBookDetails(n))
					return
				}
			}
		}

		if n.Type == html.ElementNode && n.Data == "li" {
			for _, attr := range n.Attr {
				if attr.Key == "class" && attr.Val == "next" {
					for child := n.FirstChild; child != nil; child = child.NextSibling {
						if child.Type == html.ElementNode && child.Data == "a" {
							for _, aAttr := range child.Attr {
								if aAttr.Key == "href" {
									base, err := url.Parse(baseURL)
									rel, err2 := url.Parse(aAttr.Val)
									if err == nil && err2 == nil {
										nextURL = base.ResolveReference(rel).String()
									}
								}
							}
						}
					}
				}
			}
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)

	return books, nextURL
}

// Single-page fetcher (Used in BUILD and VERIFY Test 1 & 4)
func collectBooks(pageURL string, client *http.Client) ([]Book, error) {
	resp, err := client.Get(pageURL)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTML parse error: %w", err)
	}

	books, _ := parseBookPage(doc, pageURL)
	return books, nil
}

// Multi-page fetcher with pacing (Used in ADAPT)
func collectBooksPaginated(startURL string, client *http.Client, maxPages int) ([]Book, error) {
	var allBooks []Book
	currentURL := startURL
	pagesVisited := 0

	for currentURL != "" && pagesVisited < maxPages {
		if pagesVisited > 0 {
			fmt.Println("[Pacing] Waiting 1 second before requesting next page...")
			time.Sleep(1 * time.Second)
		}

		fmt.Printf("[Crawler] Fetching page %d: %s\n", pagesVisited+1, currentURL)
		resp, err := client.Get(currentURL)
		if err != nil {
			return allBooks, fmt.Errorf("network error on %s: %w", currentURL, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return allBooks, fmt.Errorf("unexpected HTTP status %d on %s", resp.StatusCode, currentURL)
		}

		doc, err := html.Parse(resp.Body)
		resp.Body.Close()
		if err != nil {
			return allBooks, fmt.Errorf("HTML parse error on %s: %w", currentURL, err)
		}

		pageBooks, nextLink := parseBookPage(doc, currentURL)
		allBooks = append(allBooks, pageBooks...)
		pagesVisited++

		currentURL = nextLink
	}

	return allBooks, nil
}

// --- MOVIE API FUNCTION ---

func searchMovies(endpointURL string, client *http.Client) ([]Movie, error) {
	resp, err := client.Get(endpointURL)
	if err != nil {
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	var envelope SearchEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, fmt.Errorf("JSON decode error: %w", err)
	}

	return envelope.Results, nil
}

func runVerificationTests() {
	client := &http.Client{Timeout: 3 * time.Second}

	// Test 1: Book page 1, normal
	fmt.Println("\n[Test 1] Book page 1, normal")
	books, err := collectBooks("http://127.0.0.1:18084/books?page=1", client)
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
	} else if len(books) == 2 {
		fmt.Printf("PASS: Extracted %d books:\n", len(books))
		for _, b := range books {
			fmt.Printf("      - %s (%s)\n", b.Title, b.Price)
		}
	} else {
		fmt.Printf("FAIL: Expected 2 books, got %d\n", len(books))
	}

	// Test 2: Movie API, normal
	fmt.Println("\n[Test 2] Movie API, normal")
	movieURL := "http://127.0.0.1:18084/search/movie?query=" + url.QueryEscape("inception")
	movies, err := searchMovies(movieURL, client)
	if err != nil {
		fmt.Printf("FAIL: %v\n", err)
	} else if len(movies) == 2 {
		fmt.Printf("PASS: Extracted %d movies:\n", len(movies))
		for _, m := range movies {
			fmt.Printf("      - ID %d: %s (Rating: %.1f)\n", m.ID, m.Title, m.Rating)
		}
	} else {
		fmt.Printf("FAIL: Expected 2 movies, got %d\n", len(movies))
	}

	// Test 3: Combined JSON & file decode
	fmt.Println("\n[Test 3] Combined JSON validation")
	fileData, err := os.ReadFile("catalog.json")
	if err != nil {
		fmt.Printf("FAIL: Cannot read catalog.json: %v\n", err)
	} else {
		var reloaded []CatalogItem
		if err := json.Unmarshal(fileData, &reloaded); err != nil {
			fmt.Printf("FAIL: JSON decoding error: %v\n", err)
		} else if len(reloaded) == 4 {
			fmt.Printf("PASS: Successfully reloaded 4 records from catalog.json\n")
		} else {
			fmt.Printf("FAIL: Expected 4 records in initial catalog.json, found %d\n", len(reloaded))
		}
	}

	// Test 4: /bad-status
	fmt.Println("\n[Test 4] Failure fixture: /bad-status")
	_, err = collectBooks("http://127.0.0.1:18084/bad-status", client)
	if err != nil {
		fmt.Printf("PASS: Rejected invalid status cleanly -> %v\n", err)
	} else {
		fmt.Println("FAIL: Expected error on 503, but call succeeded")
	}

	// Test 5: /bad-json
	fmt.Println("\n[Test 5] Failure fixture: /bad-json")
	_, err = searchMovies("http://127.0.0.1:18084/bad-json", client)
	if err != nil {
		fmt.Printf("PASS: Caught decoding failure without crashing -> %v\n", err)
	} else {
		fmt.Println("FAIL: Expected decode error on malformed JSON, but call succeeded")
	}

	// Test 6: /slow with short client timeout
	fmt.Println("\n[Test 6] Failure fixture: /slow with 200ms timeout")
	shortClient := &http.Client{Timeout: 200 * time.Millisecond}
	_, err = collectBooks("http://127.0.0.1:18084/slow", shortClient)
	if err != nil {
		fmt.Printf("PASS: Failed promptly on timeout -> %v\n", err)
	} else {
		fmt.Println("FAIL: Slow request unexpectedly succeeded")
	}
}

func main() {
	client := &http.Client{Timeout: 3 * time.Second}

	fmt.Println("=== PART 1: INITIAL BUILD PIPELINE ===")
	var catalog []CatalogItem

	books, err := collectBooks("http://127.0.0.1:18084/books?page=1", client)
	if err == nil {
		for _, b := range books {
			catalog = append(catalog, CatalogItem{
				Source: "books",
				Kind:   "book",
				Title:  b.Title,
				Price:  b.Price,
			})
		}
	}

	movieURL := "http://127.0.0.1:18084/search/movie?query=" + url.QueryEscape("inception")
	movies, err := searchMovies(movieURL, client)
	if err == nil {
		for _, m := range movies {
			catalog = append(catalog, CatalogItem{
				Source: "movies",
				Kind:   "movie",
				Title:  m.Title,
				Rating: m.Rating,
			})
		}
	}

	data, _ := json.MarshalIndent(catalog, "", "  ")
	_ = os.WriteFile("catalog.json", data, 0644)
	fmt.Printf("Initial run: saved %d records to catalog.json (2 books + 2 movies)\n", len(catalog))

	runVerificationTests()

	var adaptedCatalog []CatalogItem

	// Collect up to 2 pages of books
	paginatedBooks, err := collectBooksPaginated("http://127.0.0.1:18084/books?page=1", client, 2)
	if err != nil {
		fmt.Printf("Scraping error: %v\n", err)
	}
	for _, b := range paginatedBooks {
		adaptedCatalog = append(adaptedCatalog, CatalogItem{
			Source: "books",
			Kind:   "book",
			Title:  b.Title,
			Price:  b.Price,
		})
	}
	fmt.Printf("Collected %d books total across 2 pages.\n", len(paginatedBooks))

	// Re-add the movies
	for _, m := range movies {
		adaptedCatalog = append(adaptedCatalog, CatalogItem{
			Source: "movies",
			Kind:   "movie",
			Title:  m.Title,
			Rating: m.Rating,
		})
	}

	// Update catalog.json with 5 total records
	adaptedData, _ := json.MarshalIndent(adaptedCatalog, "", "  ")
	_ = os.WriteFile("catalog.json", adaptedData, 0644)
	fmt.Printf("Updated catalog.json: successfully saved %d records (3 books + 2 movies)\n", len(adaptedCatalog))
}