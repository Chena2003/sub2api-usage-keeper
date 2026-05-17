package quota

import (
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
