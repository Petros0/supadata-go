package supadata

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Test helpers
func newTestClient(server *httptest.Server) *Supadata {
	return NewSupadata(
		WithAPIKey("test-api-key"),
		WithBaseURL(server.URL),
	)
}

func jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func errorResponse(w http.ResponseWriter, status int, errID ErrorIdentifier, message, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":   string(errID),
		"message": message,
		"details": details,
	})
}

// =============================================================================
// Constructor & Configuration Tests
// =============================================================================

func TestNewSupadata_DefaultConfig(t *testing.T) {
	// Clear env var for this test
	originalKey := os.Getenv("SUPADATA_API_KEY")
	_ = os.Setenv("SUPADATA_API_KEY", "env-api-key")
	defer func() { _ = os.Setenv("SUPADATA_API_KEY", originalKey) }()

	client := NewSupadata()

	if client.config.apiKey != "env-api-key" {
		t.Errorf("expected apiKey from env, got %q", client.config.apiKey)
	}
	if client.config.baseURL != BaseUrl {
		t.Errorf("expected baseURL %q, got %q", BaseUrl, client.config.baseURL)
	}
	if client.config.client.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", client.config.client.Timeout)
	}
}

func TestNewSupadata_WithAPIKey(t *testing.T) {
	client := NewSupadata(WithAPIKey("custom-key"))

	if client.config.apiKey != "custom-key" {
		t.Errorf("expected apiKey %q, got %q", "custom-key", client.config.apiKey)
	}
}

func TestNewSupadata_WithTimeout(t *testing.T) {
	client := NewSupadata(WithTimeout(30 * time.Second))

	if client.config.client.Timeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", client.config.client.Timeout)
	}
}

func TestNewSupadata_WithClient(t *testing.T) {
	customClient := &http.Client{Timeout: 10 * time.Second}
	client := NewSupadata(WithClient(customClient))

	if client.config.client != customClient {
		t.Error("expected custom client to be used")
	}
}

func TestNewSupadata_WithBaseURL(t *testing.T) {
	client := NewSupadata(WithBaseURL("https://custom.api.com"))

	if client.config.baseURL != "https://custom.api.com" {
		t.Errorf("expected baseURL %q, got %q", "https://custom.api.com", client.config.baseURL)
	}
}

func TestNewSupadata_MultipleOptions(t *testing.T) {
	client := NewSupadata(
		WithAPIKey("multi-key"),
		WithTimeout(45*time.Second),
		WithBaseURL("https://multi.api.com"),
	)

	if client.config.apiKey != "multi-key" {
		t.Errorf("expected apiKey %q, got %q", "multi-key", client.config.apiKey)
	}
	if client.config.client.Timeout != 45*time.Second {
		t.Errorf("expected timeout 45s, got %v", client.config.client.Timeout)
	}
	if client.config.baseURL != "https://multi.api.com" {
		t.Errorf("expected baseURL %q, got %q", "https://multi.api.com", client.config.baseURL)
	}
}

// =============================================================================
// Request Building Tests
// =============================================================================

func TestRequest_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if got := r.Header.Get("x-api-key"); got != "test-api-key" {
			t.Errorf("expected x-api-key %q, got %q", "test-api-key", got)
		}
		if got := r.Header.Get("User-Agent"); got != "supadata-go/1.0.0" {
			t.Errorf("expected User-Agent %q, got %q", "supadata-go/1.0.0", got)
		}
		jsonResponse(w, http.StatusOK, map[string]any{
			"content": []any{},
			"lang":    "en",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	_, _ = client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})
}

func TestRequest_QueryParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		// Verify all query params are properly encoded
		if got := q.Get("url"); got != "https://youtube.com/watch?v=test&foo=bar" {
			t.Errorf("expected url with special chars, got %q", got)
		}
		if got := q.Get("lang"); got != "es" {
			t.Errorf("expected lang %q, got %q", "es", got)
		}
		if got := q.Get("text"); got != "true" {
			t.Errorf("expected text %q, got %q", "true", got)
		}
		if got := q.Get("chunkSize"); got != "500" {
			t.Errorf("expected chunkSize %q, got %q", "500", got)
		}
		if got := q.Get("mode"); got != "generate" {
			t.Errorf("expected mode %q, got %q", "generate", got)
		}

		jsonResponse(w, http.StatusOK, map[string]any{"content": []any{}, "lang": "es"})
	}))
	defer server.Close()

	client := newTestClient(server)
	_, _ = client.Transcript(&TranscriptParams{
		Url:       "https://youtube.com/watch?v=test&foo=bar",
		Lang:      "es",
		Text:      true,
		ChunkSize: 500,
		Mode:      Generate,
	})
}

