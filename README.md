# supadata-go

[![Go Reference](https://pkg.go.dev/badge/github.com/petros0/supadata-go.svg)](https://pkg.go.dev/github.com/petros0/supadata-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/petros0/supadata-go)](https://goreportcard.com/report/github.com/petros0/supadata-go)
[![CI](https://github.com/petros0/supadata-go/actions/workflows/ci.yml/badge.svg)](https://github.com/petros0/supadata-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go SDK for the [Supadata API](https://supadata.ai) - extract transcripts from YouTube, Instagram, and other video
platforms.

**This is an UNOFFICIAL library and is not affiliated with https://supadata.ai/**

## Supported Features

OpenAPI specification: https://supadata.ai/api/v1/openapi.json

| Endpoint                                                                                                 | Supported |
|----------------------------------------------------------------------------------------------------------|-----------|
| [Universal Transcript](https://docs.supadata.ai/api-reference/endpoint/transcript/transcript)            | ✅         |
| [Universal Transcript Result](https://docs.supadata.ai/api-reference/endpoint/transcript/transcript-get) | ✅         |
| [Universal Metadata](https://docs.supadata.ai/api-reference/endpoint/metadata/metadata)                  | ✅         |
| [Youtube Search](https://docs.supadata.ai/api-reference/endpoint/youtube/search)                         | ❌         |
| [Youtube Video](https://docs.supadata.ai/api-reference/endpoint/youtube/video-get)                       | ❌         |
| [Youtube Video Batch](https://docs.supadata.ai/api-reference/endpoint/youtube/video-batch)               | ❌         |
| [Youtube Transcript](https://docs.supadata.ai/api-reference/endpoint/youtube/transcript)                 | ❌         |
| [Youtube Transcript](https://docs.supadata.ai/api-reference/endpoint/youtube/transcript)                 | ❌         |
| [Youtube Transcript Batch](https://docs.supadata.ai/api-reference/endpoint/youtube/transcript-batch)     | ❌         |
| [Youtube Transcript Translate](https://docs.supadata.ai/api-reference/endpoint/youtube/translation)      | ❌         |
| [Youtube Channel](https://docs.supadata.ai/api-reference/endpoint/youtube/channel)                       | ❌         |
| [Youtube Playlist](https://docs.supadata.ai/api-reference/endpoint/youtube/playlist)                     | ❌         |
| [Youtube Channel Videos](https://docs.supadata.ai/api-reference/endpoint/youtube/channel-videos)         | ❌         |
| [Youtube Playlist Videos](https://docs.supadata.ai/api-reference/endpoint/youtube/playlist-videos)       | ❌         |
| [Youtube Batch Result](https://docs.supadata.ai/api-reference/endpoint/youtube/batch-get)                | ❌         |
| [Web Scrape](https://docs.supadata.ai/api-reference/endpoint/web/scrape)                                 | ❌         |
| [Web Map](https://docs.supadata.ai/api-reference/endpoint/web/map)                                       | ❌         |
| [Web Crawl](https://docs.supadata.ai/api-reference/endpoint/web/crawl)                                   | ❌         |
| [Web Crawl Status](https://docs.supadata.ai/api-reference/endpoint/web/crawl-get)                        | ❌         |
| [Account Information](https://docs.supadata.ai/api-reference/endpoint/account/me)                        | ❌         |

## Installation

```bash
go get github.com/petros0/supadata-go
```

## Quick Start

```go
package main

import (
	"fmt"

	supadata "github.com/petros0/supadata-go"
)

func main() {
	client := supadata.NewSupadata() // Will use SUPADATA_API_KEY env var

	transcript, err := client.Transcript(&supadata.TranscriptParams{
		Url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
	})
	if err != nil {
		panic(err)
	}

	if transcript.IsAsync() {
		fmt.Println("Job ID:", transcript.Async.JobId)
	} else {
		for _, segment := range transcript.Sync.Content {
			fmt.Println(segment.Text)
		}
	}
}
```

## Development

The project uses https://asdf-vm.com/guide/getting-started.html for version management. To set up the development
environment, run:

```bash
asdf install
```

## Configuration

### API Key

You can provide your API key in two ways:

1. **Environment variable** (recommended):

```bash
export SUPADATA_API_KEY=your_api_key
```

```go
client := supadata.NewSupadata() // Will use SUPADATA_API_KEY env var
```

2. **Explicit configuration**:

```go
client := supadata.NewSupadata(
    WithAPIKey("sd_..."),
    WithTimeout(30*time.Second), // just override timeout of the client
)
```

### Custom HTTP client

You can provide a custom HTTP client for advanced use cases:

```go
client := supadata.NewSupadata(
	WithAPIKey("sd_..."),
	WithClient(
		&http.Client{
            Timeout:   30 * time.Second,
            Transport: http.DefaultTransport,
        },
    ),
)
```

## License

MIT License - see [LICENSE](LICENSE) for details.
