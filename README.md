# Brave CLI

A small Go command-line app for calling Brave Search APIs.

It supports both:

- **Web Search API** for search results
- **Answers API** for AI-generated answers backed by Brave search

Endpoints used:

```text
POST https://api.search.brave.com/res/v1/web/search
POST https://api.search.brave.com/res/v1/chat/completions
```

The app supports:

- embedded `.env` values at build time
- `BRAVE_API_KEY` environment override
- positional or flag-based queries
- human-friendly CLI output by default
- compact titles-only, pretty JSON, or raw output modes
- gzip-compressed responses

## Build

```bash
go build -o brave-search.exe .
```

On non-Windows:

```bash
go build -o brave-search .
```

## API key configuration

### Search mode lookup order

1. `BRAVE_API_KEY` environment variable
2. `BRAVE_SEARCH_API_KEY` environment variable
3. `BRAVE_API_KEY` in embedded `.env`
4. `search_key` in embedded `.env`
5. `answer_key` in embedded `.env`

### Answer mode lookup order

1. `BRAVE_API_KEY` environment variable
2. `BRAVE_ANSWER_API_KEY` environment variable
3. `BRAVE_API_KEY` in embedded `.env`
4. `answer_key` in embedded `.env`
5. `search_key` in embedded `.env`

Example `.env`:

```env
search_key=your_search_api_key_here
answer_key=your_answer_api_key_here
```

You can also use:

```env
BRAVE_API_KEY=your_api_key_here
```

> `.env` is embedded into the binary at build time. If you change `.env`, rebuild the app.

## Usage

Run with no arguments to show help:

```bash
brave-search
```

### Web search

Search with a positional query:

```bash
brave-search "Brave Search"
```

Search with flags:

```bash
brave-search -q "Brave Search" -country US -search-lang en -count 20
```

Print compact titles and URLs only:

```bash
brave-search -titles "golang http client"
```

### Answers API

Ask for an AI-generated answer:

```bash
brave-search -answer "What is the second highest mountain?"
```

### Output modes

Print pretty JSON:

```bash
brave-search -json "Brave Search"
brave-search -answer -json "What is the second highest mountain?"
```

Print the raw API response:

```bash
brave-search -raw "Brave Search"
brave-search -answer -raw "What is the second highest mountain?"
```

Control CLI wrapping width:

```bash
brave-search -width 100 "Brave Search"
```

Set the API key via environment variable:

```bash
export BRAVE_API_KEY="your_api_key_here"
brave-search "Brave Search"
```

On PowerShell:

```powershell
$env:BRAVE_API_KEY="your_api_key_here"
.\brave-search.exe "Brave Search"
```

## Flags

- `-q` query or prompt
- `-answer` use the Answers API instead of Web Search
- `-country` country code for search mode, default `US`
- `-search-lang` search language for search mode, default `en`
- `-count` number of results for search mode, default `20`
- `-timeout` request timeout, default `30s`
- `-width` CLI output wrap width, default `88`
- `-titles` print compact titles, URLs, and descriptions in search mode
- `-json` print pretty JSON
- `-raw` print raw response body

## Default CLI output

### Search mode

By default, search results are printed in a human-friendly terminal format with:

- numbered results
- indented URLs
- wrapped descriptions
- separators between entries

### Answer mode

By default, answers are printed in a readable CLI format with:

- a heading
- the prompt used
- wrapped answer text
- token usage when available

## API details

See [`API.md`](./API.md) for the Search API and Answers API request/response details used by this project.

## Notes

- If the API returns a non-2xx status, the app prints the error body.
- Authentication and rate-limit errors are reported with clearer messages.
- The default output format is human-friendly CLI text.
- Use `-json` if you want machine-friendly pretty JSON output.
