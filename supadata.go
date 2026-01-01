package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	BaseUrl = "https://api.supadata.ai/v1"
)

type ErrorIdentifier string

const (
	InvalidRequest        ErrorIdentifier = "invalid-request"
	InternalError         ErrorIdentifier = "internal-error"
	Forbidden             ErrorIdentifier = "forbidden"
	Unauthorized          ErrorIdentifier = "unauthorized"
	UpgradeRequired       ErrorIdentifier = "upgrade-required"
	TranscriptUnavailable ErrorIdentifier = "transcript-unavailable"
	NotFound              ErrorIdentifier = "not-found"
	LimitExceeded         ErrorIdentifier = "limit-exceeded"
)

type ErrorResponse struct {
	ErrorIdentifier  ErrorIdentifier `json:"error"`
	Message          string          `json:"message"`
	Details          string          `json:"details"`
	DocumentationUrl string          `json:"documentationUrl"`
}

func (e *ErrorResponse) Error() string {
	return fmt.Sprintf("%s: %s", e.ErrorIdentifier, e.Message)
}

type Transcript struct {
	Sync  *SyncTranscript
	Async *AsyncTranscript
}

func (r *Transcript) IsAsync() bool {
	return r.Async != nil
}

type SyncTranscript struct {
	Content []struct {
		Text     string `json:"text"`
		Offset   int    `json:"offset"`
		Duration int    `json:"duration"`
		Lang     string `json:"lang"`
	} `json:"content"`
	Lang           string   `json:"lang"`
	AvailableLangs []string `json:"availableLangs"`
}

type AsyncTranscript struct {
	JobId string `json:"jobId"`
}

type TranscriptModeParam string

const (
	Native   TranscriptModeParam = "native"
	Auto     TranscriptModeParam = "auto"
	Generate TranscriptModeParam = "all"
)

type TranscriptParams struct {
	Url       string
	Lang      string
	Text      bool
	ChunkSize int
	Mode      TranscriptModeParam
}

type SupadataConfig struct {
	APIKey string
	Client *http.Client
}

type Supadata struct {
	config *SupadataConfig
}

func (s *Supadata) setDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "supadata-go/1.0.0")
	req.Header.Set("x-api-key", s.config.APIKey)
}

func NewSupadata(config *SupadataConfig) *Supadata {
	defaultClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: http.DefaultTransport,
	}

	apiKey := os.Getenv("SUPADATA_API_KEY")

	if config == nil && apiKey != "" {
		config = &SupadataConfig{
			APIKey: apiKey,
			Client: defaultClient,
		}

	}

	if config != nil && config.Client == nil {
		config.Client = defaultClient
	}
	return &Supadata{
		config: config,
	}
}

func (s *Supadata) prepareRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, BaseUrl+endpoint, body)
	if err != nil {
		return nil, err
	}
	s.setDefaultHeaders(req)
	return req, nil
}

// Universal Endpoints
func (s *Supadata) Transcript(params *TranscriptParams) (*Transcript, error) {
	req, err := s.prepareRequest("GET", "/transcript", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("url", params.Url)
	if params.Lang != "" {
		q.Set("lang", params.Lang)
	}
	if params.Text {
		q.Set("text", "true")
	}
	if params.ChunkSize > 0 {
		q.Set("chunkSize", fmt.Sprintf("%d", params.ChunkSize))
	}
	if params.Mode != "" {
		q.Set("mode", string(params.Mode))
	} else {
		q.Set("mode", string(Auto))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return nil, &errResp
	}

	// Check if response is async (has jobId) or sync (has content)
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	if _, hasJobId := raw["jobId"]; hasJobId {
		var async AsyncTranscript
		if err := json.Unmarshal(body, &async); err != nil {
			return nil, err
		}
		return &Transcript{Async: &async}, nil
	}

	var sync SyncTranscript
	if err := json.Unmarshal(body, &sync); err != nil {
		return nil, err
	}
	return &Transcript{Sync: &sync}, nil
}