// =============================================================================
// Transcript Method Tests - Success Cases
// =============================================================================

func TestTranscript_SyncResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transcript" {
			t.Errorf("expected path /transcript, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		jsonResponse(w, http.StatusOK, map[string]any{
			"content": []map[string]any{
				{"text": "Hello world", "offset": 0.0, "duration": 1000},
				{"text": "How are you", "offset": 1.0, "duration": 1500},
			},
			"lang":           "en",
			"availableLangs": []string{"en", "es", "fr"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsAsync() {
		t.Error("expected sync response, got async")
	}
	if result.Sync == nil {
		t.Fatal("expected Sync to be non-nil")
	}
	if len(result.Sync.Content) != 2 {
		t.Errorf("expected 2 content items, got %d", len(result.Sync.Content))
	}
	if result.Sync.Content[0].Text != "Hello world" {
		t.Errorf("expected first text %q, got %q", "Hello world", result.Sync.Content[0].Text)
	}
	if result.Sync.Lang != "en" {
		t.Errorf("expected lang %q, got %q", "en", result.Sync.Lang)
	}
	if len(result.Sync.AvailableLangs) != 3 {
		t.Errorf("expected 3 available langs, got %d", len(result.Sync.AvailableLangs))
	}
}

func TestTranscript_AsyncResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"jobId": "job-abc-123",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsAsync() {
		t.Error("expected async response, got sync")
	}
	if result.Async == nil {
		t.Fatal("expected Async to be non-nil")
	}
	if result.Async.JobId != "job-abc-123" {
		t.Errorf("expected jobId %q, got %q", "job-abc-123", result.Async.JobId)
	}
}

func TestTranscript_MinimalParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		// Only url and default mode should be set
		if got := q.Get("url"); got != "https://youtube.com/watch?v=123" {
			t.Errorf("expected url param, got %q", got)
		}
		if got := q.Get("mode"); got != "auto" {
			t.Errorf("expected default mode 'auto', got %q", got)
		}
		// These should be empty
		if got := q.Get("lang"); got != "" {
			t.Errorf("expected empty lang, got %q", got)
		}
		if got := q.Get("text"); got != "" {
			t.Errorf("expected empty text, got %q", got)
		}
		if got := q.Get("chunkSize"); got != "" {
			t.Errorf("expected empty chunkSize, got %q", got)
		}

		jsonResponse(w, http.StatusOK, map[string]any{"content": []any{}, "lang": "en"})
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTranscript_AllModeParams(t *testing.T) {
	modes := []TranscriptModeParam{Native, Auto, Generate}

	for _, mode := range modes {
		t.Run(string(mode), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("mode"); got != string(mode) {
					t.Errorf("expected mode %q, got %q", mode, got)
				}
				jsonResponse(w, http.StatusOK, map[string]any{"content": []any{}, "lang": "en"})
			}))
			defer server.Close()

			client := newTestClient(server)
			_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123", Mode: mode})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// =============================================================================
// Transcript Method Tests - Error Cases
// =============================================================================

func TestTranscript_InvalidRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusBadRequest, InvalidRequest, "Invalid URL format", "URL must be a valid video URL")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "invalid"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != InvalidRequest {
		t.Errorf("expected error %q, got %q", InvalidRequest, errResp.ErrorIdentifier)
	}
}

func TestTranscript_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusUnauthorized, Unauthorized, "Invalid API key", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != Unauthorized {
		t.Errorf("expected error %q, got %q", Unauthorized, errResp.ErrorIdentifier)
	}
}

func TestTranscript_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusForbidden, Forbidden, "Access denied", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != Forbidden {
		t.Errorf("expected error %q, got %q", Forbidden, errResp.ErrorIdentifier)
	}
}

func TestTranscript_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, NotFound, "Video not found", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=notfound"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != NotFound {
		t.Errorf("expected error %q, got %q", NotFound, errResp.ErrorIdentifier)
	}
}

func TestTranscript_TranscriptUnavailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, TranscriptUnavailable, "No transcript available", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != TranscriptUnavailable {
		t.Errorf("expected error %q, got %q", TranscriptUnavailable, errResp.ErrorIdentifier)
	}
}

func TestTranscript_LimitExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusTooManyRequests, LimitExceeded, "Rate limit exceeded", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != LimitExceeded {
		t.Errorf("expected error %q, got %q", LimitExceeded, errResp.ErrorIdentifier)
	}
}

func TestTranscript_InternalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusInternalServerError, InternalError, "Internal server error", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != InternalError {
		t.Errorf("expected error %q, got %q", InternalError, errResp.ErrorIdentifier)
	}
}

