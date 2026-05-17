package quota

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"sub2api-usage-keeper/internal/sub2api"
)

func TestNormalizeSub2APIAccountRedactsSecrets(t *testing.T) {
	resetAt := time.Date(2026, 5, 17, 12, 34, 56, 0, time.UTC)
	credentials := json.RawMessage(`{
		"email":"user@example.com",
		"access_token":"access-secret",
		"refresh_token":"refresh-secret",
		"id_token":"id-secret",
		"plan_type":"pro"
	}`)
	row := sub2api.AccountRow{
		ID:               4,
		Platform:         "openai",
		Type:             "oauth",
		Status:           "active",
		Schedulable:      true,
		Credentials:      credentials,
		RateLimitResetAt: &resetAt,
	}

	normalized := NormalizeSub2APIAccount(row, nil)
	serialized, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("marshal normalized account: %v", err)
	}
	output := string(serialized)

	for _, forbidden := range []string{"access-secret", "refresh-secret", "id-secret", "user@example.com", "access_token", "refresh_token", "id_token", "email"} {
		if strings.Contains(output, forbidden) {
			t.Fatalf("serialized account leaked %q: %s", forbidden, output)
		}
	}

	if normalized.Provider != "openai" {
		t.Fatalf("Provider = %q, want openai", normalized.Provider)
	}
	if normalized.PlanType != "pro" {
		t.Fatalf("PlanType = %q, want pro", normalized.PlanType)
	}
	if normalized.Status != "active" {
		t.Fatalf("Status = %q, want active", normalized.Status)
	}
	if normalized.DisplayName != "openai oauth #4" {
		t.Fatalf("DisplayName = %q, want openai oauth #4", normalized.DisplayName)
	}
	if normalized.ResetAt == nil || !normalized.ResetAt.Equal(resetAt) {
		t.Fatalf("ResetAt = %v, want %v", normalized.ResetAt, resetAt)
	}
}

func TestNormalizeSub2APIAccountPublicCredentialKeysOnlyExposeAllowedKeys(t *testing.T) {
	credentials := json.RawMessage(`{
		"client_id":"client-123",
		"organization_id":"org-456",
		"chatgpt_account_id":"account-789",
		"plan_type":"pro"
	}`)
	row := sub2api.AccountRow{
		ID:          4,
		Platform:    "openai",
		Credentials: credentials,
	}

	normalized := NormalizeSub2APIAccount(row, nil)

	if len(normalized.CredentialKeys) != 1 || normalized.CredentialKeys[0] != "plan_type" {
		t.Fatalf("CredentialKeys = %#v, want []string{\"plan_type\"}", normalized.CredentialKeys)
	}
}

func TestNormalizeSub2APIAccountUsageTotals(t *testing.T) {
	row := sub2api.AccountRow{ID: 4, Platform: "openai"}
	usage := sub2api.AccountUsageRow{
		TotalRequests:       42,
		InputTokens:         100,
		OutputTokens:        25,
		CacheCreationTokens: 5,
		CacheReadTokens:     7,
		ActualCost:          0.42,
	}

	normalized := NormalizeSub2APIAccount(row, &usage)

	if normalized.Usage.TotalRequests != 42 {
		t.Fatalf("Usage.TotalRequests = %d, want 42", normalized.Usage.TotalRequests)
	}
	if normalized.Usage.TotalTokens != 137 {
		t.Fatalf("Usage.TotalTokens = %d, want 137", normalized.Usage.TotalTokens)
	}
}
