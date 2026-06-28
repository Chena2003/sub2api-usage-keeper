package quota

import (
	"encoding/json"
	"strconv"
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

	normalized := NormalizeSub2APIAccount(row, nil, nil)
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

	normalized := NormalizeSub2APIAccount(row, nil, nil)

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

	account := NormalizeSub2APIAccount(row, nil, nil)
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

	account := NormalizeSub2APIAccount(row, &usage, &usage)

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

func TestBuildFiveHourWindowReadsAnthropicExtraUtilization(t *testing.T) {
	start := time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC)
	end := start.Add(5 * time.Hour)
	weeklyReset := float64(start.Add(72 * time.Hour).Unix())
	row := sub2api.AccountRow{
		ID:                  7,
		Platform:            "claude",
		Type:                "max",
		Status:              "active",
		Schedulable:         true,
		SessionWindowStart:  &start,
		SessionWindowEnd:    &end,
		SessionWindowStatus: "active",
		Extra: json.RawMessage(`{
			"session_window_utilization": 0.42,
			"passive_usage_7d_utilization": 0.10,
			"passive_usage_7d_reset": ` + strconv.FormatFloat(weeklyReset, 'f', 0, 64) + `
		}`),
	}

	account := NormalizeSub2APIAccount(row, nil, nil)

	if account.FiveHourWindow.Utilization == nil || *account.FiveHourWindow.Utilization != 42 {
		t.Fatalf("FiveHourWindow.Utilization = %#v, want 42", account.FiveHourWindow.Utilization)
	}
	if account.WeeklyWindow.Utilization == nil || *account.WeeklyWindow.Utilization != 10 {
		t.Fatalf("WeeklyWindow.Utilization = %#v, want 10", account.WeeklyWindow.Utilization)
	}
	if account.WeeklyWindow.RefreshAt == nil {
		t.Fatalf("WeeklyWindow.RefreshAt = nil, want unix %v", weeklyReset)
	}
}

func TestBuildWindowsReadCodexExtraPercent(t *testing.T) {
	row := sub2api.AccountRow{
		ID:          9,
		Platform:    "openai",
		Type:        "oauth",
		Status:      "active",
		Schedulable: true,
		Extra: json.RawMessage(`{
			"codex_5h_used_percent": 73.5,
			"codex_5h_reset_at": "2026-05-19T15:00:00Z",
			"codex_7d_used_percent": 12,
			"codex_7d_reset_at": "2026-05-25T00:00:00Z"
		}`),
	}

	account := NormalizeSub2APIAccount(row, nil, nil)

	if account.FiveHourWindow.Utilization == nil || *account.FiveHourWindow.Utilization != 73.5 {
		t.Fatalf("FiveHourWindow.Utilization = %#v, want 73.5", account.FiveHourWindow.Utilization)
	}
	if account.WeeklyWindow.Utilization == nil || *account.WeeklyWindow.Utilization != 12 {
		t.Fatalf("WeeklyWindow.Utilization = %#v, want 12", account.WeeklyWindow.Utilization)
	}
	wantReset := time.Date(2026, 5, 19, 15, 0, 0, 0, time.UTC)
	if account.FiveHourWindow.RefreshAt == nil || !account.FiveHourWindow.RefreshAt.Equal(wantReset) {
		t.Fatalf("FiveHourWindow.RefreshAt = %#v, want %s", account.FiveHourWindow.RefreshAt, wantReset)
	}
}

func TestBuildFiveHourWindowFallsBackToSessionStatus(t *testing.T) {
	start := time.Date(2026, 5, 19, 8, 0, 0, 0, time.UTC)
	end := start.Add(5 * time.Hour)
	row := sub2api.AccountRow{
		ID:                  7,
		Platform:            "claude",
		Status:              "active",
		Schedulable:         true,
		SessionWindowStart:  &start,
		SessionWindowEnd:    &end,
		SessionWindowStatus: "allowed_warning",
	}

	account := NormalizeSub2APIAccount(row, nil, nil)

	if account.FiveHourWindow.Utilization == nil || *account.FiveHourWindow.Utilization != 80 {
		t.Fatalf("FiveHourWindow.Utilization = %#v, want 80 fallback", account.FiveHourWindow.Utilization)
	}
}