func TestTranscript_UpgradeRequired(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusPaymentRequired, UpgradeRequired, "Plan upgrade required", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != UpgradeRequired {
		t.Errorf("expected error %q, got %q", UpgradeRequired, errResp.ErrorIdentifier)
	}
}

func TestTranscript_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{invalid json"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestTranscript_NonJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte("Bad Gateway"))
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Transcript(&TranscriptParams{Url: "https://youtube.com/watch?v=123"})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Should get a generic error since body isn't valid JSON
	if err.Error() != "request failed with status 502" {
		t.Errorf("expected generic error message, got %q", err.Error())
	}
}

// =============================================================================
// TranscriptResult Method Tests
// =============================================================================

func TestTranscriptResult_Queued(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/transcript/job-123" {
			t.Errorf("expected path /transcript/job-123, got %s", r.URL.Path)
		}
		jsonResponse(w, http.StatusOK, map[string]any{
			"status": "queued",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.TranscriptResult("job-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != Queued {
		t.Errorf("expected status %q, got %q", Queued, result.Status)
	}
}

func TestTranscriptResult_Active(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"status": "active",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.TranscriptResult("job-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != Active {
		t.Errorf("expected status %q, got %q", Active, result.Status)
	}
}

func TestTranscriptResult_Completed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"status": "completed",
			"content": []map[string]any{
				{"text": "Transcript content", "offset": 0.0, "duration": 1000},
			},
			"lang":           "en",
			"availableLangs": []string{"en", "es"},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.TranscriptResult("job-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != Completed {
		t.Errorf("expected status %q, got %q", Completed, result.Status)
	}
	if len(result.Content) != 1 {
		t.Errorf("expected 1 content item, got %d", len(result.Content))
	}
	if result.Lang != "en" {
		t.Errorf("expected lang %q, got %q", "en", result.Lang)
	}
}

func TestTranscriptResult_Failed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"status": "failed",
			"error": map[string]any{
				"error":   "transcript-unavailable",
				"message": "Could not generate transcript",
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.TranscriptResult("job-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != Failed {
		t.Errorf("expected status %q, got %q", Failed, result.Status)
	}
	if result.Error == nil {
		t.Fatal("expected error info, got nil")
	}
	if result.Error.ErrorIdentifier != TranscriptUnavailable {
		t.Errorf("expected error identifier %q, got %q", TranscriptUnavailable, result.Error.ErrorIdentifier)
	}
}

func TestTranscriptResult_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, NotFound, "Job not found", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.TranscriptResult("invalid-job")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != NotFound {
		t.Errorf("expected error %q, got %q", NotFound, errResp.ErrorIdentifier)
	}
}

func TestTranscriptResult_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusUnauthorized, Unauthorized, "Invalid API key", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.TranscriptResult("job-123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != Unauthorized {
		t.Errorf("expected error %q, got %q", Unauthorized, errResp.ErrorIdentifier)
	}
}

// =============================================================================
// Metadata Method Tests
// =============================================================================

func TestMetadata_YouTube(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/metadata" {
			t.Errorf("expected path /metadata, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("url"); got != "https://youtube.com/watch?v=123" {
			t.Errorf("expected url query param, got %q", got)
		}

		jsonResponse(w, http.StatusOK, map[string]any{
			"platform":    "youtube",
			"type":        "video",
			"id":          "123",
			"url":         "https://youtube.com/watch?v=123",
			"title":       "Test Video",
			"description": "A test video",
			"author": map[string]any{
				"displayName": "Test Channel",
				"username":    "testchannel",
				"avatarUrl":   "https://example.com/avatar.jpg",
				"verified":    true,
			},
			"stats": map[string]any{
				"likes":    1000,
				"comments": 50,
				"views":    10000,
			},
			"media": map[string]any{
				"type":         "video",
				"duration":     120.5,
				"thumbnailUrl": "https://example.com/thumb.jpg",
			},
			"tags":      []string{"test", "video"},
			"createdAt": "2024-01-15T10:30:00Z",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.Metadata("https://youtube.com/watch?v=123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Platform != YouTube {
		t.Errorf("expected platform %q, got %q", YouTube, result.Platform)
	}
	if result.Type != Video {
		t.Errorf("expected type %q, got %q", Video, result.Type)
	}
	if result.Title != "Test Video" {
		t.Errorf("expected title %q, got %q", "Test Video", result.Title)
	}
	if result.Author.DisplayName != "Test Channel" {
		t.Errorf("expected author name %q, got %q", "Test Channel", result.Author.DisplayName)
	}
	if !result.Author.Verified {
		t.Error("expected author to be verified")
	}
	if result.Stats.Views == nil || *result.Stats.Views != 10000 {
		t.Errorf("expected views 10000, got %v", result.Stats.Views)
	}
}

func TestMetadata_AllPlatforms(t *testing.T) {
	platforms := []struct {
		url      string
		platform MetadataPlatform
	}{
		{"https://youtube.com/watch?v=123", YouTube},
		{"https://tiktok.com/@user/video/123", TikTok},
		{"https://instagram.com/p/abc123", Instagram},
		{"https://twitter.com/user/status/123", Twitter},
		{"https://facebook.com/video/123", Facebook},
	}

	for _, tc := range platforms {
		t.Run(string(tc.platform), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, http.StatusOK, map[string]any{
					"platform":    string(tc.platform),
					"type":        "video",
					"id":          "123",
					"url":         tc.url,
					"title":       "Test",
					"description": "",
					"author":      map[string]any{},
					"stats":       map[string]any{},
					"media":       map[string]any{"type": "video"},
					"createdAt":   "2024-01-15T10:30:00Z",
				})
			}))
			defer server.Close()

			client := newTestClient(server)
			result, err := client.Metadata(tc.url)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Platform != tc.platform {
				t.Errorf("expected platform %q, got %q", tc.platform, result.Platform)
			}
		})
	}
}

