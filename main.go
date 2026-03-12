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
	"unicode/utf8"
)

const (
	searchEndpoint   = "https://api.search.brave.com/res/v1/web/search"
	answerEndpoint   = "https://api.search.brave.com/res/v1/chat/completions"
	defaultWrapWidth = 88
	minimumWrapWidth = 40
	resultIndent     = "    "
	separator        = "────────────────────────────────────────────────────────────────────────────────"
)

//go:embed .env
var embeddedEnv string

type searchRequest struct {
	Query      string `json:"q"`
	Country    string `json:"country"`
	SearchLang string `json:"search_lang"`
	Count      int    `json:"count"`
}

type searchResponse struct {
	Query struct {
		Original string `json:"original"`
	} `json:"query"`
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

type answerRequest struct {
	Stream   bool            `json:"stream"`
	Messages []answerMessage `json:"messages"`
}

type answerMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type answerResponse struct {
	Model   string `json:"model"`
	ID      string `json:"id"`
	Object  string `json:"object"`
	Choices []struct {
		Index        int    `json:"index"`
		FinishReason string `json:"finish_reason"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Delta struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Usage struct {
		CompletionTokens int `json:"completion_tokens"`
		PromptTokens     int `json:"prompt_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "Brave CLI\n\n")
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "  brave-search [flags] <query>\n\n")
		fmt.Fprintf(out, "Examples:\n")
		fmt.Fprintf(out, "  brave-search \"Brave Search\"\n")
		fmt.Fprintf(out, "  brave-search -q \"golang http client\" -count 10\n")
		fmt.Fprintf(out, "  brave-search -answer \"What is the second highest mountain?\"\n")
		fmt.Fprintf(out, "  brave-search -titles \"golang http client\"\n")
		fmt.Fprintf(out, "  brave-search -json \"golang http client\"\n\n")
		fmt.Fprintf(out, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(out, "\nEmbedded API keys (.env at build time):\n")
		fmt.Fprintf(out, "  Search mode:\n")
		fmt.Fprintf(out, "    1. SEARCH_KEY\n")
		fmt.Fprintf(out, "    2. ANSWER_KEY\n")
		fmt.Fprintf(out, "  Answer mode:\n")
		fmt.Fprintf(out, "    1. ANSWER_KEY\n")
		fmt.Fprintf(out, "    2. SEARCH_KEY\n")
	}

	q := flag.String("q", "", "query or prompt")
	country := flag.String("country", "US", "country code for search mode")
	searchLang := flag.String("search-lang", "en", "search language for search mode")
	count := flag.Int("count", 20, "number of results for search mode")
	timeout := flag.Duration("timeout", 30*time.Second, "request timeout")
	width := flag.Int("width", defaultWrapWidth, "CLI output wrap width")
	answerMode := flag.Bool("answer", false, "use the Answers API instead of Web Search")
	titlesOnly := flag.Bool("titles", false, "print compact titles and URLs only in search mode")
	jsonOutput := flag.Bool("json", false, "print pretty JSON")
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

	mode := "search"
	if *answerMode {
		mode = "answer"
	}

	apiKey := resolveAPIKey(mode)
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "no embedded API key found; add SEARCH_KEY and/or ANSWER_KEY to .env and rebuild")
		os.Exit(1)
	}

	var respBody []byte
	var err error
	if *answerMode {
		respBody, err = callAnswerAPI(apiKey, query, *timeout)
	} else {
		respBody, err = callSearchAPI(apiKey, query, *country, *searchLang, *count, *timeout)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *raw {
		fmt.Println(string(respBody))
		return
	}

	if *jsonOutput {
		printPrettyJSON(respBody)
		return
	}

	if *answerMode {
		printAnswer(respBody, query, sanitizeWidth(*width))
		return
	}

	if *titlesOnly {
		printTitles(respBody, sanitizeWidth(*width))
		return
	}

	printCLIResults(respBody, query, sanitizeWidth(*width))
}

func callSearchAPI(apiKey, query, country, searchLang string, count int, timeout time.Duration) ([]byte, error) {
	payload := searchRequest{
		Query:      query,
		Country:    country,
		SearchLang: searchLang,
		Count:      count,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal search request: %w", err)
	}

	return doJSONRequest(searchEndpoint, apiKey, body, timeout)
}

func callAnswerAPI(apiKey, prompt string, timeout time.Duration) ([]byte, error) {
	payload := answerRequest{
		Stream: false,
		Messages: []answerMessage{{
			Role:    "user",
			Content: prompt,
		}},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal answer request: %w", err)
	}

	return doJSONRequest(answerEndpoint, apiKey, body, timeout)
}

func doJSONRequest(endpoint, apiKey string, body []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("x-subscription-token", apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	reader := io.Reader(resp.Body)
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("open gzip response: %w", err)
		}
		defer gz.Close()
		reader = gz
	}

	respBody, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, formatHTTPError(resp.StatusCode, resp.Status, respBody)
	}

	return respBody, nil
}

