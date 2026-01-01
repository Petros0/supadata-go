# supadata-go

[![Go Reference](https://pkg.go.dev/badge/github.com/petros0/supadata-go.svg)](https://pkg.go.dev/github.com/petros0/supadata-go)
[![CI](https://github.com/petros0/supadata-go/actions/workflows/ci.yml/badge.svg)](https://github.com/petros0/supadata-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go SDK for the [Supadata API](https://supadata.ai) - extract transcripts from YouTube, Instagram, and other video platforms.

**This is an UNOFFICIAL library and is not affiliated with https://supadata.ai/**

## Installation

```bash
go get github.com/petros0/supadata-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "os"

    supadata "github.com/petros0/supadata-go"
)

func main() {
    client := supadata.NewSupadata(&supadata.SupadataConfig{
        APIKey: os.Getenv("SUPADATA_API_KEY"),
    })

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

The project uses https://asdf-vm.com/guide/getting-started.html for version management. To set up the development environment, run:

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
client := supadata.NewSupadata(nil) // Will use SUPADATA_API_KEY env var
```

2. **Explicit configuration**:
```go
client := supadata.NewSupadata(&supadata.SupadataConfig{
    APIKey: "your_api_key",
})
```

### Custom HTTP Client

You can provide a custom HTTP client for advanced use cases:

```go
client := supadata.NewSupadata(&supadata.SupadataConfig{
    APIKey: os.Getenv("SUPADATA_API_KEY"),
    Client: &http.Client{
        Timeout: 30 * time.Second,
    },
})
```

## License

MIT License - see [LICENSE](LICENSE) for details.
