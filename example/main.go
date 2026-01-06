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

	// Make sure to uncomment the examples you want to run
	// These are commented out to avoid unnecessary API calls during testing
	accountInfoExample(client)
	// Universal examples
	// transcriptExample(client)
	// metadataExample(client)

	// Web examples
	// scrapeExample(client)
	// mapExample(client)
	// crawlExample(client) -- requires using a paid plan

	// YouTube examples
	// youtubeSearchExample(client)
	// youtubeVideoExample(client)
	// youtubeVideoBatchExample(client) -- requires using a paid plan
	// youtubeTranscriptExample(client)
	// youtubeTranscriptBatchExample(client) -- requires using a paid plan
	// youtubeTranscriptTranslateExample(client) -- requires using a paid plan
	// youtubeChannelExample(client)
	// youtubePlaylistExample(client)
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

// accountInfoExample demonstrates the Me endpoint for account information
func accountInfoExample(client *supadata.Supadata) {
	fmt.Println("=== Account Info Example ===")

	info, err := client.Me()
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Organization ID: %s\n", info.OrganizationId)
	fmt.Printf("Plan: %s\n", info.Plan)
	fmt.Printf("Credits: %d / %d used\n", info.UsedCredits, info.MaxCredits)
	fmt.Println()
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

// scrapeExample demonstrates the Scrape endpoint
func scrapeExample(client *supadata.Supadata) {
	fmt.Println("=== Web Scrape Example ===")

	result, err := client.Scrape(&supadata.ScrapeParams{
		Url:  "https://docs.supadata.ai",
		Lang: "en",
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("URL: %s\n", result.Url)
	fmt.Printf("Name: %s\n", result.Name)
	fmt.Printf("Description: %s\n", result.Description)
	fmt.Printf("Characters: %d\n", result.CountCharacters)
	fmt.Printf("Links found: %d\n", len(result.Urls))

	// Print first 100 chars of content
	if len(result.Content) > 100 {
		fmt.Printf("Content preview: %s...\n", result.Content[:100])
	} else {
		fmt.Printf("Content: %s\n", result.Content)
	}
	fmt.Println()
}

// mapExample demonstrates the Map endpoint
func mapExample(client *supadata.Supadata) {
	fmt.Println("=== Web Map Example ===")

	result, err := client.Map(&supadata.MapParams{
		Url: "https://docs.supadata.ai",
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Found %d URLs:\n", len(result.Urls))
	for i, url := range result.Urls {
		if i >= 10 {
			fmt.Printf("  ... and %d more\n", len(result.Urls)-10)
			break
		}
		fmt.Printf("  - %s\n", url)
	}
	fmt.Println()
}

// crawlExample demonstrates the Crawl and CrawlResult endpoints
func crawlExample(client *supadata.Supadata) {
	fmt.Println("=== Web Crawl Example ===")

	// Start crawl job
	job, err := client.Crawl(&supadata.CrawlBody{
		Url:   "https://docs.supadata.ai",
		Limit: 10,
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Crawl job started with ID: %s\n", job.JobId)

	// Poll for results
	for {
		result, err := client.CrawlResult(job.JobId, 0)
		if err != nil {
			handleError(err)
			return
		}

		fmt.Printf("Status: %s\n", result.Status)

		switch result.Status {
		case supadata.CrawlCompleted:
			fmt.Printf("Crawl completed! Found %d pages\n", len(result.Pages))
			for i, page := range result.Pages {
				if i >= 5 {
					fmt.Printf("  ... and %d more pages\n", len(result.Pages)-5)
					break
				}
				fmt.Printf("  - %s (%d chars)\n", page.Name, page.CountCharacters)
			}
			return

		case supadata.CrawlFailed:
			fmt.Println("Crawl job failed")
			return

		case supadata.Cancelled:
			fmt.Println("Crawl job was cancelled")
			return

		case supadata.Scraping:
			fmt.Println("Still crawling, waiting 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}
}

// =============================================================================
// YouTube Examples
// =============================================================================

// youtubeSearchExample demonstrates the YouTube Search endpoint
func youtubeSearchExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Search Example ===")

	result, err := client.YouTubeSearch(&supadata.YouTubeSearchParams{
		Query:      "Never Gonna Give You Up",
		Type:       supadata.SearchTypeVideo,
		UploadDate: supadata.UploadDateMonth,
		SortBy:     supadata.SortByViews,
		Limit:      10,
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Query: %s\n", result.Query)
	fmt.Printf("Total results: %d\n", result.TotalResults)
	fmt.Printf("Results returned: %d\n", len(result.Results))

	for i, item := range result.Results {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(result.Results)-5)
			break
		}
		fmt.Printf("  - [%s] %s\n", item.Type, item.Title)
		if item.ViewCount != nil {
			fmt.Printf("    Views: %d\n", *item.ViewCount)
		}
	}

	if result.NextPageToken != "" {
		fmt.Printf("Next page token: %s\n", result.NextPageToken)
	}
	fmt.Println()
}

// youtubeVideoExample demonstrates the YouTube Video endpoint
func youtubeVideoExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Video Example ===")

	video, err := client.YouTubeVideo("dQw4w9WgXcQ")
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("ID: %s\n", video.Id)
	fmt.Printf("Title: %s\n", video.Title)
	fmt.Printf("Duration: %d seconds\n", video.Duration)
	fmt.Printf("Channel: %s\n", video.Channel.Name)

	if video.ViewCount != nil {
		fmt.Printf("Views: %d\n", *video.ViewCount)
	}
	if video.LikeCount != nil {
		fmt.Printf("Likes: %d\n", *video.LikeCount)
	}

	if len(video.Tags) > 0 {
		fmt.Printf("Tags: %v\n", video.Tags[:min(5, len(video.Tags))])
	}

	fmt.Printf("Available transcript languages: %v\n", video.TranscriptLanguages)
	fmt.Println()
}

// youtubeVideoBatchExample demonstrates the YouTube Video Batch endpoint
func youtubeVideoBatchExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Video Batch Example ===")

	// Start a batch job for multiple videos
	job, err := client.YouTubeVideoBatch(&supadata.YouTubeVideoBatchParams{
		VideoIds: []string{"dQw4w9WgXcQ", "9bZkp7q19f0", "kJQP7kiw5Fk"},
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Batch job started with ID: %s\n", job.JobId)

	// Poll for results
	pollYouTubeBatch(client, job.JobId)
}

// youtubeTranscriptExample demonstrates the YouTube Transcript endpoint
func youtubeTranscriptExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Transcript Example ===")

	result, err := client.YouTubeTranscript(&supadata.YouTubeTranscriptParams{
		VideoId: "dQw4w9WgXcQ",
		Lang:    "en",
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Language: %s\n", result.Lang)
	fmt.Printf("Available languages: %v\n", result.AvailableLangs)
	fmt.Printf("Segments: %d\n", len(result.Content))

	for i, segment := range result.Content {
		if i >= 5 {
			fmt.Println("  ...")
			break
		}
		fmt.Printf("  [%.2fs] %s\n", segment.Offset, segment.Text)
	}
	fmt.Println()
}

// youtubeTranscriptBatchExample demonstrates the YouTube Transcript Batch endpoint
func youtubeTranscriptBatchExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Transcript Batch Example ===")

	// Start a batch job for transcripts from a playlist
	job, err := client.YouTubeTranscriptBatch(&supadata.YouTubeTranscriptBatchParams{
		VideoIds: []string{"dQw4w9WgXcQ", "9bZkp7q19f0"},
		Lang:     "en",
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Transcript batch job started with ID: %s\n", job.JobId)

	// Poll for results
	pollYouTubeBatch(client, job.JobId)
}

// youtubeTranscriptTranslateExample demonstrates the YouTube Transcript Translate endpoint
func youtubeTranscriptTranslateExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Transcript Translate Example ===")

	result, err := client.YouTubeTranscriptTranslate(&supadata.YouTubeTranscriptTranslateParams{
		VideoId: "dQw4w9WgXcQ",
		Lang:    "es", // Translate to Spanish
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Translated to: %s\n", result.Lang)
	fmt.Printf("Segments: %d\n", len(result.Content))

	for i, segment := range result.Content {
		if i >= 5 {
			fmt.Println("  ...")
			break
		}
		fmt.Printf("  [%.2fs] %s\n", segment.Offset, segment.Text)
	}
	fmt.Println()
}

// pollYouTubeBatch demonstrates polling for YouTube batch job results
func pollYouTubeBatch(client *supadata.Supadata, jobId string) {
	fmt.Printf("Polling for batch job: %s\n", jobId)

	for {
		result, err := client.YouTubeBatchResult(jobId)
		if err != nil {
			handleError(err)
			return
		}

		fmt.Printf("Status: %s\n", result.Status)

		switch result.Status {
		case supadata.BatchCompleted:
			fmt.Printf("Batch completed!\n")
			fmt.Printf("Stats: %d total, %d succeeded, %d failed\n",
				result.Stats.Total, result.Stats.Succeeded, result.Stats.Failed)

			for i, item := range result.Results {
				if i >= 5 {
					fmt.Printf("  ... and %d more\n", len(result.Results)-5)
					break
				}
				if item.ErrorCode != "" {
					fmt.Printf("  - %s: ERROR (%s)\n", item.VideoId, item.ErrorCode)
				} else if item.Video != nil {
					fmt.Printf("  - %s: %s\n", item.VideoId, item.Video.Title)
				} else if item.Transcript != nil {
					fmt.Printf("  - %s: %d segments\n", item.VideoId, len(item.Transcript.Content))
				}
			}
			return

		case supadata.BatchFailed:
			fmt.Println("Batch job failed")
			return

		case supadata.BatchQueued, supadata.BatchActive:
			fmt.Println("Still processing, waiting 2 seconds...")
			time.Sleep(2 * time.Second)
		}
	}
}

// youtubeChannelExample demonstrates the YouTube Channel endpoint
func youtubeChannelExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Channel Example ===")

	channel, err := client.YouTubeChannel("@GoogleDevelopers")
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("ID: %s\n", channel.Id)
	fmt.Printf("Name: %s\n", channel.Name)
	if channel.Description != "" {
		fmt.Printf("Description: %.100s...\n", channel.Description)
	}

	if channel.SubscriberCount != nil {
		fmt.Printf("Subscribers: %d\n", *channel.SubscriberCount)
	}
	if channel.VideoCount != nil {
		fmt.Printf("Videos: %d\n", *channel.VideoCount)
	}
	if channel.ViewCount != nil {
		fmt.Printf("Total Views: %d\n", *channel.ViewCount)
	}

	// Get channel videos
	videos, err := client.YouTubeChannelVideos(&supadata.YouTubeChannelVideosParams{
		Id:    channel.Id,
		Limit: 10,
		Type:  supadata.ChannelVideoTypeAll,
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Recent videos: %d\n", len(videos.VideoIds))
	fmt.Printf("Recent shorts: %d\n", len(videos.ShortIds))
	fmt.Printf("Recent live streams: %d\n", len(videos.LiveIds))
	fmt.Println()
}

// youtubePlaylistExample demonstrates the YouTube Playlist endpoint
func youtubePlaylistExample(client *supadata.Supadata) {
	fmt.Println("=== YouTube Playlist Example ===")

	// Use a known public playlist ID
	playlist, err := client.YouTubePlaylist("PLj6h78yzYM2N8nw1YcqqKveySH6_0VnI0")
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("ID: %s\n", playlist.Id)
	fmt.Printf("Title: %s\n", playlist.Title)
	fmt.Printf("Video Count: %d\n", playlist.VideoCount)
	fmt.Printf("Channel: %s\n", playlist.Channel.Name)

	if playlist.ViewCount != nil {
		fmt.Printf("Views: %d\n", *playlist.ViewCount)
	}

	// Get playlist videos
	videos, err := client.YouTubePlaylistVideos(&supadata.YouTubePlaylistVideosParams{
		Id:    playlist.Id,
		Limit: 20,
	})
	if err != nil {
		handleError(err)
		return
	}

	fmt.Printf("Video IDs in playlist: %d\n", len(videos.VideoIds))
	for i, videoId := range videos.VideoIds {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(videos.VideoIds)-5)
			break
		}
		fmt.Printf("  - %s\n", videoId)
	}
	fmt.Println()
}
