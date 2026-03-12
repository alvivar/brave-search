package main

import (
	"bytes"
	"compress/gzip"
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const endpoint = "https://api.search.brave.com/res/v1/web/search"

//go:embed .env
var embeddedEnv string

type searchRequest struct {
	Query      string `json:"q"`
	Country    string `json:"country"`
	SearchLang string `json:"search_lang"`
	Count      int    `json:"count"`
}

func main() {
	apiKey := resolveAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "no API key found; set BRAVE_API_KEY or add BRAVE_API_KEY/search_key/answer_key to .env before building")
		os.Exit(1)
	}

	q := flag.String("q", "Brave Search", "search query")
	country := flag.String("country", "US", "country code")
	searchLang := flag.String("search-lang", "en", "search language")
	count := flag.Int("count", 20, "number of results")
	flag.Parse()

	payload := searchRequest{
		Query:      *q,
		Country:    *country,
		SearchLang: *searchLang,
		Count:      *count,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal request: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		fmt.Fprintf(os.Stderr, "create request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("x-subscription-token", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "send request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	reader := io.Reader(resp.Body)
	if resp.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open gzip response: %v\n", err)
			os.Exit(1)
		}
		defer gz.Close()
		reader = gz
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read response: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		fmt.Fprintf(os.Stderr, "request failed: %s\n%s\n", resp.Status, respBody)
		os.Exit(1)
	}

	var pretty bytes.Buffer
	if err := json.Indent(&pretty, respBody, "", "  "); err != nil {
		fmt.Println(string(respBody))
		return
	}

	fmt.Println(pretty.String())
}

func resolveAPIKey() string {
	if value := strings.TrimSpace(os.Getenv("BRAVE_API_KEY")); value != "" {
		return value
	}

	values := parseDotEnv(embeddedEnv)
	for _, key := range []string{"BRAVE_API_KEY", "search_key", "answer_key"} {
		if value := strings.TrimSpace(values[key]); value != "" {
			return value
		}
	}

	return ""
}

func parseDotEnv(content string) map[string]string {
	values := make(map[string]string)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		values[key] = value
	}

	return values
}