func TestMetadata_AllTypes(t *testing.T) {
	types := []MetadataType{Video, Image, Carousel, Post}

	for _, mediaType := range types {
		t.Run(string(mediaType), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				jsonResponse(w, http.StatusOK, map[string]any{
					"platform":    "instagram",
					"type":        string(mediaType),
					"id":          "123",
					"url":         "https://instagram.com/p/123",
					"title":       "Test",
					"description": "",
					"author":      map[string]any{},
					"stats":       map[string]any{},
					"media":       map[string]any{"type": string(mediaType)},
					"createdAt":   "2024-01-15T10:30:00Z",
				})
			}))
			defer server.Close()

			client := newTestClient(server)
			result, err := client.Metadata("https://instagram.com/p/123")

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Type != mediaType {
				t.Errorf("expected type %q, got %q", mediaType, result.Type)
			}
		})
	}
}

func TestMetadata_CarouselWithItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"platform":    "instagram",
			"type":        "carousel",
			"id":          "123",
			"url":         "https://instagram.com/p/123",
			"title":       "Carousel Post",
			"description": "",
			"author":      map[string]any{},
			"stats":       map[string]any{},
			"media": map[string]any{
				"type": "carousel",
				"items": []map[string]any{
					{"type": "image", "url": "https://example.com/1.jpg"},
					{"type": "video", "url": "https://example.com/2.mp4", "duration": 30.0},
				},
			},
			"createdAt": "2024-01-15T10:30:00Z",
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.Metadata("https://instagram.com/p/123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Media.Items) != 2 {
		t.Errorf("expected 2 media items, got %d", len(result.Media.Items))
	}
}

func TestMetadata_WithAdditionalData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w, http.StatusOK, map[string]any{
			"platform":    "youtube",
			"type":        "video",
			"id":          "123",
			"url":         "https://youtube.com/watch?v=123",
			"title":       "Test",
			"description": "",
			"author":      map[string]any{},
			"stats":       map[string]any{},
			"media":       map[string]any{"type": "video"},
			"createdAt":   "2024-01-15T10:30:00Z",
			"additionalData": map[string]any{
				"customField": "customValue",
				"nested": map[string]any{
					"key": "value",
				},
			},
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.Metadata("https://youtube.com/watch?v=123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AdditionalData == nil {
		t.Fatal("expected additionalData, got nil")
	}
	if result.AdditionalData["customField"] != "customValue" {
		t.Errorf("expected customField value, got %v", result.AdditionalData["customField"])
	}
}

func TestMetadata_InvalidURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusBadRequest, InvalidRequest, "Invalid URL", "URL must be from a supported platform")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Metadata("invalid-url")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != InvalidRequest {
		t.Errorf("expected error %q, got %q", InvalidRequest, errResp.ErrorIdentifier)
	}
}

func TestMetadata_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusNotFound, NotFound, "Content not found", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Metadata("https://youtube.com/watch?v=deleted")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != NotFound {
		t.Errorf("expected error %q, got %q", NotFound, errResp.ErrorIdentifier)
	}
}

func TestMetadata_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusUnauthorized, Unauthorized, "Invalid API key", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Metadata("https://youtube.com/watch?v=123")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != Unauthorized {
		t.Errorf("expected error %q, got %q", Unauthorized, errResp.ErrorIdentifier)
	}
}

