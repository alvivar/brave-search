# Brave Search APIs

This document covers both of the APIs used by this project with equal emphasis:

- **Search API** — Search the web from a large independent index of web pages.
- **Answers API (Chat Completions)** — API for AI-generated answers backed by real-time web search and verifiable sources.

## API overview

| API                        | Purpose                                                               | Endpoint                                                    |
| -------------------------- | --------------------------------------------------------------------- | ----------------------------------------------------------- |
| Search                     | Search the web from a large independent index of web pages.           | `POST https://api.search.brave.com/res/v1/web/search`       |
| Answers (Chat Completions) | Generate AI answers backed by live web search and verifiable sources. | `POST https://api.search.brave.com/res/v1/chat/completions` |

## Authentication

Both APIs use the same authentication header:

```http
x-subscription-token: <YOUR_API_KEY>
```

## Search API

Search the web from a large independent index of web pages.

### Endpoint

- Method: `POST`
- URL: `https://api.search.brave.com/res/v1/web/search`
- Content type: `application/json`
- Response type: `application/json`
- Compression: `gzip` supported via `Accept-Encoding: gzip`

### Example request

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

### Request body

The CLI currently sends this JSON payload:

```json
{
  "q": "Brave Search",
  "country": "US",
  "search_lang": "en",
  "count": 20
}
```

### Request fields

| Field         | Type   | Required | Description                                              |
| ------------- | ------ | -------- | -------------------------------------------------------- |
| `q`           | string | Yes      | Search query text.                                       |
| `country`     | string | No       | Country code used for regionalization, for example `US`. |
| `search_lang` | string | No       | Preferred search language, for example `en`.             |
| `count`       | number | No       | Number of requested results.                             |

### Example response

The Brave Search Web Search API can return a rich response containing multiple result sections. A sample response shape is shown below:

```json
{
  "type": "search",
  "query": {
    "original": "string",
    "show_strict_warning": true,
    "altered": "string",
    "cleaned": "string",
    "safesearch": true,
    "is_navigational": true,
    "is_geolocal": true,
    "local_decision": "string",
    "local_locations_idx": 1,
    "is_trending": true,
    "is_news_breaking": true,
    "ask_for_location": true,
    "language": {
      "main": "string"
    },
    "spellcheck_off": true,
    "country": "string",
    "bad_results": true,
    "should_fallback": true,
    "lat": "string",
    "long": "string",
    "postal_code": "string",
    "city": "string",
    "header_country": "string",
    "more_results_available": true,
    "state": "string",
    "custom_location_label": "string",
    "reddit_cluster": "string",
    "summary_key": "string",
    "search_operators": {
      "applied": false,
      "cleaned_query": "string",
      "sites": ["string"]
    }
  },
  "discussions": {
    "type": "search",
    "results": [
      {
        "title": "string",
        "url": "string",
        "is_source_local": false,
        "is_source_both": false,
        "description": "",
        "page_age": "string",
        "page_fetched": "string",
        "fetched_content_timestamp": 1,
        "profile": {
          "name": "string",
          "url": "string",
          "long_name": "string",
          "img": "string"
        },
        "language": "en",
        "...": "[Additional Properties Truncated]"
      }
    ],
    "mutated_by_goggles": false
  },
  "faq": {
    "type": "faq",
    "results": [
      {
        "question": "string",
        "answer": "string",
        "title": "string",
        "url": "string",
        "meta_url": {
          "scheme": "string",
          "netloc": "string",
          "hostname": "string",
          "favicon": "string",
          "path": "string"
        }
      }
    ]
  },
  "infobox": {
    "type": "graph",
    "results": [
      {
        "title": "string",
        "url": "string",
        "is_source_local": false,
        "is_source_both": false,
        "description": "",
        "page_age": "string",
        "page_fetched": "string",
        "fetched_content_timestamp": 1,
        "profile": {
          "name": "string",
          "url": "string",
          "long_name": "string",
          "img": "string"
        },
        "language": "string",
        "...": "[Additional Properties Truncated]"
      }
    ]
  },
  "locations": {
    "type": "locations",
    "results": [
      {
        "title": "string",
        "url": "string",
        "is_source_local": false,
        "is_source_both": false,
        "description": "",
        "page_age": "string",
        "page_fetched": "string",
        "fetched_content_timestamp": 1,
        "profile": {
          "name": "string",
          "url": "string",
          "long_name": "string",
          "img": "string"
        },
        "language": "string",
        "...": "[Additional Properties Truncated]"
      }
    ],
    "provider": {}
  },
  "mixed": {
    "type": "mixed",
    "main": [
      {
        "type": "string",
        "index": 1,
        "all": false
      }
    ],
    "top": [
      {
        "type": "string",
        "index": 1,
        "all": false
      }
    ],
    "side": [
      {
        "type": "string",
        "index": 1,
        "all": false
      }
    ]
  },
  "news": {
    "type": "news",
    "results": [
      {
        "title": "string",
        "url": "string",
        "is_source_local": false,
        "is_source_both": false,
        "description": "",
        "page_age": "string",
        "page_fetched": "string",
        "fetched_content_timestamp": 1,
        "profile": {
          "name": "string",
          "url": "string",
          "long_name": "string",
          "img": "string"
        },
        "language": "string",
        "...": "[Additional Properties Truncated]"
      }
    ],
    "mutated_by_goggles": false
  },
  "videos": {
    "type": "videos",
    "results": [
      {
        "type": "video_result",
        "url": "string",
        "title": "string",
        "description": "string",
        "age": "string",
        "page_age": "string",
        "page_fetched": "string",
        "fetched_content_timestamp": 1,
        "video": {
          "duration": "string",
          "views": 1,
          "creator": "string",
          "publisher": "string",
          "requires_subscription": true,
          "tags": ["string"],
          "author": {
            "name": "string",
            "url": "string",
            "long_name": "string",
            "img": "string"
          }
        },
        "meta_url": {
          "scheme": "string",
          "netloc": "string",
          "hostname": "string",
          "favicon": "string",
          "path": "string"
        },
        "...": "[Additional Properties Truncated]"
      }
    ],
    "mutated_by_goggles": false
  },
  "web": {
    "type": "search",
    "results": [
      {
        "title": "string",
        "url": "string",
        "is_source_local": false,
        "is_source_both": false,
        "description": "",
        "page_age": "string",
        "page_fetched": "string",
        "fetched_content_timestamp": 1,
        "profile": {
          "name": "string",
          "url": "string",
          "long_name": "string",
          "img": "string"
        },
        "language": "en",
        "...": "[Additional Properties Truncated]"
      }
    ],
    "family_friendly": true
  },
  "summarizer": {
    "type": "summarizer",
    "key": "string"
  },
  "rich": {
    "type": "rich",
    "hint": {
      "vertical": "calculator",
      "callback_key": "string"
    }
  }
}
```

