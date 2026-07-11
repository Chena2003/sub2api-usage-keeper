package quota

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

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

func NormalizeSub2APIAccount(row sub2api.AccountRow, usage *sub2api.AccountUsageRow, fiveHourUsage *sub2api.AccountUsageRow) Sub2APIAccountQuota {
	provider := strings.TrimSpace(row.Platform)
	if provider == "" {
		provider = "account"
	}

	statusDetail, statusResetAt, hasError := deriveSub2APIStatus(row)

	return Sub2APIAccountQuota{
		ID:                  row.ID,
		Provider:            provider,
		AccountType:         strings.TrimSpace(row.Type),
		DisplayName:         sub2APIDisplayName(row.ID, row.Name, provider, row.Type),
		PlanType:            sub2APIPlanType(row.Credentials),
		Status:              normalizedSub2APIStatus(row.Status, row.Schedulable),
		StatusDetail:        statusDetail,
		StatusResetAt:       statusResetAt,
		HasError:            hasError,
		Schedulable:         row.Schedulable,
		SessionWindowStatus: strings.TrimSpace(row.SessionWindowStatus),
		ResetAt:             row.RateLimitResetAt,
		ExpiresAt:           row.ExpiresAt,
		LastUsedAt:          row.LastUsedAt,
		CredentialKeys:      publicCredentialKeys(row.Credentials),
		Usage:               normalizeSub2APIAccountUsage(usage),
		FiveHourWindow:      buildFiveHourWindow(row, fiveHourUsage),
		WeeklyWindow:        buildWeeklyWindow(row),
	}
}

