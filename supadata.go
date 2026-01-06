package supadata

import (
	"bytes"
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
	Duration float64 `json:"duration"`
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
	Tags           []string       `json:"tags,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	AdditionalData map[string]any `json:"additionalData,omitempty"`
}

type AccountInfo struct {
	OrganizationId string `json:"organizationId"`
	Plan           string `json:"plan"`
	MaxCredits     int    `json:"maxCredits"`
	UsedCredits    int    `json:"usedCredits"`
}

type ScrapeParams struct {
	Url     string
	NoLinks bool
	Lang    string
}

type ScrapeResult struct {
	Url             string   `json:"url"`
	Content         string   `json:"content"`
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	OgUrl           string   `json:"ogUrl"`
	CountCharacters int      `json:"countCharacters"`
	Urls            []string `json:"urls"`
}

type MapParams struct {
	Url     string
	NoLinks bool
	Lang    string
}

type MapResult struct {
	Urls []string `json:"urls"`
}

type CrawlBody struct {
	Url   string `json:"url"`
	Limit int    `json:"limit,omitempty"`
}

type CrawlJob struct {
	JobId string `json:"jobId"`
}

// CrawlStatus represents the status of a crawl job
type CrawlStatus string

const (
	Scraping       CrawlStatus = "scraping"
	CrawlCompleted CrawlStatus = "completed"
	CrawlFailed    CrawlStatus = "failed"
	Cancelled      CrawlStatus = "cancelled"
)

type CrawlPage struct {
	Url             string `json:"url"`
	Content         string `json:"content"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	OgUrl           string `json:"ogUrl"`
	CountCharacters int    `json:"countCharacters"`
}

type CrawlResult struct {
	Status CrawlStatus `json:"status"`
	Pages  []CrawlPage `json:"pages,omitempty"`
	Next   string      `json:"next,omitempty"`
}

// YouTube Types

// YouTubeSearchUploadDate filter for search results
type YouTubeSearchUploadDate string

const (
	UploadDateAll   YouTubeSearchUploadDate = "all"
	UploadDateHour  YouTubeSearchUploadDate = "hour"
	UploadDateToday YouTubeSearchUploadDate = "today"
	UploadDateWeek  YouTubeSearchUploadDate = "week"
	UploadDateMonth YouTubeSearchUploadDate = "month"
	UploadDateYear  YouTubeSearchUploadDate = "year"
)

// YouTubeSearchType filter for search results
type YouTubeSearchType string

const (
	SearchTypeAll      YouTubeSearchType = "all"
	SearchTypeVideo    YouTubeSearchType = "video"
	SearchTypeChannel  YouTubeSearchType = "channel"
	SearchTypePlaylist YouTubeSearchType = "playlist"
	SearchTypeMovie    YouTubeSearchType = "movie"
)

// YouTubeSearchDuration filter for search results
type YouTubeSearchDuration string

const (
	DurationAll    YouTubeSearchDuration = "all"
	DurationShort  YouTubeSearchDuration = "short"
	DurationMedium YouTubeSearchDuration = "medium"
	DurationLong   YouTubeSearchDuration = "long"
)

// YouTubeSearchSortBy sort order for search results
type YouTubeSearchSortBy string

const (
	SortByRelevance YouTubeSearchSortBy = "relevance"
	SortByRating    YouTubeSearchSortBy = "rating"
	SortByDate      YouTubeSearchSortBy = "date"
	SortByViews     YouTubeSearchSortBy = "views"
)

// YouTubeSearchFeature special features filter
type YouTubeSearchFeature string

const (
	FeatureHD             YouTubeSearchFeature = "hd"
	FeatureSubtitles      YouTubeSearchFeature = "subtitles"
	FeatureCreativeCommon YouTubeSearchFeature = "creative-commons"
	Feature3D             YouTubeSearchFeature = "3d"
	FeatureLive           YouTubeSearchFeature = "live"
	Feature4K             YouTubeSearchFeature = "4k"
	Feature360            YouTubeSearchFeature = "360"
	FeatureLocation       YouTubeSearchFeature = "location"
	FeatureHDR            YouTubeSearchFeature = "hdr"
	FeatureVR180          YouTubeSearchFeature = "vr180"
)

type YouTubeSearchParams struct {
	Query         string
	UploadDate    YouTubeSearchUploadDate
	Type          YouTubeSearchType
	Duration      YouTubeSearchDuration
	SortBy        YouTubeSearchSortBy
	Features      []YouTubeSearchFeature
	Limit         int
	NextPageToken string
}