func TestDeriveSub2APIStatusPriority(t *testing.T) {
	past := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)

	cases := []struct {
		name         string
		row          sub2api.AccountRow
		wantDetail   string
		wantHasReset bool
		wantHasError bool
	}{
		{
			name:         "error status",
			row:          sub2api.AccountRow{Status: "error", Schedulable: true},
			wantDetail:   "error",
			wantHasError: true,
		},
		{
			name:         "rate limited active",
			row:          sub2api.AccountRow{Status: "active", Schedulable: true, RateLimitResetAt: &future},
			wantDetail:   "rate_limited",
			wantHasReset: true,
		},
		{
			name:       "overloaded",
			row:        sub2api.AccountRow{Status: "active", Schedulable: true, OverloadUntil: &future},
			wantDetail: "overloaded", wantHasReset: true,
		},
		{
			name:       "temp unschedulable",
			row:        sub2api.AccountRow{Status: "active", Schedulable: true, TempUnschedulableUntil: &future},
			wantDetail: "temp_unschedulable", wantHasReset: true,
		},
		{
			name:       "expired rate limit ignored -> paused",
			row:        sub2api.AccountRow{Status: "active", Schedulable: false, RateLimitResetAt: &past},
			wantDetail: "paused",
		},
		{
			name:       "inactive status",
			row:        sub2api.AccountRow{Status: "disabled", Schedulable: true},
			wantDetail: "inactive",
		},
		{
			name:       "active",
			row:        sub2api.AccountRow{Status: "active", Schedulable: true},
			wantDetail: "active",
		},
		{
			name:         "error message sets hasError but keeps schedule state",
			row:          sub2api.AccountRow{Status: "active", Schedulable: true, ErrorMessage: "upstream 500"},
			wantDetail:   "active",
			wantHasError: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			detail, resetAt, hasError := deriveSub2APIStatus(tc.row)
			if detail != tc.wantDetail {
				t.Fatalf("detail = %q, want %q", detail, tc.wantDetail)
			}
			if (resetAt != nil) != tc.wantHasReset {
				t.Fatalf("resetAt present = %v, want %v", resetAt != nil, tc.wantHasReset)
			}
			if hasError != tc.wantHasError {
				t.Fatalf("hasError = %v, want %v", hasError, tc.wantHasError)
			}
		})
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

	normalized := NormalizeSub2APIAccount(row, &usage, nil)

	if normalized.Usage.TotalRequests != 42 {
		t.Fatalf("Usage.TotalRequests = %d, want 42", normalized.Usage.TotalRequests)
	}
	if normalized.Usage.TotalTokens != 137 {
		t.Fatalf("Usage.TotalTokens = %d, want 137", normalized.Usage.TotalTokens)
	}
}

