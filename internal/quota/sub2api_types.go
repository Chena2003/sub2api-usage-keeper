package quota

import "time"

type Sub2APIAccountUsage struct {
	TotalRequests     int64   `json:"totalRequests"`
	InputTokens       int64   `json:"inputTokens"`
	OutputTokens      int64   `json:"outputTokens"`
	CacheTokens       int64   `json:"cacheTokens"`
	TotalTokens       int64   `json:"totalTokens"`
	ActualCost        float64 `json:"actualCost"`
	AverageDurationMS float64 `json:"averageDurationMs"`
}

type Sub2APIQuotaWindow struct {
	Utilization *float64   `json:"utilization,omitempty"`
	Consumed    int64      `json:"consumed"`
	Limit       *int64     `json:"limit,omitempty"`
	Remaining   *int64     `json:"remaining,omitempty"`
	Ratio       *float64   `json:"ratio,omitempty"`
	WindowStart *time.Time `json:"windowStart,omitempty"`
	WindowEnd   *time.Time `json:"windowEnd,omitempty"`
	RefreshAt   *time.Time `json:"refreshAt,omitempty"`
	Status      string     `json:"status"`
}

type Sub2APIRankingRow struct {
	Dimension     string  `json:"dimension"`
	Name          string  `json:"name"`
	TotalRequests int64   `json:"totalRequests"`
	InputTokens   int64   `json:"inputTokens"`
	OutputTokens  int64   `json:"outputTokens"`
	CacheTokens   int64   `json:"cacheTokens"`
	TotalTokens   int64   `json:"totalTokens"`
	ActualCost    float64 `json:"actualCost"`
	Share         float64 `json:"share"`
}

type Sub2APIRankingTrendPoint struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
	Tokens int64  `json:"tokens"`
}

type Sub2APIEvent struct {
	ID                   int64     `json:"id"`
	CreatedAt            time.Time `json:"createdAt"`
	User                 string    `json:"user"`
	APIKey               string    `json:"apiKey"`
	Model                string    `json:"model"`
	RequestedModel       string    `json:"requestedModel"`
	UpstreamModel        string    `json:"upstreamModel"`
	AccountID            int64     `json:"accountId"`
	AccountName          string    `json:"accountName"`
	Status               string    `json:"status"`
	InputTokens          int64     `json:"inputTokens"`
	OutputTokens         int64     `json:"outputTokens"`
	CacheTokens          int64     `json:"cacheTokens"`
	TotalTokens          int64     `json:"totalTokens"`
	ActualCost           float64   `json:"actualCost"`
	DurationMS           int64     `json:"durationMs"`
	FirstTokenDurationMS *int64    `json:"firstTokenDurationMs,omitempty"`
}

type Sub2APIEventsResponse struct {
	Events []Sub2APIEvent `json:"events"`
	Total  int64          `json:"total"`
	Page   int            `json:"page"`
	Limit  int            `json:"limit"`
}

type Sub2APIAccountQuota struct {
	ID                  int64               `json:"id"`
	Provider            string              `json:"provider"`
	AccountType         string              `json:"accountType"`
	DisplayName         string              `json:"displayName"`
	PlanType            string              `json:"planType,omitempty"`
	Status              string              `json:"status"`
	StatusDetail        string              `json:"statusDetail"`
	StatusResetAt       *time.Time          `json:"statusResetAt,omitempty"`
	HasError            bool                `json:"hasError"`
	Schedulable         bool                `json:"schedulable"`
	SessionWindowStatus string              `json:"sessionWindowStatus,omitempty"`
	ResetAt             *time.Time          `json:"resetAt,omitempty"`
	ExpiresAt           *time.Time          `json:"expiresAt,omitempty"`
	LastUsedAt          *time.Time          `json:"lastUsedAt,omitempty"`
	CredentialKeys      []string            `json:"credentialKeys"`
	Usage               Sub2APIAccountUsage `json:"usage"`
	FiveHourWindow      Sub2APIQuotaWindow  `json:"fiveHourWindow"`
	WeeklyWindow        Sub2APIQuotaWindow  `json:"weeklyWindow"`
}

type Sub2APIOverview struct {
	AccountCount       int64   `json:"accountCount"`
	ActiveAccountCount int64   `json:"activeAccountCount"`
	TotalRequests      int64   `json:"totalRequests"`
	InputTokens        int64   `json:"inputTokens"`
	OutputTokens       int64   `json:"outputTokens"`
	CacheTokens        int64   `json:"cacheTokens"`
	TotalTokens        int64   `json:"totalTokens"`
	ActualCost         float64 `json:"actualCost"`
	AccountCost        float64 `json:"accountCost"`
	ActiveUsers        int64   `json:"activeUsers"`
}
