package supadata

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

type TranscriptContent struct {
	Text     string  `json:"text"`
	Offset   float64 `json:"offset"`
	Duration int     `json:"duration"`
	Lang     string  `json:"lang"`
}

type SyncTranscript struct {
	Content        []TranscriptContent `json:"content"`
	Lang           string              `json:"lang"`
	AvailableLangs []string            `json:"availableLangs"`
}

type AsyncTranscript struct {
	JobId string `json:"jobId"`
}

type TranscriptModeParam string

const (
	Native   TranscriptModeParam = "native"
	Auto     TranscriptModeParam = "auto"
	Generate TranscriptModeParam = "generate"
)

type TranscriptParams struct {
	Url       string
	Lang      string
	Text      bool
	ChunkSize int
	Mode      TranscriptModeParam
}

type TranscriptResultStatus string

const (
	Queued    TranscriptResultStatus = "queued"
	Active    TranscriptResultStatus = "active"
	Completed TranscriptResultStatus = "completed"
	Failed    TranscriptResultStatus = "failed"
)

type TranscriptResult struct {
	Status         TranscriptResultStatus `json:"status"`
	Error          *ErrorResponse         `json:"error,omitempty"`
	Content        []TranscriptContent    `json:"content,omitempty"`
	Lang           string                 `json:"lang,omitempty"`
	AvailableLangs []string               `json:"availableLangs,omitempty"`
}

type MetadataPlatform string

const (
	YouTube   MetadataPlatform = "youtube"
	TikTok    MetadataPlatform = "tiktok"
	Instagram MetadataPlatform = "instagram"
	Twitter   MetadataPlatform = "twitter"
	Facebook  MetadataPlatform = "facebook"
)

type MetadataType string

const (
	Video    MetadataType = "video"
	Image    MetadataType = "image"
	Carousel MetadataType = "carousel"
	Post     MetadataType = "post"
)

type Metadata struct {
	Platform    MetadataPlatform `json:"platform"`
	Type        MetadataType     `json:"type"`
	Id          string           `json:"id"`
	Url         string           `json:"url"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Author      struct {
		DisplayName string `json:"displayName"`
		Username    string `json:"username"`
		AvatarUrl   string `json:"avatarUrl"`
		Verified    bool   `json:"verified"`
	} `json:"author"`
	Stats struct {
		Likes    *int `json:"likes"`
		Comments *int `json:"comments"`
		Shares   *int `json:"shares"`
		Views    *int `json:"views"`
	} `json:"stats"`
	Media struct {
		Type         string  `json:"type"`
		Duration     float64 `json:"duration,omitempty"`
		ThumbnailUrl string  `json:"thumbnailUrl,omitempty"`
		Url          string  `json:"url,omitempty"`
		Items        []struct {
			Type         string  `json:"type"`
			Duration     float64 `json:"duration,omitempty"`
			ThumbnailUrl string  `json:"thumbnailUrl,omitempty"`
			Url          string  `json:"url,omitempty"`
		} `json:"items,omitempty"`
	} `json:"media"`
	Tags           []string               `json:"tags,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	AdditionalData map[string]interface{} `json:"additionalData,omitempty"`
}

type SupadataConfig struct {
	apiKey  string
	client  *http.Client
	timeout time.Duration
}

type Supadata struct {
	config *SupadataConfig
}

func (s *Supadata) setDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "supadata-go/1.0.0")
	req.Header.Set("x-api-key", s.config.apiKey)
}

type SupadataOption func(*SupadataConfig)

func WithAPIKey(apiKey string) SupadataOption {
	return func(config *SupadataConfig) {
		config.apiKey = apiKey
	}
}

func WithTimeout(timeout time.Duration) SupadataOption {
	return func(config *SupadataConfig) {
		config.client.Timeout = timeout
	}
}

func WithClient(client *http.Client) SupadataOption {
	return func(config *SupadataConfig) {
		config.client = client
	}
}

func NewSupadata(opts ...SupadataOption) *Supadata {
	defaultClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: http.DefaultTransport,
	}

	c := &SupadataConfig{
		apiKey: os.Getenv("SUPADATA_API_KEY"),
		client: defaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return &Supadata{
		config: c,
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

// handleResponse is a generic function that handles HTTP responses and unmarshals them into the specified type
func handleResponse[T any](resp *http.Response) (*T, error) {
	body, err := handleRawResponse(resp)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// handleRawResponse handles HTTP responses and returns the raw body bytes for custom processing
func handleRawResponse(resp *http.Response) ([]byte, error) {
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
	return body, nil
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

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := handleRawResponse(resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

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

func (s *Supadata) TranscriptResult(jobId string) (*TranscriptResult, error) {
	req, err := s.prepareRequest("GET", "/transcript/"+jobId, nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return handleResponse[TranscriptResult](resp)
}

func (s *Supadata) Metadata(url string) (*Metadata, error) {
	req, err := s.prepareRequest("GET", "/metadata", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("url", url)
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return handleResponse[Metadata](resp)
}
