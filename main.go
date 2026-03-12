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

type searchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Brave Search CLI\n\n")
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "  brave-search [flags] <query>\n\n")
		fmt.Fprintf(out, "Examples:\n")
		fmt.Fprintf(out, "  brave-search \"Brave Search\"\n")
		fmt.Fprintf(out, "  brave-search -q \"golang http client\" -count 10\n")
		fmt.Fprintf(out, "  brave-search -titles \"golang http client\"\n\n")
		fmt.Fprintf(out, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(out, "\nAPI key lookup order:\n")
		fmt.Fprintf(out, "  1. BRAVE_API_KEY environment variable\n")
		fmt.Fprintf(out, "  2. BRAVE_API_KEY in embedded .env\n")
		fmt.Fprintf(out, "  3. search_key in embedded .env\n")
		fmt.Fprintf(out, "  4. answer_key in embedded .env\n")
	}

	q := flag.String("q", "", "search query")
	country := flag.String("country", "US", "country code")
	searchLang := flag.String("search-lang", "en", "search language")
	count := flag.Int("count", 20, "number of results")
	timeout := flag.Duration("timeout", 30*time.Second, "request timeout")
	titlesOnly := flag.Bool("titles", false, "print titles and URLs only")
	raw := flag.Bool("raw", false, "print raw response body")
	flag.Parse()

	query := strings.TrimSpace(*q)
	if query == "" && flag.NArg() > 0 {
		query = strings.TrimSpace(strings.Join(flag.Args(), " "))
	}

	if query == "" {
		flag.Usage()
		return
	}

	apiKey := resolveAPIKey()
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "no API key found; set BRAVE_API_KEY or add BRAVE_API_KEY/search_key/answer_key to .env before building")
		os.Exit(1)
	}

	payload := searchRequest{
		Query:      query,
		Country:    *country,
		SearchLang: *searchLang,
		Count:      *count,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal request: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
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
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
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
		handleHTTPError(resp.StatusCode, resp.Status, respBody)
	}

	if *raw {
		fmt.Println(string(respBody))
		return
	}

	if *titlesOnly {
		printTitles(respBody)
		return
	}

	printPrettyJSON(respBody)
}

func handleHTTPError(statusCode int, status string, body []byte) {
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		fmt.Fprintf(os.Stderr, "authentication failed: %s\n", status)
	case http.StatusTooManyRequests:
		fmt.Fprintf(os.Stderr, "rate limited: %s\n", status)
	default:
		fmt.Fprintf(os.Stderr, "request failed: %s\n", status)
	}

	if len(body) > 0 {
		fmt.Fprintf(os.Stderr, "%s\n", body)
	}
	os.Exit(1)
}

func printTitles(body []byte) {
	var response searchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		printPrettyJSON(body)
		return
	}

	if len(response.Web.Results) == 0 {
		fmt.Println("No web results.")
		return
	}

	for i, result := range response.Web.Results {
		fmt.Printf("%d. %s\n", i+1, strings.TrimSpace(result.Title))
		fmt.Printf("   %s\n", strings.TrimSpace(result.URL))
		if description := strings.TrimSpace(result.Description); description != "" {
			fmt.Printf("   %s\n", description)
		}
		fmt.Println()
	}
}

func printPrettyJSON(body []byte) {
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		fmt.Println(string(body))
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
