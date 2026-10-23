package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func books(w http.ResponseWriter, r *http.Request) {
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	switch page {
	case "1":
		fmt.Fprint(w, `<!doctype html><html><body>
<article class="product_pod"><h3><a title="The Go Handbook" href="/book/go">Go</a></h3><p class="price_color">£12.50</p><p class="availability">In stock</p></article>
<article class="product_pod"><h3><a title="Networking Basics" href="/book/net">Net</a></h3><p class="price_color">£25.00</p><p class="availability">In stock</p></article>
<ul><li class="next"><a href="/books?page=2">next</a></li></ul>
</body></html>`)
	case "2":
		fmt.Fprint(w, `<!doctype html><html><body>
<article class="product_pod"><h3><a title="Reliable Systems" href="/book/reliable">Systems</a></h3><p class="price_color">£30.00</p><p class="availability">In stock</p></article>
</body></html>`)
	default:
		http.NotFound(w, r)
	}
}

func movieSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Query().Get("query") == "" {
		http.Error(w, `{"error":"missing query"}`, http.StatusBadRequest)
		return
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"page": 1,
		"results": []map[string]any{
			{"id": 27205, "title": "Inception", "release_date": "2010-07-15", "vote_average": 8.4},
			{"id": 19995, "title": "Avatar", "release_date": "2009-12-15", "vote_average": 7.5},
		},
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/books", books)
	mux.HandleFunc("/search/movie", movieSearch)
	mux.HandleFunc("/bad-status", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server unavailable", http.StatusServiceUnavailable)
	})
	mux.HandleFunc("/bad-json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results": [broken]}`)
	})
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			fmt.Fprint(w, "slow response")
		case <-r.Context().Done():
			return
		}
	})
	server := &http.Server{
		Addr: "127.0.0.1:18084", Handler: mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	fmt.Println("Practice server: http://127.0.0.1:18084")
	fmt.Println("Try /books?page=1 and /search/movie?query=inception")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Println("Server error:", err)
	}
}