### Top-level response fields

| Field         | Type   | Description                                                                |
| ------------- | ------ | -------------------------------------------------------------------------- |
| `type`        | string | Overall response type, typically `search`.                                 |
| `query`       | object | Metadata about how the query was interpreted.                              |
| `discussions` | object | Discussion-style results, such as forum or community content.              |
| `faq`         | object | Frequently asked question results.                                         |
| `infobox`     | object | Knowledge graph or infobox-style content.                                  |
| `locations`   | object | Local or location-based results.                                           |
| `mixed`       | object | Layout instructions describing how sections are arranged.                  |
| `news`        | object | News results for the query.                                                |
| `videos`      | object | Video results.                                                             |
| `web`         | object | Standard web search results.                                               |
| `summarizer`  | object | Summarization metadata, including a key.                                   |
| `rich`        | object | Rich vertical hint data, such as calculator or other special result types. |

### Important sections

### `query`

The `query` object describes the processed query and search context. Common fields include:

- `original`: original query text
- `altered`: corrected or altered query text
- `cleaned`: normalized query text
- `country`: resolved country
- `language.main`: detected primary language
- `more_results_available`: whether more results exist
- `search_operators`: metadata about query operators such as `site:`

### `web`

The `web` section contains the standard search results most CLI users care about.

Example shape:

```json
{
  "type": "search",
  "results": [
    {
      "title": "string",
      "url": "string",
      "description": "string",
      "language": "en"
    }
  ],
  "family_friendly": true
}
```

Common `web.results[]` fields:

| Field                       | Type    | Description                                     |
| --------------------------- | ------- | ----------------------------------------------- |
| `title`                     | string  | Result title.                                   |
| `url`                       | string  | Result URL.                                     |
| `description`               | string  | Snippet or summary text.                        |
| `language`                  | string  | Result language code.                           |
| `page_age`                  | string  | Age of the page if available.                   |
| `page_fetched`              | string  | Fetch timestamp if available.                   |
| `fetched_content_timestamp` | number  | Fetch timestamp as a numeric value.             |
| `profile`                   | object  | Source metadata.                                |
| `is_source_local`           | boolean | Whether the source is local.                    |
| `is_source_both`            | boolean | Whether the result is both local and non-local. |

