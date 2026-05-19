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

func TestNormalizeSub2APIAccountKeepsSecretsOutOfPublicJSON(t *testing.T) {
	row := sub2api.AccountRow{
		ID:          42,
		Platform:    "openai",
		Type:        "plus",
		Status:      "active",
		Schedulable: true,
		Credentials: json.RawMessage(`{"access_token":"access-secret","refresh_token":"refresh-secret","api_key":"api-secret","email":"owner@example.com","plan_type":"pro","model_mapping":{"gpt-4":"gpt-4o"}}`),
	}

	account := NormalizeSub2APIAccount(row, nil)
	body, err := json.Marshal(account)
	if err != nil {
		t.Fatalf("marshal account: %v", err)
	}
	bodyText := string(body)

	for _, forbidden := range []string{"access_token", "refresh_token", "api_key", "email", "owner@example.com", "access-secret", "refresh-secret", "api-secret"} {
		if strings.Contains(bodyText, forbidden) {
			t.Fatalf("public JSON leaked %q in %s", forbidden, bodyText)
		}
	}
	if len(account.CredentialKeys) != 2 || account.CredentialKeys[0] != "model_mapping" || account.CredentialKeys[1] != "plan_type" {
		t.Fatalf("CredentialKeys = %#v, want []string{\"model_mapping\", \"plan_type\"}", account.CredentialKeys)
	}
}

func TestNormalizeSub2APIAccountQuotaWindows(t *testing.T) {
	start := time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC)
	end := start.Add(5 * time.Hour)
	row := sub2api.AccountRow{
		ID:                  7,
		Platform:            "claude",
		Type:                "max",
		Status:              "active",
		Schedulable:         true,
		SessionWindowStart:  &start,
		SessionWindowEnd:    &end,
		SessionWindowStatus: "active",
	}
	usage := sub2api.AccountUsageRow{AccountID: 7, InputTokens: 10, OutputTokens: 5, CacheCreationTokens: 2, CacheReadTokens: 3}

	account := NormalizeSub2APIAccount(row, &usage)

	if account.FiveHourWindow.RefreshAt == nil || !account.FiveHourWindow.RefreshAt.Equal(end) {
		t.Fatalf("FiveHourWindow.RefreshAt = %#v, want %s", account.FiveHourWindow.RefreshAt, end)
	}
	if account.FiveHourWindow.Consumed != 20 {
		t.Fatalf("FiveHourWindow.Consumed = %d, want 20", account.FiveHourWindow.Consumed)
	}
	if account.FiveHourWindow.Status != "active" {
		t.Fatalf("FiveHourWindow.Status = %q, want active", account.FiveHourWindow.Status)
	}
	if account.WeeklyWindow.Status != "unknown" {
		t.Fatalf("WeeklyWindow.Status = %q, want unknown", account.WeeklyWindow.Status)
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

func TestNormalizeSub2APIRankingsComputesShares(t *testing.T) {
	rows := []sub2api.RankingRow{
		{Name: "user-a", TotalRequests: 2, InputTokens: 10, OutputTokens: 10, ActualCost: 0.2},
		{Name: "user-b", TotalRequests: 1, InputTokens: 5, CacheCreationTokens: 5, ActualCost: 0.1},
	}

	rankings := NormalizeSub2APIRankings("user", rows)

	if len(rankings) != 2 {
		t.Fatalf("len(rankings) = %d, want 2", len(rankings))
	}
	if rankings[0].Dimension != "user" || rankings[0].Name != "user-a" || rankings[0].TotalTokens != 20 || rankings[0].Share != float64(20)/float64(30) {
		t.Fatalf("first ranking = %#v", rankings[0])
	}
	if rankings[1].CacheTokens != 5 || rankings[1].Share != float64(10)/float64(30) {
		t.Fatalf("second ranking = %#v", rankings[1])
	}
}

func TestNormalizeSub2APIRankingsUsesZeroShareWhenTotalIsZero(t *testing.T) {
	rankings := NormalizeSub2APIRankings("model", []sub2api.RankingRow{{Name: "gpt-4o"}})

	if len(rankings) != 1 {
		t.Fatalf("len(rankings) = %d, want 1", len(rankings))
	}
	if rankings[0].Share != 0 {
		t.Fatalf("Share = %f, want 0", rankings[0].Share)
	}
}

func TestNormalizeSub2APIEventsMapsUsageEventRows(t *testing.T) {
	createdAt := time.Date(2026, 5, 19, 10, 11, 12, 0, time.UTC)
	rows := []sub2api.UsageEventRow{{
		ID:                  9,
		CreatedAt:           createdAt,
		User:                "user-a",
		APIKey:              "key-a",
		Model:               "gpt-4o",
		RequestedModel:      "gpt-4",
		UpstreamModel:       "upstream-gpt-4o",
		AccountID:           7,
		AccountName:         "openai #7",
		Status:              "success",
		InputTokens:         10,
		OutputTokens:        5,
		CacheCreationTokens: 2,
		CacheReadTokens:     3,
		ActualCost:          0.12,
		DurationMS:          456,
	}}

	events := NormalizeSub2APIEvents(rows)

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	event := events[0]
	if event.ID != 9 || !event.CreatedAt.Equal(createdAt) || event.User != "user-a" || event.APIKey != "key-a" || event.Model != "gpt-4o" || event.RequestedModel != "gpt-4" || event.UpstreamModel != "upstream-gpt-4o" || event.AccountID != 7 || event.AccountName != "openai #7" || event.Status != "success" {
		t.Fatalf("event identity fields = %#v", event)
	}
	if event.InputTokens != 10 || event.OutputTokens != 5 || event.CacheTokens != 5 || event.TotalTokens != 20 || event.ActualCost != 0.12 || event.DurationMS != 456 {
		t.Fatalf("event usage fields = %#v", event)
	}
}
