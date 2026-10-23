package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func fetch(client *http.Client, endpoint string) {
	resp, err := client.Get(endpoint)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 150))
	if err != nil {
		fmt.Println("Read error:", err)
		return
	}
	fmt.Printf("GET %s\nHTTP status: %d\nContent-Type: %s\nBody prefix: %s\n\n",
		endpoint, resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func main() {
	client := &http.Client{Timeout: time.Second}
	fetch(client, "http://127.0.0.1:18084/books?page=1")
	fetch(client, "http://127.0.0.1:18084/bad-status")
}
