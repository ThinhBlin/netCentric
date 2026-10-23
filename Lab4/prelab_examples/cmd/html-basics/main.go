package main

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/html"
)

func attribute(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func titles(root *html.Node) []string {
	var result []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			if title := attribute(n, "title"); title != "" {
				result = append(result, title)
			}
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return result
}

func main() {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get("http://127.0.0.1:18084/books?page=1")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status:", resp.StatusCode)
		return
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}
	for _, title := range titles(doc) {
		fmt.Println("Book title:", title)
	}
}