type YouTubeSearchResultItem struct {
	Type            string `json:"type"`
	Id              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Thumbnail       string `json:"thumbnail"`
	Duration        int    `json:"duration,omitempty"`
	ViewCount       *int   `json:"viewCount,omitempty"`
	UploadDate      string `json:"uploadDate,omitempty"`
	ChannelId       string `json:"channelId,omitempty"`
	ChannelName     string `json:"channelName,omitempty"`
	SubscriberCount *int   `json:"subscriberCount,omitempty"`
	VideoCount      *int   `json:"videoCount,omitempty"`
}

type YouTubeSearchResult struct {
	Query         string                    `json:"query"`
	Results       []YouTubeSearchResultItem `json:"results"`
	TotalResults  int                       `json:"totalResults"`
	NextPageToken string                    `json:"nextPageToken,omitempty"`
}

type YouTubeVideoChannel struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type YouTubeVideo struct {
	Id                  string              `json:"id"`
	Title               string              `json:"title"`
	Description         string              `json:"description"`
	Duration            int                 `json:"duration"`
	Channel             YouTubeVideoChannel `json:"channel"`
	Tags                []string            `json:"tags"`
	Thumbnail           string              `json:"thumbnail"`
	UploadDate          *string             `json:"uploadDate"`
	ViewCount           *int                `json:"viewCount"`
	LikeCount           *int                `json:"likeCount"`
	TranscriptLanguages []string            `json:"transcriptLanguages"`
}

type YouTubeVideoBatchParams struct {
	VideoIds   []string `json:"videoIds,omitempty"`
	PlaylistId string   `json:"playlistId,omitempty"`
	ChannelId  string   `json:"channelId,omitempty"`
	Limit      int      `json:"limit,omitempty"`
}

type YouTubeBatchJob struct {
	JobId string `json:"jobId"`
}

type YouTubeTranscriptParams struct {
	Url       string
	VideoId   string
	Text      bool
	ChunkSize int
	Lang      string
}

type YouTubeTranscriptResult struct {
	Content        []TranscriptContent `json:"content"`
	Lang           string              `json:"lang"`
	AvailableLangs []string            `json:"availableLangs"`
}

type YouTubeTranscriptBatchParams struct {
	VideoIds   []string `json:"videoIds,omitempty"`
	PlaylistId string   `json:"playlistId,omitempty"`
	ChannelId  string   `json:"channelId,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Lang       string   `json:"lang,omitempty"`
	Text       bool     `json:"text,omitempty"`
}

type YouTubeTranscriptTranslateParams struct {
	Url       string
	VideoId   string
	Text      bool
	ChunkSize int
	Lang      string
}

type YouTubeTranscriptTranslateResult struct {
	Content []TranscriptContent `json:"content"`
	Lang    string              `json:"lang"`
}

type YouTubeChannel struct {
	Id              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description,omitempty"`
	SubscriberCount *int   `json:"subscriberCount,omitempty"`
	VideoCount      *int   `json:"videoCount,omitempty"`
	ViewCount       *int   `json:"viewCount,omitempty"`
	Thumbnail       string `json:"thumbnail,omitempty"`
	Banner          string `json:"banner,omitempty"`
}

type YouTubePlaylist struct {
	Id          string              `json:"id"`
	Title       string              `json:"title"`
	Description string              `json:"description,omitempty"`
	VideoCount  int                 `json:"videoCount"`
	ViewCount   *int                `json:"viewCount,omitempty"`
	LastUpdated *string             `json:"lastUpdated,omitempty"`
	Channel     YouTubeVideoChannel `json:"channel"`
}

// YouTubeChannelVideoType filter for channel videos
type YouTubeChannelVideoType string

const (
	ChannelVideoTypeAll   YouTubeChannelVideoType = "all"
	ChannelVideoTypeVideo YouTubeChannelVideoType = "video"
	ChannelVideoTypeShort YouTubeChannelVideoType = "short"
	ChannelVideoTypeLive  YouTubeChannelVideoType = "live"
)

type YouTubeChannelVideosParams struct {
	Id    string
	Limit int
	Type  YouTubeChannelVideoType
}

type YouTubeChannelVideosResult struct {
	VideoIds []string `json:"videoIds"`
	ShortIds []string `json:"shortIds"`
	LiveIds  []string `json:"liveIds"`
}

type YouTubePlaylistVideosParams struct {
	Id    string
	Limit int
}

type YouTubePlaylistVideosResult struct {
	VideoIds []string `json:"videoIds"`
	ShortIds []string `json:"shortIds"`
	LiveIds  []string `json:"liveIds"`
}

// YouTubeBatchStatus represents the status of a batch job
type YouTubeBatchStatus string

