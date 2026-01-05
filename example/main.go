package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/petros0/supadata-go"
)

func main() {
	// Create a new Supadata client using environment variable
	// Set SUPADATA_API_KEY before running: export SUPADATA_API_KEY=your_api_key
	client := supadata.NewSupadata(
		// WithAPIKey can override the environment variable if needed
		// supadata.WithAPIKey("os.Getenv("YOUR_KEY_VARIABLE")"), //

		// Optional: Set custom timeout (default is 60 seconds)
		supadata.WithTimeout(30 * time.Second),

		// Optional: Use custom base URL for testing
		// supadata.WithBaseURL("https://custom.api.example.com/v1"),

		// Optional: Use custom HTTP client
		// supadata.WithClient(&http.Client{
		// 	Timeout: 15 * time.Second,
		// }),
	)

	if os.Getenv("SUPADATA_API_KEY") == "" {
		log.Fatal("SUPADATA_API_KEY environment variable is required")
	}

	// Run examples
	transcriptExample(client)
	metadataExample(client)
}

// handleError demonstrates proper error handling for Supadata API errors
func handleError(err error) {
	var apiErr *supadata.ErrorResponse
	if errors.As(err, &apiErr) {
		// Handle specific API errors
		fmt.Printf("API Error: %s\n", apiErr.ErrorIdentifier)
		fmt.Printf("Message: %s\n", apiErr.Message)
		if apiErr.Details != "" {
			fmt.Printf("Details: %s\n", apiErr.Details)
		}
		if apiErr.DocumentationUrl != "" {
			fmt.Printf("Documentation: %s\n", apiErr.DocumentationUrl)
		}

		// Handle specific error types
		switch apiErr.ErrorIdentifier {
		case supadata.Unauthorized:
			fmt.Println("Check your API key")
		case supadata.LimitExceeded:
			fmt.Println("Rate limit exceeded, please wait before retrying")
		case supadata.TranscriptUnavailable:
			fmt.Println("Transcript not available for this content")
		case supadata.NotFound:
			fmt.Println("Resource not found")
		}
	} else {
		// Handle non-API errors (network issues, etc.)
		fmt.Printf("Error: %v\n", err)
	}
}

// transcriptExample demonstrates the Transcript endpoint
func transcriptExample(client *supadata.Supadata) {
	fmt.Println("=== Transcript Example ===")

	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	transcript, err := client.Transcript(&supadata.TranscriptParams{
		Url:  url,
		Lang: "en",
		Mode: supadata.Auto,
	})
	if err != nil {
		handleError(err)
		return
	}

	// Check if response is async or sync
	if transcript.IsAsync() {
		fmt.Printf("Async job started with ID: %s\n", transcript.Async.JobId)
		// Poll for results
		pollTranscript(client, transcript.Async.JobId)
	} else {
		fmt.Printf("Language: %s\n", transcript.Sync.Lang)
		fmt.Printf("Available languages: %v\n", transcript.Sync.AvailableLangs)
		fmt.Printf("Segments: %d\n", len(transcript.Sync.Content))

		// Print first few segments
		for i, segment := range transcript.Sync.Content {
			if i >= 3 {
				fmt.Println("...")
				break
			}
			fmt.Printf("  [%.2fs] %s\n", segment.Offset, segment.Text)
		}
	}
	fmt.Println()
}

// pollTranscript demonstrates the TranscriptResult endpoint for async jobs
func pollTranscript(client *supadata.Supadata, jobId string) {
	fmt.Printf("Polling for transcript job: %s\n", jobId)

	for {
		result, err := client.TranscriptResult(jobId)
		if err != nil {
			handleError(err)
			return
		}

		fmt.Printf("Status: %s\n", result.Status)

		switch result.Status {
		case supadata.Completed:
			fmt.Printf("Transcript completed!\n")
			fmt.Printf("Language: %s\n", result.Lang)
			fmt.Printf("Segments: %d\n", len(result.Content))
			for i, segment := range result.Content {
				if i >= 3 {
					fmt.Println("...")
					break
				}
				fmt.Printf("  [%.2fs] %s\n", segment.Offset, segment.Text)
			}
			return

		case supadata.Failed:
			fmt.Println("Transcript job failed")
			if result.Error != nil {
				fmt.Printf("Error: %s - %s\n", result.Error.ErrorIdentifier, result.Error.Message)
			}
			return

		case supadata.Queued, supadata.Active:
			fmt.Println("Still processing, waiting 2 seconds...")
			time.Sleep(2 * time.Second)
		}
	}
}

// metadataExample demonstrates the Metadata endpoint
func metadataExample(client *supadata.Supadata) {
	fmt.Println("=== Metadata Example ===")

	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	metadata, err := client.Metadata(url)
	if err != nil {
		handleError(err)
		return
	}

	// Basic info
	fmt.Printf("Platform: %s\n", metadata.Platform)
	fmt.Printf("Type: %s\n", metadata.Type)
	fmt.Printf("ID: %s\n", metadata.Id)
	fmt.Printf("Title: %s\n", metadata.Title)
	fmt.Printf("Description: %.100s...\n", metadata.Description)

	// Author info
	fmt.Printf("Author: %s (@%s)\n", metadata.Author.DisplayName, metadata.Author.Username)
	if metadata.Author.Verified {
		fmt.Println("  (Verified)")
	}

	// Stats (nullable fields)
	fmt.Println("Stats:")
	if metadata.Stats.Views != nil {
		fmt.Printf("  Views: %d\n", *metadata.Stats.Views)
	}
	if metadata.Stats.Likes != nil {
		fmt.Printf("  Likes: %d\n", *metadata.Stats.Likes)
	}
	if metadata.Stats.Comments != nil {
		fmt.Printf("  Comments: %d\n", *metadata.Stats.Comments)
	}

	// Media info
	fmt.Printf("Media Type: %s\n", metadata.Media.Type)
	if metadata.Media.Duration > 0 {
		fmt.Printf("Duration: %.0f seconds\n", metadata.Media.Duration)
	}

	// Tags
	if len(metadata.Tags) > 0 {
		fmt.Printf("Tags: %v\n", metadata.Tags)
	}

	// Created at
	fmt.Printf("Created: %s\n", metadata.CreatedAt.Format(time.RFC3339))
	fmt.Println()
}
