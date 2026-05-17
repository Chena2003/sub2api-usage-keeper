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

type Sub2APIAccountQuota struct {
	ID                  int64               `json:"id"`
	Provider            string              `json:"provider"`
	AccountType         string              `json:"accountType"`
	DisplayName         string              `json:"displayName"`
	PlanType            string              `json:"planType,omitempty"`
	Status              string              `json:"status"`
	Schedulable         bool                `json:"schedulable"`
	SessionWindowStatus string              `json:"sessionWindowStatus,omitempty"`
	ResetAt             *time.Time          `json:"resetAt,omitempty"`
	ExpiresAt           *time.Time          `json:"expiresAt,omitempty"`
	LastUsedAt          *time.Time          `json:"lastUsedAt,omitempty"`
	CredentialKeys      []string            `json:"credentialKeys"`
	Usage               Sub2APIAccountUsage `json:"usage"`
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
