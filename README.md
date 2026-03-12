# Brave Search CLI

A small Go command-line app for calling the Brave Search Web Search API.

It sends a `POST` request to:

```text
https://api.search.brave.com/res/v1/web/search
```

The app supports:

- embedded `.env` values at build time
- `BRAVE_API_KEY` environment override
- positional or flag-based queries
- pretty JSON, raw output, or titles-only output
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

The binary looks up the API key in this order:

1. `BRAVE_API_KEY` environment variable
2. `BRAVE_API_KEY` in embedded `.env`
3. `search_key` in embedded `.env`
4. `answer_key` in embedded `.env`

Example `.env`:

```env
BRAVE_API_KEY=your_api_key_here
```

or:

```env
search_key=your_api_key_here
```

> `.env` is embedded into the binary at build time. If you change `.env`, rebuild the app.

## Usage

Run with no arguments to show help:

```bash
brave-search
```

Search with a positional query:

```bash
brave-search "Brave Search"
```

Search with flags:

```bash
brave-search -q "Brave Search" -country US -search-lang en -count 20
```

Print titles and URLs only:

```bash
brave-search -titles "golang http client"
```

Print the raw API response:

```bash
brave-search -raw "Brave Search"
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

- `-q` search query
- `-country` country code, default `US`
- `-search-lang` search language, default `en`
- `-count` number of results, default `20`
- `-timeout` request timeout, default `30s`
- `-titles` print titles, URLs, and descriptions only
- `-raw` print raw response body

## Example curl equivalent

This app is equivalent to a request like:

```bash
curl -s --compressed -X 'POST' \
  'https://api.search.brave.com/res/v1/web/search' \
  -H 'accept: application/json' \
  -H 'Accept-Encoding: gzip' \
  -H 'x-subscription-token: <YOUR_API_KEY>' \
  -H 'Content-Type: application/json' \
  -d '{
  "q": "Brave Search",
  "country": "US",
  "search_lang": "en",
  "count": 20
}'
```

## Notes

- If the API returns a non-2xx status, the app prints the error body.
- Authentication and rate-limit errors are reported with clearer messages.
- The default output format is pretty-printed JSON.