func formatHTTPError(statusCode int, status string, body []byte) error {
	var message string
	switch statusCode {
	case http.StatusUnauthorized, http.StatusForbidden:
		message = fmt.Sprintf("authentication failed: %s", status)
	case http.StatusTooManyRequests:
		message = fmt.Sprintf("rate limited: %s", status)
	default:
		message = fmt.Sprintf("request failed: %s", status)
	}

	if len(body) == 0 {
		return fmt.Errorf("%s", message)
	}
	return fmt.Errorf("%s\n%s", message, body)
}

func printCLIResults(body []byte, fallbackQuery string, width int) {
	var response searchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		printPrettyJSON(body)
		return
	}

	query := strings.TrimSpace(response.Query.Original)
	if query == "" {
		query = fallbackQuery
	}

	results := response.Web.Results
	if len(results) == 0 {
		fmt.Printf("No web results for %q.\n", query)
		return
	}

	fmt.Printf("Brave Search results for %q\n", query)
	fmt.Printf("%d result(s)\n", len(results))
	fmt.Println(separator)

	for i, result := range results {
		title := strings.TrimSpace(result.Title)
		if title == "" {
			title = "(untitled result)"
		}

		fmt.Printf("[%d] %s\n", i+1, title)
		fmt.Println(indentLines(strings.TrimSpace(result.URL), resultIndent))

		description := strings.TrimSpace(result.Description)
		if description != "" {
			fmt.Println()
			for _, line := range wrapText(description, width-len(resultIndent)) {
				fmt.Println(resultIndent + line)
			}
		}

		if i < len(results)-1 {
			fmt.Println()
			fmt.Println(separator)
		}
	}
}

func printAnswer(body []byte, prompt string, width int) {
	var response answerResponse
	if err := json.Unmarshal(body, &response); err != nil {
		printPrettyJSON(body)
		return
	}

	content := extractAnswerText(response)
	if content == "" {
		printPrettyJSON(body)
		return
	}

	fmt.Println("Brave Answer")
	fmt.Println(separator)
	fmt.Printf("Prompt: %s\n", prompt)
	if model := strings.TrimSpace(response.Model); model != "" {
		fmt.Printf("Model:  %s\n", model)
	}
	fmt.Println()

	for _, line := range wrapParagraphs(content, width) {
		fmt.Println(line)
	}

	if response.Usage.TotalTokens > 0 {
		fmt.Println()
		fmt.Println(separator)
		fmt.Printf("Tokens: prompt=%d completion=%d total=%d\n", response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)
	}
}

func extractAnswerText(response answerResponse) string {
	parts := make([]string, 0, len(response.Choices))
	for _, choice := range response.Choices {
		content := strings.TrimSpace(choice.Message.Content)
		if content == "" {
			content = strings.TrimSpace(choice.Delta.Content)
		}
		if content != "" {
			parts = append(parts, content)
		}
	}
	return strings.TrimSpace(strings.Join(parts, "\n\n"))
}

func printTitles(body []byte, width int) {
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
		title := strings.TrimSpace(result.Title)
		if title == "" {
			title = "(untitled result)"
		}

		fmt.Printf("%d. %s\n", i+1, title)
		fmt.Printf("   %s\n", strings.TrimSpace(result.URL))
		if description := strings.TrimSpace(result.Description); description != "" {
			for _, line := range wrapText(description, width-3) {
				fmt.Printf("   %s\n", line)
			}
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

func wrapParagraphs(text string, width int) []string {
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		paragraph = strings.TrimSpace(paragraph)
		if paragraph == "" {
			if len(lines) > 0 && lines[len(lines)-1] != "" {
				lines = append(lines, "")
			}
			continue
		}
		lines = append(lines, wrapText(paragraph, width)...)
	}
	return lines
}

func wrapText(text string, width int) []string {
	text = strings.Join(strings.Fields(text), " ")
	if text == "" {
		return nil
	}

	if width < minimumWrapWidth {
		width = minimumWrapWidth
	}

	words := strings.Fields(text)
	lines := make([]string, 0, len(words)/8+1)
	current := words[0]
	currentWidth := runeWidth(current)

	for _, word := range words[1:] {
		wordWidth := runeWidth(word)
		if currentWidth+1+wordWidth <= width {
			current += " " + word
			currentWidth += 1 + wordWidth
			continue
		}

		lines = append(lines, current)
		current = word
		currentWidth = wordWidth
	}

	lines = append(lines, current)
	return lines
}

func indentLines(text, indent string) string {
	if text == "" {
		return indent
	}

	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = indent + line
	}
	return strings.Join(lines, "\n")
}

func runeWidth(value string) int {
	return utf8.RuneCountInString(value)
}

func sanitizeWidth(width int) int {
	if width < minimumWrapWidth {
		return minimumWrapWidth
	}
	return width
}

func resolveAPIKey(mode string) string {
	values := parseDotEnv(embeddedEnv)

	keys := []string{"SEARCH_KEY", "ANSWER_KEY"}
	if mode == "answer" {
		keys = []string{"ANSWER_KEY", "SEARCH_KEY"}
	}
	for _, key := range keys {
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

		key = strings.ToUpper(strings.TrimSpace(key))
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		values[key] = value
	}

	return values
}