// =============================================================================
// Error Response Tests
// =============================================================================

func TestErrorResponse_Error(t *testing.T) {
	err := &ErrorResponse{
		ErrorIdentifier: InvalidRequest,
		Message:         "Test error message",
		Details:         "Some details",
	}

	expected := "invalid-request: Test error message"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}

func TestErrorResponse_AllIdentifiers(t *testing.T) {
	identifiers := []ErrorIdentifier{
		InvalidRequest,
		InternalError,
		Forbidden,
		Unauthorized,
		UpgradeRequired,
		TranscriptUnavailable,
		NotFound,
		LimitExceeded,
	}

	for _, id := range identifiers {
		t.Run(string(id), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				errorResponse(w, http.StatusBadRequest, id, "Test message", "")
			}))
			defer server.Close()

			client := newTestClient(server)
			_, err := client.Metadata("https://youtube.com/watch?v=123")

			if err == nil {
				t.Fatal("expected error, got nil")
			}
			errResp, ok := err.(*ErrorResponse)
			if !ok {
				t.Fatalf("expected *ErrorResponse, got %T", err)
			}
			if errResp.ErrorIdentifier != id {
				t.Errorf("expected error %q, got %q", id, errResp.ErrorIdentifier)
			}
		})
	}
}

// =============================================================================
// Union Type Tests
// =============================================================================

func TestTranscript_IsAsync_True(t *testing.T) {
	transcript := &Transcript{
		Async: &AsyncTranscript{JobId: "job-123"},
	}

	if !transcript.IsAsync() {
		t.Error("expected IsAsync() to return true")
	}
}

func TestTranscript_IsAsync_False(t *testing.T) {
	transcript := &Transcript{
		Sync: &SyncTranscript{
			Content: []TranscriptContent{},
			Lang:    "en",
		},
	}

	if transcript.IsAsync() {
		t.Error("expected IsAsync() to return false")
	}
}

func TestTranscript_SyncFields(t *testing.T) {
	transcript := &Transcript{
		Sync: &SyncTranscript{
			Content: []TranscriptContent{
				{Text: "Hello", Offset: 0, Duration: 1000},
			},
			Lang:           "en",
			AvailableLangs: []string{"en", "es"},
		},
	}

	if transcript.Sync.Lang != "en" {
		t.Errorf("expected lang %q, got %q", "en", transcript.Sync.Lang)
	}
	if len(transcript.Sync.Content) != 1 {
		t.Errorf("expected 1 content item, got %d", len(transcript.Sync.Content))
	}
}

func TestTranscript_AsyncFields(t *testing.T) {
	transcript := &Transcript{
		Async: &AsyncTranscript{JobId: "job-abc-123"},
	}

	if transcript.Async.JobId != "job-abc-123" {
		t.Errorf("expected jobId %q, got %q", "job-abc-123", transcript.Async.JobId)
	}
}

// =============================================================================
// Me (Account Info) Method Tests
// =============================================================================

func TestMe_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me" {
			t.Errorf("expected path /me, got %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		jsonResponse(w, http.StatusOK, map[string]any{
			"organizationId": "550e8400-e29b-41d4-a716-446655440000",
			"plan":           "Pro",
			"maxCredits":     100000,
			"usedCredits":    15000,
		})
	}))
	defer server.Close()

	client := newTestClient(server)
	result, err := client.Me()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OrganizationId != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected organizationId %q, got %q", "550e8400-e29b-41d4-a716-446655440000", result.OrganizationId)
	}
	if result.Plan != "Pro" {
		t.Errorf("expected plan %q, got %q", "Pro", result.Plan)
	}
	if result.MaxCredits != 100000 {
		t.Errorf("expected maxCredits %d, got %d", 100000, result.MaxCredits)
	}
	if result.UsedCredits != 15000 {
		t.Errorf("expected usedCredits %d, got %d", 15000, result.UsedCredits)
	}
}

func TestMe_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusUnauthorized, Unauthorized, "Invalid API key", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Me()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != Unauthorized {
		t.Errorf("expected error %q, got %q", Unauthorized, errResp.ErrorIdentifier)
	}
}

func TestMe_InternalError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		errorResponse(w, http.StatusInternalServerError, InternalError, "Internal server error", "")
	}))
	defer server.Close()

	client := newTestClient(server)
	_, err := client.Me()

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	errResp, ok := err.(*ErrorResponse)
	if !ok {
		t.Fatalf("expected *ErrorResponse, got %T", err)
	}
	if errResp.ErrorIdentifier != InternalError {
		t.Errorf("expected error %q, got %q", InternalError, errResp.ErrorIdentifier)
	}
}
