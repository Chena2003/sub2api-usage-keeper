package quota

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"sub2api-usage-keeper/internal/sub2api"
)

var secretCredentialKeys = map[string]struct{}{
	"access_token":  {},
	"refresh_token": {},
	"id_token":      {},
	"api_key":       {},
	"password":      {},
	"session_key":   {},
}

var publicCredentialKeyAllowlist = map[string]struct{}{
	"model_mapping": {},
	"plan_type":     {},
}

func NormalizeSub2APIAccount(row sub2api.AccountRow, usage *sub2api.AccountUsageRow) Sub2APIAccountQuota {
	provider := strings.TrimSpace(row.Platform)
	if provider == "" {
		provider = "account"
	}

	return Sub2APIAccountQuota{
		ID:                  row.ID,
		Provider:            provider,
		AccountType:         strings.TrimSpace(row.Type),
		DisplayName:         sub2APIDisplayName(row.ID, provider, row.Type),
		PlanType:            sub2APIPlanType(row.Credentials),
		Status:              normalizedSub2APIStatus(row.Status, row.Schedulable),
		Schedulable:         row.Schedulable,
		SessionWindowStatus: strings.TrimSpace(row.SessionWindowStatus),
		ResetAt:             row.RateLimitResetAt,
		ExpiresAt:           row.ExpiresAt,
		LastUsedAt:          row.LastUsedAt,
		CredentialKeys:      publicCredentialKeys(row.Credentials),
		Usage:               normalizeSub2APIAccountUsage(usage),
		FiveHourWindow:      buildFiveHourWindow(row, usage),
		WeeklyWindow:        buildUnknownQuotaWindow(),
	}
}

func sub2APIPlanType(credentials json.RawMessage) string {
	fields := credentialFields(credentials)
	if len(fields) == 0 {
		return ""
	}
	planType, ok := fields["plan_type"].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(planType)
}

func publicCredentialKeys(credentials json.RawMessage) []string {
	fields := credentialFields(credentials)
	if len(fields) == 0 {
		return nil
	}

	keys := make([]string, 0, len(fields))
	for key, value := range fields {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" || isEmptyCredentialValue(value) || isSecretCredentialKey(key) || !isAllowedPublicCredentialKey(key) {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func credentialFields(credentials json.RawMessage) map[string]any {
	if len(credentials) == 0 {
		return nil
	}
	var fields map[string]any
	if err := json.Unmarshal(credentials, &fields); err != nil || len(fields) == 0 {
		return nil
	}
	return fields
}

func isEmptyCredentialValue(value any) bool {
	if value == nil {
		return true
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text) == ""
	}
	return false
}

func isSecretCredentialKey(key string) bool {
	if _, ok := secretCredentialKeys[key]; ok {
		return true
	}
	return key == "email" || strings.Contains(key, "token") || strings.Contains(key, "secret")
}

func isAllowedPublicCredentialKey(key string) bool {
	_, ok := publicCredentialKeyAllowlist[key]
	return ok
}

func sub2APIDisplayName(id int64, provider string, accountType string) string {
	provider = strings.TrimSpace(provider)
	accountType = strings.TrimSpace(accountType)
	if accountType == "" {
		return fmt.Sprintf("%s #%d", provider, id)
	}
	return fmt.Sprintf("%s %s #%d", provider, accountType, id)
}

func normalizedSub2APIStatus(status string, schedulable bool) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "unknown"
	}
	if !schedulable && status == "active" {
		return "paused"
	}
	return status
}

func normalizeSub2APIAccountUsage(usage *sub2api.AccountUsageRow) Sub2APIAccountUsage {
	if usage == nil {
		return Sub2APIAccountUsage{}
	}
	cacheTokens := usage.CacheCreationTokens + usage.CacheReadTokens
	return Sub2APIAccountUsage{
		TotalRequests:     usage.TotalRequests,
		InputTokens:       usage.InputTokens,
		OutputTokens:      usage.OutputTokens,
		CacheTokens:       cacheTokens,
		TotalTokens:       usage.InputTokens + usage.OutputTokens + cacheTokens,
		ActualCost:        usage.ActualCost,
		AverageDurationMS: usage.AverageDurationMS,
	}
}

