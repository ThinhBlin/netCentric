package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{Timeout: 250 * time.Millisecond}
	resp, err := client.Get("http://127.0.0.1:18084/slow")
	if err != nil {
		fmt.Println("Slow request failed (expected):", err)
	} else {
		resp.Body.Close()
		fmt.Println("Unexpected success:", resp.StatusCode)
	}

	resp, err = client.Get("http://127.0.0.1:18084/bad-json")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println(resp.Status)
		return
	}
	var output map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		fmt.Println("Invalid JSON detected (expected):", err)
	}
}