// deriveSub2APIStatus mirrors the gateway's AccountStatusIndicator priority:
// error -> rate-limited(429) -> overloaded(529) -> temp-unschedulable -> paused
// -> inactive -> active. Transient states carry their recovery time. The raw
// error_message / temp_unschedulable_reason text is intentionally NOT exposed.
func deriveSub2APIStatus(row sub2api.AccountRow) (detail string, resetAt *time.Time, hasError bool) {
	now := time.Now()
	status := strings.TrimSpace(strings.ToLower(row.Status))
	hasError = status == "error" || strings.TrimSpace(row.ErrorMessage) != ""

	switch {
	case status == "error":
		return "error", nil, hasError
	case row.RateLimitResetAt != nil && row.RateLimitResetAt.After(now):
		return "rate_limited", row.RateLimitResetAt, hasError
	case row.OverloadUntil != nil && row.OverloadUntil.After(now):
		return "overloaded", row.OverloadUntil, hasError
	case row.TempUnschedulableUntil != nil && row.TempUnschedulableUntil.After(now):
		return "temp_unschedulable", row.TempUnschedulableUntil, hasError
	case !row.Schedulable:
		return "paused", nil, hasError
	case status != "active" && status != "":
		return "inactive", nil, hasError
	default:
		return "active", nil, hasError
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
		return []string{}
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

func sub2APIDisplayName(id int64, name string, provider string, accountType string) string {
	if n := strings.TrimSpace(name); n != "" {
		return n
	}
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

// buildFiveHourWindow reports the upstream-provided 5-hour utilization (0-100%).
// Source of truth is account.extra, persisted by the gateway from the provider
// /usage API — never an inferred token limit. Consumed tokens (from usage_logs)
// are kept only as a window stat. No extra data -> status "unknown", no fake bar.
func buildFiveHourWindow(row sub2api.AccountRow, fiveHourUsage *sub2api.AccountUsageRow) Sub2APIQuotaWindow {
	extra := row.ExtraMap()
	platform := strings.ToLower(strings.TrimSpace(row.Platform))

	window := Sub2APIQuotaWindow{
		Consumed: accountUsageTokens(fiveHourUsage),
		Status:   normalizeWindowStatus(row.SessionWindowStatus),
	}
	if row.SessionWindowStart != nil {
		window.WindowStart = row.SessionWindowStart
	}

	if isOpenAIPlatform(platform) {
		// Codex stores percent directly (0-100).
		if pct, ok := extraFloat(extra, "codex_5h_used_percent"); ok {
			window.Utilization = &pct
		}
		if reset, ok := extraResetTime(extra, "codex_5h_reset_at"); ok {
			window.WindowEnd = reset
			window.RefreshAt = reset
		}
	} else {
		// Anthropic stores a 0-1 ratio.
		if ratio, ok := extraFloat(extra, "session_window_utilization"); ok {
			pct := ratio * 100
			window.Utilization = &pct
		}
		if row.SessionWindowEnd != nil {
			window.WindowEnd = row.SessionWindowEnd
			window.RefreshAt = row.SessionWindowEnd
		}
	}

	// Fall back to session window status when no explicit utilization is stored.
	if window.Utilization == nil {
		if fallback, ok := sessionStatusUtilization(row.SessionWindowStatus); ok {
			window.Utilization = &fallback
		}
	}

	if row.RateLimitResetAt != nil {
		window.RefreshAt = row.RateLimitResetAt
	}
	if row.RateLimitedAt != nil {
		window.Status = "rate_limited"
	}
	if window.Utilization == nil && window.Consumed == 0 && window.RefreshAt == nil {
		window.Status = "unknown"
	}
	return window
}

// buildWeeklyWindow reports the upstream-provided 7-day utilization (0-100%) from
// account.extra. No extra data -> status "unknown", no fabricated limit.
func buildWeeklyWindow(row sub2api.AccountRow) Sub2APIQuotaWindow {
	extra := row.ExtraMap()
	platform := strings.ToLower(strings.TrimSpace(row.Platform))

	window := Sub2APIQuotaWindow{Status: "unknown"}

	if isOpenAIPlatform(platform) {
		if pct, ok := extraFloat(extra, "codex_7d_used_percent"); ok {
			window.Utilization = &pct
			window.Status = normalizeWindowStatus(row.SessionWindowStatus)
		}
		if reset, ok := extraResetTime(extra, "codex_7d_reset_at"); ok {
			window.WindowEnd = reset
			window.RefreshAt = reset
		}
	} else {
		if ratio, ok := extraFloat(extra, "passive_usage_7d_utilization"); ok {
			pct := ratio * 100
			window.Utilization = &pct
			window.Status = normalizeWindowStatus(row.SessionWindowStatus)
		}
		if reset, ok := extraResetTime(extra, "passive_usage_7d_reset"); ok {
			window.WindowEnd = reset
			window.RefreshAt = reset
		}
	}

	return window
}

func isOpenAIPlatform(platform string) bool {
	return platform == "openai" || platform == "codex"
}

// extraFloat reads a numeric value from the decoded extra map.
// Handles json.Number, float64, and numeric strings.
func extraFloat(extra map[string]any, key string) (float64, bool) {
	if extra == nil {
		return 0, false
	}
	raw, ok := extra[key]
	if !ok {
		return 0, false
	}
	switch v := raw.(type) {
	case float64:
		return v, true
	case json.Number:
		if f, err := v.Float64(); err == nil {
			return f, true
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(v), 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// extraResetTime reads a reset timestamp stored either as a unix epoch
// (seconds) or an RFC3339 string.
func extraResetTime(extra map[string]any, key string) (*time.Time, bool) {
	if extra == nil {
		return nil, false
	}
	raw, ok := extra[key]
	if !ok {
		return nil, false
	}
	switch v := raw.(type) {
	case float64:
		if v > 0 {
			t := time.Unix(int64(v), 0).UTC()
			return &t, true
		}
	case json.Number:
		if f, err := v.Float64(); err == nil && f > 0 {
			t := time.Unix(int64(f), 0).UTC()
			return &t, true
		}
	case string:
		s := strings.TrimSpace(v)
		if s == "" {
			return nil, false
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return &t, true
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil && f > 0 {
			t := time.Unix(int64(f), 0).UTC()
			return &t, true
		}
	}
	return nil, false
}

// sessionStatusUtilization mirrors the gateway fallback used when the provider
// has not reported a utilization value yet.
func sessionStatusUtilization(status string) (float64, bool) {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case "rejected":
		return 100, true
	case "allowed_warning":
		return 80, true
	default:
		return 0, false
	}
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
	if local, domain, ok := strings.Cut(value, "@"); ok {
		local = strings.TrimSpace(local)
		domain = strings.TrimSpace(domain)
		if local != "" && domain != "" {
			return fmt.Sprintf("%s@%s", maskEmailLocal(local), domain)
		}
	}
	digest := sha256.Sum256([]byte(strings.ToLower(value)))
	return fmt.Sprintf("user-%x", digest[:4])
}

func maskEmailLocal(local string) string {
	if len(local) <= 1 {
		return "*"
	}
	if len(local) <= 5 {
		return local[:1] + strings.Repeat("*", len(local)-1)
	}
	return local[:3] + strings.Repeat("*", len(local)-5) + local[len(local)-2:]
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

func NormalizeSub2APIRankingTrend(dimension string, rows []sub2api.RankingTrendRow) []Sub2APIRankingTrendPoint {
	points := make([]Sub2APIRankingTrendPoint, 0, len(rows))
	for _, row := range rows {
		name := row.Name
		if dimension == "user" {
			name = maskedSub2APIUser(row.Name)
		}
		if dimension == "api_key" {
			name = maskedSub2APIKey(row.Name)
		}
		points = append(points, Sub2APIRankingTrendPoint{
			Bucket: row.Bucket,
			Name:   name,
			Tokens: row.Tokens,
		})
	}
	return points
}

func NormalizeSub2APIEvents(rows []sub2api.UsageEventRow) []Sub2APIEvent {
	events := make([]Sub2APIEvent, 0, len(rows))
	for _, row := range rows {
		cacheTokens := row.CacheCreationTokens + row.CacheReadTokens
		events = append(events, Sub2APIEvent{
			ID:                   row.ID,
			CreatedAt:            row.CreatedAt,
			User:                 maskedSub2APIUser(row.User),
			APIKey:               maskedSub2APIKey(row.APIKey),
			Model:                row.Model,
			RequestedModel:       row.RequestedModel,
			UpstreamModel:        row.UpstreamModel,
			AccountID:            row.AccountID,
			AccountName:          row.AccountName,
			Status:               row.Status,
			InputTokens:          row.InputTokens,
			OutputTokens:         row.OutputTokens,
			CacheTokens:          cacheTokens,
			TotalTokens:          row.TotalTokens(),
			ActualCost:           row.ActualCost,
			DurationMS:           row.DurationMS,
			FirstTokenDurationMS: row.FirstTokenMS,
		})
	}
	return events
}