func buildFiveHourWindow(row sub2api.AccountRow, usage *sub2api.AccountUsageRow) Sub2APIQuotaWindow {
	window := Sub2APIQuotaWindow{
		Consumed: accountUsageTokens(usage),
		Status:   normalizeWindowStatus(row.SessionWindowStatus),
	}
	if row.SessionWindowStart != nil {
		window.WindowStart = row.SessionWindowStart
	}
	if row.SessionWindowEnd != nil {
		window.WindowEnd = row.SessionWindowEnd
		window.RefreshAt = row.SessionWindowEnd
	}
	if row.RateLimitResetAt != nil {
		window.RefreshAt = row.RateLimitResetAt
	}
	if row.RateLimitedAt != nil {
		window.Status = "rate_limited"
	}
	return window
}

func buildUnknownQuotaWindow() Sub2APIQuotaWindow {
	return Sub2APIQuotaWindow{Status: "unknown"}
}

func normalizeWindowStatus(value string) string {
	normalized := strings.TrimSpace(strings.ToLower(value))
	switch normalized {
	case "", "unknown":
		return "unknown"
	case "active", "normal", "ok":
		return "active"
	case "limited", "rate_limited":
		return "rate_limited"
	case "exhausted":
		return "exhausted"
	default:
		return normalized
	}
}

func accountUsageTokens(usage *sub2api.AccountUsageRow) int64 {
	if usage == nil {
		return 0
	}
	return usage.TotalTokens()
}

func maskedSub2APIKey(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown key"
	}
	digest := sha256.Sum256([]byte(strings.ToLower(value)))
	return fmt.Sprintf("key-%x", digest[:4])
}

func maskedSub2APIUser(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "unknown user"
	}
	digest := sha256.Sum256([]byte(strings.ToLower(value)))
	return fmt.Sprintf("user-%x", digest[:4])
}

func NormalizeSub2APIRankings(dimension string, rows []sub2api.RankingRow) []Sub2APIRankingRow {
	var totalTokens int64
	for _, row := range rows {
		totalTokens += row.TotalTokens()
	}

	rankings := make([]Sub2APIRankingRow, 0, len(rows))
	for _, row := range rows {
		cacheTokens := row.CacheCreationTokens + row.CacheReadTokens
		rowTokens := row.TotalTokens()
		share := 0.0
		if totalTokens > 0 {
			share = float64(rowTokens) / float64(totalTokens)
		}
		name := row.Name
		if dimension == "user" {
			name = maskedSub2APIUser(row.Name)
		}
		if dimension == "api_key" {
			name = maskedSub2APIKey(row.Name)
		}
		rankings = append(rankings, Sub2APIRankingRow{
			Dimension:     dimension,
			Name:          name,
			TotalRequests: row.TotalRequests,
			InputTokens:   row.InputTokens,
			OutputTokens:  row.OutputTokens,
			CacheTokens:   cacheTokens,
			TotalTokens:   rowTokens,
			ActualCost:    row.ActualCost,
			Share:         share,
		})
	}
	return rankings
}

func NormalizeSub2APIEvents(rows []sub2api.UsageEventRow) []Sub2APIEvent {
	events := make([]Sub2APIEvent, 0, len(rows))
	for _, row := range rows {
		cacheTokens := row.CacheCreationTokens + row.CacheReadTokens
		events = append(events, Sub2APIEvent{
			ID:             row.ID,
			CreatedAt:      row.CreatedAt,
			User:           maskedSub2APIUser(row.User),
			APIKey:         maskedSub2APIKey(row.APIKey),
			Model:          row.Model,
			RequestedModel: row.RequestedModel,
			UpstreamModel:  row.UpstreamModel,
			AccountID:      row.AccountID,
			AccountName:    row.AccountName,
			Status:         row.Status,
			InputTokens:    row.InputTokens,
			OutputTokens:   row.OutputTokens,
			CacheTokens:    cacheTokens,
			TotalTokens:    row.TotalTokens(),
			ActualCost:     row.ActualCost,
			DurationMS:     row.DurationMS,
		})
	}
	return events
}
