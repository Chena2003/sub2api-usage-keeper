package sub2api

import (
	"encoding/json"
	"sort"
	"time"
)

type AccountRow struct {
	ID                 int64           `json:"id" gorm:"column:id"`
	Name               string          `json:"name" gorm:"column:name"`
	Platform           string          `json:"platform" gorm:"column:platform"`
	Type               string          `json:"type" gorm:"column:type"`
	Status             string          `json:"status" gorm:"column:status"`
	Schedulable        bool            `json:"schedulable" gorm:"column:schedulable"`
	Credentials        json.RawMessage `json:"-" gorm:"column:credentials"`
	RateLimitedAt      *time.Time      `json:"rateLimitedAt" gorm:"column:rate_limited_at"`
	RateLimitResetAt   *time.Time      `json:"rateLimitResetAt" gorm:"column:rate_limit_reset_at"`
	SessionWindowStart *time.Time      `json:"sessionWindowStart" gorm:"column:session_window_start"`
	SessionWindowEnd   *time.Time      `json:"sessionWindowEnd" gorm:"column:session_window_end"`
	SessionWindowStatus string         `json:"sessionWindowStatus" gorm:"column:session_window_status"`
	ExpiresAt          *time.Time      `json:"expiresAt" gorm:"column:expires_at"`
	RateMultiplier     *float64        `json:"rateMultiplier" gorm:"column:rate_multiplier"`
	LoadFactor         *int            `json:"loadFactor" gorm:"column:load_factor"`
	LastUsedAt         *time.Time      `json:"lastUsedAt" gorm:"column:last_used_at"`
	UpdatedAt          time.Time       `json:"updatedAt" gorm:"column:updated_at"`
}

func (r AccountRow) CredentialKeys() []string {
	if len(r.Credentials) == 0 {
		return nil
	}

	var credentials map[string]any
	if err := json.Unmarshal(r.Credentials, &credentials); err != nil || len(credentials) == 0 {
		return nil
	}

	keys := make([]string, 0, len(credentials))
	for key := range credentials {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type UsageOverviewRow struct {
	BucketStart         time.Time `json:"bucketStart" gorm:"column:bucket_start"`
	BucketDate          time.Time `json:"bucketDate" gorm:"column:bucket_date"`
	TotalRequests       int64     `json:"totalRequests" gorm:"column:total_requests"`
	InputTokens         int64     `json:"inputTokens" gorm:"column:input_tokens"`
	OutputTokens        int64     `json:"outputTokens" gorm:"column:output_tokens"`
	CacheCreationTokens int64     `json:"cacheCreationTokens" gorm:"column:cache_creation_tokens"`
	CacheReadTokens     int64     `json:"cacheReadTokens" gorm:"column:cache_read_tokens"`
	TotalCost           float64   `json:"totalCost" gorm:"column:total_cost"`
	ActualCost          float64   `json:"actualCost" gorm:"column:actual_cost"`
	AccountCost         float64   `json:"accountCost" gorm:"column:account_cost"`
	TotalDurationMS     int64     `json:"totalDurationMS" gorm:"column:total_duration_ms"`
	ActiveUsers         int64     `json:"activeUsers" gorm:"column:active_users"`
}

func (r UsageOverviewRow) TotalTokens() int64 {
	return r.InputTokens + r.OutputTokens + r.CacheCreationTokens + r.CacheReadTokens
}

type ModelUsageRow struct {
	Model               string  `json:"model" gorm:"column:model"`
	RequestedModel      string  `json:"requestedModel" gorm:"column:requested_model"`
	UpstreamModel       string  `json:"upstreamModel" gorm:"column:upstream_model"`
	TotalRequests       int64   `json:"totalRequests" gorm:"column:total_requests"`
	InputTokens         int64   `json:"inputTokens" gorm:"column:input_tokens"`
	OutputTokens        int64   `json:"outputTokens" gorm:"column:output_tokens"`
	CacheCreationTokens int64   `json:"cacheCreationTokens" gorm:"column:cache_creation_tokens"`
	CacheReadTokens     int64   `json:"cacheReadTokens" gorm:"column:cache_read_tokens"`
	TotalCost           float64 `json:"totalCost" gorm:"column:total_cost"`
	ActualCost          float64 `json:"actualCost" gorm:"column:actual_cost"`
	AverageDurationMS   float64 `json:"averageDurationMS" gorm:"column:average_duration_ms"`
}

func (r ModelUsageRow) TotalTokens() int64 {
	return r.InputTokens + r.OutputTokens + r.CacheCreationTokens + r.CacheReadTokens
}

type AccountUsageRow struct {
	AccountID           int64      `json:"accountId" gorm:"column:account_id"`
	TotalRequests       int64      `json:"totalRequests" gorm:"column:total_requests"`
	InputTokens         int64      `json:"inputTokens" gorm:"column:input_tokens"`
	OutputTokens        int64      `json:"outputTokens" gorm:"column:output_tokens"`
	CacheCreationTokens int64      `json:"cacheCreationTokens" gorm:"column:cache_creation_tokens"`
	CacheReadTokens     int64      `json:"cacheReadTokens" gorm:"column:cache_read_tokens"`
	TotalCost           float64    `json:"totalCost" gorm:"column:total_cost"`
	ActualCost          float64    `json:"actualCost" gorm:"column:actual_cost"`
	AverageDurationMS   float64    `json:"averageDurationMS" gorm:"column:average_duration_ms"`
	LastRequestAt       *time.Time `json:"lastRequestAt" gorm:"column:last_request_at"`
}

func (r AccountUsageRow) TotalTokens() int64 {
	return r.InputTokens + r.OutputTokens + r.CacheCreationTokens + r.CacheReadTokens
}