func TestNormalizeSub2APIRankingsComputesShares(t *testing.T) {
	rows := []sub2api.RankingRow{
		{Name: "gpt-4o", TotalRequests: 2, InputTokens: 10, OutputTokens: 10, ActualCost: 0.2},
		{Name: "claude-sonnet-4", TotalRequests: 1, InputTokens: 5, CacheCreationTokens: 5, ActualCost: 0.1},
	}

	rankings := NormalizeSub2APIRankings("model", rows)

	if len(rankings) != 2 {
		t.Fatalf("len(rankings) = %d, want 2", len(rankings))
	}
	if rankings[0].Dimension != "model" || rankings[0].Name != "gpt-4o" || rankings[0].TotalTokens != 20 || rankings[0].Share != float64(20)/float64(30) {
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

func TestNormalizeSub2APIRankingsMasksUserIdentifiers(t *testing.T) {
	rankings := NormalizeSub2APIRankings("user", []sub2api.RankingRow{{Name: "user@example.com", TotalRequests: 1, InputTokens: 10}})

	if len(rankings) != 1 {
		t.Fatalf("len(rankings) = %d, want 1", len(rankings))
	}
	if rankings[0].Name == "" {
		t.Fatal("Name is empty, want masked display value")
	}
	if rankings[0].Name == "user@example.com" {
		t.Fatalf("Name leaked raw user identifier: %#v", rankings[0])
	}

	body, err := json.Marshal(rankings)
	if err != nil {
		t.Fatalf("marshal rankings: %v", err)
	}
	if strings.Contains(string(body), "user@example.com") {
		t.Fatalf("serialized rankings leaked raw user identifier: %s", body)
	}
}

func TestNormalizeSub2APIRankingsMasksAPIKeyIdentifiers(t *testing.T) {
	rankings := NormalizeSub2APIRankings("api_key", []sub2api.RankingRow{{Name: "sk-live-secret", TotalRequests: 1, InputTokens: 10}})

	if len(rankings) != 1 {
		t.Fatalf("len(rankings) = %d, want 1", len(rankings))
	}
	if rankings[0].Name == "" {
		t.Fatal("Name is empty, want masked display value")
	}
	if rankings[0].Name == "sk-live-secret" {
		t.Fatalf("Name leaked raw API key identifier: %#v", rankings[0])
	}

	body, err := json.Marshal(rankings)
	if err != nil {
		t.Fatalf("marshal rankings: %v", err)
	}
	if strings.Contains(string(body), "sk-live-secret") {
		t.Fatalf("serialized rankings leaked raw API key identifier: %s", body)
	}
}

func TestNormalizeSub2APIEventsMasksUserIdentifiers(t *testing.T) {
	events := NormalizeSub2APIEvents([]sub2api.UsageEventRow{{User: "user@example.com"}})

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].User == "" {
		t.Fatal("User is empty, want masked display value")
	}
	if events[0].User == "user@example.com" {
		t.Fatalf("User leaked raw user identifier: %#v", events[0])
	}

	body, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	if strings.Contains(string(body), "user@example.com") {
		t.Fatalf("serialized events leaked raw user identifier: %s", body)
	}
}

func TestNormalizeSub2APIUserEmailMasksLocalPartAsFirstThreeAndLastTwo(t *testing.T) {
	rankings := NormalizeSub2APIRankings("user", []sub2api.RankingRow{{Name: "12345678901@qq.com"}})
	events := NormalizeSub2APIEvents([]sub2api.UsageEventRow{{User: "laijiachen@example.com"}})

	if len(rankings) != 1 || rankings[0].Name != "123******01@qq.com" {
		t.Fatalf("ranking user = %#v, want 123******01@qq.com", rankings)
	}
	if len(events) != 1 || events[0].User != "lai*****en@example.com" {
		t.Fatalf("event user = %#v, want lai*****en@example.com", events)
	}
	if strings.Contains(rankings[0].Name, "12345678901") || strings.Contains(events[0].User, "laijiachen") {
		t.Fatalf("masked email leaked raw local part: rankings=%#v events=%#v", rankings, events)
	}
}

func TestNormalizeSub2APIEventsMasksAPIKeyIdentifiers(t *testing.T) {
	events := NormalizeSub2APIEvents([]sub2api.UsageEventRow{{APIKey: "sk-live-secret"}})

	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].APIKey == "" {
		t.Fatal("APIKey is empty, want masked display value")
	}
	if events[0].APIKey == "sk-live-secret" {
		t.Fatalf("APIKey leaked raw identifier: %#v", events[0])
	}

	body, err := json.Marshal(events)
	if err != nil {
		t.Fatalf("marshal events: %v", err)
	}
	if strings.Contains(string(body), "sk-live-secret") {
		t.Fatalf("serialized events leaked raw API key identifier: %s", body)
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
	if event.ID != 9 || !event.CreatedAt.Equal(createdAt) || event.User == "" || event.User == "user-a" || event.APIKey == "" || event.APIKey == "key-a" || event.Model != "gpt-4o" || event.RequestedModel != "gpt-4" || event.UpstreamModel != "upstream-gpt-4o" || event.AccountID != 7 || event.AccountName != "openai #7" || event.Status != "success" {
		t.Fatalf("event identity fields = %#v", event)
	}
	if event.InputTokens != 10 || event.OutputTokens != 5 || event.CacheTokens != 5 || event.TotalTokens != 20 || event.ActualCost != 0.12 || event.DurationMS != 456 {
		t.Fatalf("event usage fields = %#v", event)
	}
}