const (
	BatchQueued    YouTubeBatchStatus = "queued"
	BatchActive    YouTubeBatchStatus = "active"
	BatchCompleted YouTubeBatchStatus = "completed"
	BatchFailed    YouTubeBatchStatus = "failed"
)

type YouTubeBatchResultItem struct {
	VideoId    string                   `json:"videoId"`
	Transcript *YouTubeTranscriptResult `json:"transcript,omitempty"`
	Video      *YouTubeVideo            `json:"video,omitempty"`
	ErrorCode  string                   `json:"errorCode,omitempty"`
}

type YouTubeBatchStats struct {
	Total     int `json:"total"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
}

type YouTubeBatchResult struct {
	Status      YouTubeBatchStatus       `json:"status"`
	Results     []YouTubeBatchResultItem `json:"results,omitempty"`
	Stats       YouTubeBatchStats        `json:"stats"`
	CompletedAt *string                  `json:"completedAt,omitempty"`
}

type Config struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

type Supadata struct {
	config *Config
}

func (s *Supadata) setDefaultHeaders(req *http.Request) {
	req.Header.Set("User-Agent", "supadata-go/1.0.0")
	req.Header.Set("x-api-key", s.config.apiKey)
}

type ConfigOption func(*Config)

func WithAPIKey(apiKey string) ConfigOption {
	return func(config *Config) {
		config.apiKey = apiKey
	}
}

func WithTimeout(timeout time.Duration) ConfigOption {
	return func(config *Config) {
		config.client.Timeout = timeout
	}
}

func WithClient(client *http.Client) ConfigOption {
	return func(config *Config) {
		config.client = client
	}
}

func WithBaseURL(baseURL string) ConfigOption {
	return func(config *Config) {
		config.baseURL = baseURL
	}
}

func NewSupadata(opts ...ConfigOption) *Supadata {
	defaultClient := &http.Client{
		Timeout:   60 * time.Second,
		Transport: http.DefaultTransport,
	}

	c := &Config{
		apiKey:  os.Getenv("SUPADATA_API_KEY"),
		baseURL: BaseUrl,
		client:  defaultClient,
	}

	for _, opt := range opts {
		opt(c)
	}

	return &Supadata{
		config: c,
	}

}

func (s *Supadata) prepareRequest(method, endpoint string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, s.config.baseURL+endpoint, body)
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

// Transcript initiates a transcript request (sync or async)
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

// TranscriptResult retrieves the result of an async transcript job
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

// Metadata retrieves metadata for a given URL
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

// Account Endpoints

// Me retrieves account information
func (s *Supadata) Me() (*AccountInfo, error) {
	req, err := s.prepareRequest("GET", "/me", nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[AccountInfo](resp)
}

// Web Endpoints

// Scrape extracts content from a webpage as markdown
func (s *Supadata) Scrape(params *ScrapeParams) (*ScrapeResult, error) {
	req, err := s.prepareRequest("GET", "/web/scrape", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("url", params.Url)
	if params.NoLinks {
		q.Set("noLinks", "true")
	}
	if params.Lang != "" {
		q.Set("lang", params.Lang)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[ScrapeResult](resp)
}

// Map discovers all URLs on a website
func (s *Supadata) Map(params *MapParams) (*MapResult, error) {
	req, err := s.prepareRequest("GET", "/web/map", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("url", params.Url)
	if params.NoLinks {
		q.Set("noLinks", "true")
	}
	if params.Lang != "" {
		q.Set("lang", params.Lang)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[MapResult](resp)
}

// Crawl initiates an async crawl job for a website
func (s *Supadata) Crawl(params *CrawlBody) (*CrawlJob, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := s.prepareRequest("POST", "/web/crawl", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[CrawlJob](resp)
}

// CrawlResult retrieves the status and results of a crawl job
func (s *Supadata) CrawlResult(jobId string, skip int) (*CrawlResult, error) {
	req, err := s.prepareRequest("GET", "/web/crawl/"+jobId, nil)
	if err != nil {
		return nil, err
	}

	if skip > 0 {
		q := req.URL.Query()
		q.Set("skip", fmt.Sprintf("%d", skip))
		req.URL.RawQuery = q.Encode()
	}

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[CrawlResult](resp)
}

// YouTube Endpoints

// YouTubeSearch searches YouTube for videos, channels, or playlists
func (s *Supadata) YouTubeSearch(params *YouTubeSearchParams) (*YouTubeSearchResult, error) {
	req, err := s.prepareRequest("GET", "/youtube/search", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("query", params.Query)
	if params.UploadDate != "" {
		q.Set("uploadDate", string(params.UploadDate))
	}
	if params.Type != "" {
		q.Set("type", string(params.Type))
	}
	if params.Duration != "" {
		q.Set("duration", string(params.Duration))
	}
	if params.SortBy != "" {
		q.Set("sortBy", string(params.SortBy))
	}
	if len(params.Features) > 0 {
		for _, f := range params.Features {
			q.Add("features", string(f))
		}
	}
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.NextPageToken != "" {
		q.Set("nextPageToken", params.NextPageToken)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeSearchResult](resp)
}

// YouTubeVideo retrieves metadata for a YouTube video
func (s *Supadata) YouTubeVideo(id string) (*YouTubeVideo, error) {
	req, err := s.prepareRequest("GET", "/youtube/video", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("id", id)
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeVideo](resp)
}

// YouTubeVideoBatch initiates a batch job to retrieve multiple video metadata
func (s *Supadata) YouTubeVideoBatch(params *YouTubeVideoBatchParams) (*YouTubeBatchJob, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := s.prepareRequest("POST", "/youtube/video/batch", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeBatchJob](resp)
}

// YouTubeTranscript retrieves the transcript for a YouTube video
func (s *Supadata) YouTubeTranscript(params *YouTubeTranscriptParams) (*YouTubeTranscriptResult, error) {
	req, err := s.prepareRequest("GET", "/youtube/transcript", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if params.Url != "" {
		q.Set("url", params.Url)
	}
	if params.VideoId != "" {
		q.Set("videoId", params.VideoId)
	}
	if params.Text {
		q.Set("text", "true")
	}
	if params.ChunkSize > 0 {
		q.Set("chunkSize", fmt.Sprintf("%d", params.ChunkSize))
	}
	if params.Lang != "" {
		q.Set("lang", params.Lang)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeTranscriptResult](resp)
}

// YouTubeTranscriptBatch initiates a batch job to retrieve transcripts for multiple videos
func (s *Supadata) YouTubeTranscriptBatch(params *YouTubeTranscriptBatchParams) (*YouTubeBatchJob, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	req, err := s.prepareRequest("POST", "/youtube/transcript/batch", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeBatchJob](resp)
}

// YouTubeTranscriptTranslate retrieves a translated transcript for a YouTube video
func (s *Supadata) YouTubeTranscriptTranslate(params *YouTubeTranscriptTranslateParams) (*YouTubeTranscriptTranslateResult, error) {
	req, err := s.prepareRequest("GET", "/youtube/transcript/translate", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	if params.Url != "" {
		q.Set("url", params.Url)
	}
	if params.VideoId != "" {
		q.Set("videoId", params.VideoId)
	}
	if params.Text {
		q.Set("text", "true")
	}
	if params.ChunkSize > 0 {
		q.Set("chunkSize", fmt.Sprintf("%d", params.ChunkSize))
	}
	q.Set("lang", params.Lang)
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeTranscriptTranslateResult](resp)
}

// YouTubeChannel retrieves metadata for a YouTube channel
func (s *Supadata) YouTubeChannel(id string) (*YouTubeChannel, error) {
	req, err := s.prepareRequest("GET", "/youtube/channel", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("id", id)
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeChannel](resp)
}

// YouTubePlaylist retrieves metadata for a YouTube playlist
func (s *Supadata) YouTubePlaylist(id string) (*YouTubePlaylist, error) {
	req, err := s.prepareRequest("GET", "/youtube/playlist", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("id", id)
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubePlaylist](resp)
}

// YouTubeChannelVideos retrieves video IDs from a YouTube channel
func (s *Supadata) YouTubeChannelVideos(params *YouTubeChannelVideosParams) (*YouTubeChannelVideosResult, error) {
	req, err := s.prepareRequest("GET", "/youtube/channel/videos", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("id", params.Id)
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Type != "" {
		q.Set("type", string(params.Type))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeChannelVideosResult](resp)
}

// YouTubePlaylistVideos retrieves video IDs from a YouTube playlist
func (s *Supadata) YouTubePlaylistVideos(params *YouTubePlaylistVideosParams) (*YouTubePlaylistVideosResult, error) {
	req, err := s.prepareRequest("GET", "/youtube/playlist/videos", nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Set("id", params.Id)
	if params.Limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	req.URL.RawQuery = q.Encode()

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubePlaylistVideosResult](resp)
}

// YouTubeBatchResult retrieves the status and results of a batch job
func (s *Supadata) YouTubeBatchResult(jobId string) (*YouTubeBatchResult, error) {
	req, err := s.prepareRequest("GET", "/youtube/batch/"+jobId, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.config.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return handleResponse[YouTubeBatchResult](resp)
}