### `news`, `videos`, `discussions`, `locations`, and `infobox`

These sections follow a similar pattern:

- a `type` field
- a `results` array
- optional metadata such as `mutated_by_goggles` or `provider`

The exact fields can vary by result type.

### `mixed`

The `mixed` section tells consumers how different result blocks are arranged in the response. Each entry typically includes:

- `type`: which section to render
- `index`: index into that section’s results
- `all`: whether the full section should be rendered

### `summarizer`

The `summarizer` object may include a key that can be used with summarization-related flows:

```json
{
  "type": "summarizer",
  "key": "string"
}
```

### `rich`

The `rich` section exposes hints for rich vertical experiences:

```json
{
  "type": "rich",
  "hint": {
    "vertical": "calculator",
    "callback_key": "string"
  }
}
```

### Notes

- Not every response includes every top-level section.
- Many nested result objects may include additional fields beyond the sample shown here.
- The CLI in this repository currently uses a small subset of the response for default output:
  - `query.original`
  - `web.results[].title`
  - `web.results[].url`
  - `web.results[].description`
- If you need the full payload, use the CLI `-json` or `-raw` modes.

## Answers API (Chat Completions)

API for AI-generated answers backed by real-time web search and verifiable sources.

This endpoint accepts chat-style prompts and returns an AI-generated response.

### Endpoint

- Method: `POST`
- URL: `https://api.search.brave.com/res/v1/chat/completions`
- Content type: `application/json`
- Response type: `application/json`
- Compression: `gzip` supported via `Accept-Encoding: gzip`

### Example request

```bash
curl -X POST -s --compressed "https://api.search.brave.com/res/v1/chat/completions" \
  -H "Accept: application/json" \
  -H "Accept-Encoding: gzip" \
  -H "Content-Type: application/json" \
  -d '{"stream": false, "messages": [{"role": "user", "content": "What is the second highest mountain?"}]}' \
  -H "x-subscription-token: <YOUR_BRAVE_SEARCH_API_KEY>"
```

### Request body

Example payload:

```json
{
  "stream": false,
  "messages": [
    {
      "role": "user",
      "content": "What is the second highest mountain?"
    }
  ]
}
```

### Request fields

| Field                | Type    | Required | Description                                      |
| -------------------- | ------- | -------- | ------------------------------------------------ |
| `stream`             | boolean | No       | Whether to stream completion chunks.             |
| `messages`           | array   | Yes      | Chat message list sent to the model.             |
| `messages[].role`    | string  | Yes      | Message role, for example `user` or `assistant`. |
| `messages[].content` | string  | Yes      | Message content.                                 |

### Example response

```json
{
  "model": "brave-pro",
  "system_fingerprint": "string",
  "choices": [
    {
      "delta": {
        "role": "assistant",
        "content": "string"
      },
      "finish_reason": "stop"
    }
  ],
  "created": 1,
  "id": "string",
  "object": "chat.completion.chunk",
  "usage": {
    "completion_tokens": 1,
    "prompt_tokens": 1,
    "total_tokens": 1,
    "completion_tokens_details": {
      "reasoning_tokens": 1
    }
  }
}
```

### Response fields

| Field                                              | Type   | Description                                                   |
| -------------------------------------------------- | ------ | ------------------------------------------------------------- |
| `model`                                            | string | Model used to generate the response, for example `brave-pro`. |
| `system_fingerprint`                               | string | System fingerprint for the completion runtime.                |
| `choices`                                          | array  | Returned completion choices.                                  |
| `choices[].delta`                                  | object | Incremental assistant output payload.                         |
| `choices[].delta.role`                             | string | Role for the returned message, typically `assistant`.         |
| `choices[].delta.content`                          | string | Generated assistant text.                                     |
| `choices[].finish_reason`                          | string | Reason generation stopped, for example `stop`.                |
| `created`                                          | number | Creation timestamp or numeric creation marker.                |
| `id`                                               | string | Unique completion identifier.                                 |
| `object`                                           | string | Response object type, for example `chat.completion.chunk`.    |
| `usage`                                            | object | Token usage metadata.                                         |
| `usage.completion_tokens`                          | number | Number of generated tokens.                                   |
| `usage.prompt_tokens`                              | number | Number of prompt tokens.                                      |
| `usage.total_tokens`                               | number | Total token count.                                            |
| `usage.completion_tokens_details.reasoning_tokens` | number | Count of reasoning-related completion tokens.                 |

### Notes

- This endpoint is designed for AI-generated answers backed by live web search.
- Responses may be returned as chunks when streaming is enabled.
- The response schema can include additional fields beyond the example shown here.
