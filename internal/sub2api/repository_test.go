package sub2api

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAccountCredentialKeys(t *testing.T) {
	row := AccountRow{
		Credentials: json.RawMessage(`{"email":"user@example.com","access_token":"access","refresh_token":"refresh","plan_type":"pro"}`),
	}

	got := row.CredentialKeys()
	want := []string{"access_token", "email", "plan_type", "refresh_token"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("CredentialKeys() = %v, want %v", got, want)
	}
}

func TestAccountCredentialKeysNilCases(t *testing.T) {
	tests := []struct {
		name        string
		credentials json.RawMessage
	}{
		{
			name: "empty credentials returns nil",
		},
		{
			name:        "invalid JSON returns nil",
			credentials: json.RawMessage(`{invalid-json`),
		},
		{
			name:        "empty JSON object returns nil",
			credentials: json.RawMessage(`{}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := AccountRow{Credentials: tt.credentials}

			if got := row.CredentialKeys(); got != nil {
				t.Fatalf("CredentialKeys() = %v, want nil", got)
			}
		})
	}
}

func TestAccountRowCredentialsNotSerialized(t *testing.T) {
	row := AccountRow{
		ID:          1,
		Name:        "test-account",
		Credentials: json.RawMessage(`{"access_token":"secret-access"}`),
	}

	body, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	serialized := string(body)
	if strings.Contains(serialized, "secret-access") {
		t.Fatalf("serialized AccountRow leaked credentials: %s", serialized)
	}
	if strings.Contains(serialized, "credentials") {
		t.Fatalf("serialized AccountRow included credentials field: %s", serialized)
	}
}

func TestClampDashboardDays(t *testing.T) {
	if got := ClampDashboardDays(9999); got != 90 {
		t.Fatalf("ClampDashboardDays() = %d, want 90", got)
	}
	if got := ClampDashboardDays(0); got != 1 {
		t.Fatalf("ClampDashboardDays() = %d, want 1", got)
	}
}

func TestClampDashboardHours(t *testing.T) {
	if got := ClampDashboardHours(9999); got != 2160 {
		t.Fatalf("ClampDashboardHours() = %d, want 2160", got)
	}
	if got := ClampDashboardHours(0); got != 24 {
		t.Fatalf("ClampDashboardHours() = %d, want 24", got)
	}
}

func TestClampDashboardLimit(t *testing.T) {
	if got := ClampDashboardLimit(9999, 20); got != 100 {
		t.Fatalf("ClampDashboardLimit() = %d, want 100", got)
	}
	if got := ClampDashboardLimit(0, 20); got != 20 {
		t.Fatalf("ClampDashboardLimit() = %d, want 20", got)
	}
}

func TestClampDashboardPage(t *testing.T) {
	if got := ClampDashboardPage(9999); got != 1000 {
		t.Fatalf("ClampDashboardPage() = %d, want 1000", got)
	}
	if got := ClampDashboardPage(0); got != 1 {
		t.Fatalf("ClampDashboardPage() = %d, want 1", got)
	}
}

func TestRepositoryNilDatabaseReturnsError(t *testing.T) {
	_, err := NewRepository(nil).ListAccounts(context.Background())
	if err == nil {
		t.Fatal("ListAccounts() error = nil, want database nil error")
	}
	if !strings.Contains(err.Error(), "database is nil") {
		t.Fatalf("ListAccounts() error = %q, want database is nil", err.Error())
	}
}

func TestRepositoryRankingsAndEventsUseSub2APIUsageLogColumns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE usage_logs (
			id INTEGER PRIMARY KEY,
			created_at DATETIME NOT NULL,
			user_id INTEGER,
			api_key_id INTEGER,
			account_id INTEGER,
			model TEXT,
			requested_model TEXT,
			upstream_model TEXT,
			input_tokens INTEGER,
			output_tokens INTEGER,
			cache_creation_tokens INTEGER,
			cache_read_tokens INTEGER,
			actual_cost REAL,
			duration_ms INTEGER,
			first_token_ms INTEGER
		)
	`).Error; err != nil {
		t.Fatalf("create usage_logs table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			email TEXT
		)
	`).Error; err != nil {
		t.Fatalf("create users table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE ops_error_logs (
			id INTEGER PRIMARY KEY,
			created_at DATETIME NOT NULL,
			user_id INTEGER,
			api_key_id INTEGER,
			account_id INTEGER,
			model TEXT,
			duration_ms INTEGER
		)
	`).Error; err != nil {
		t.Fatalf("create ops_error_logs table: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY,
			name TEXT
		)
	`).Error; err != nil {
		t.Fatalf("create accounts table: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO accounts (id, name) VALUES (?, ?)
	`, 3, "test-account-3").Error; err != nil {
		t.Fatalf("insert account: %v", err)
	}

	if err := db.Exec(`
		INSERT INTO users (id, email) VALUES (?, ?)
	`, 42, "user42@example.com").Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}
	createdAt := time.Now().Add(-time.Hour)
	if err := db.Exec(`
		INSERT INTO usage_logs (
			id, created_at, user_id, api_key_id, account_id, model, requested_model,
			upstream_model, input_tokens, output_tokens, cache_creation_tokens,
			cache_read_tokens, actual_cost, duration_ms
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, 1, createdAt, 42, 7, 3, "claude-sonnet", "sonnet", "claude-3-5-sonnet", 100, 20, 5, 10, 0.25, 1234).Error; err != nil {
		t.Fatalf("insert usage log: %v", err)
	}

	repository := NewRepository(db)

	rankings, err := repository.GetRankings(context.Background(), "user", time.Now().Add(-24*time.Hour), 5)
	if err != nil {
		t.Fatalf("GetRankings() error = %v", err)
	}
	if len(rankings) != 1 || rankings[0].Name != "user42@example.com" || rankings[0].TotalRequests != 1 {
		t.Fatalf("GetRankings() = %+v, want one user_id ranking", rankings)
	}

	accountRankings, err := repository.GetRankings(context.Background(), "account", time.Now().Add(-24*time.Hour), 5)
	if err != nil {
		t.Fatalf("GetRankings(account) error = %v", err)
	}
	if len(accountRankings) != 1 || accountRankings[0].Name != "test-account-3" {
		t.Fatalf("GetRankings(account) = %+v, want account name 'test-account-3'", accountRankings)
	}

	events, total, err := repository.GetEvents(context.Background(), time.Now().Add(-24*time.Hour), 1, 5)
	if err != nil {
		t.Fatalf("GetEvents() error = %v", err)
	}
	if total != 1 {
		t.Fatalf("GetEvents() total = %d, want 1", total)
	}
	if len(events) != 1 || events[0].User != "user42@example.com" || events[0].APIKey != "7" || events[0].AccountName != "3" {
		t.Fatalf("GetEvents() = %+v, want derived user/api key/account labels", events)
	}
}

func TestUsageSummaryTotalTokens(t *testing.T) {
	row := UsageOverviewRow{
		InputTokens:         10,
		OutputTokens:        20,
		CacheCreationTokens: 3,
		CacheReadTokens:     4,
	}

	if got, want := row.TotalTokens(), int64(37); got != want {
		t.Fatalf("TotalTokens() = %d, want %d", got, want)
	}
}

func TestModelUsageRowTotalTokens(t *testing.T) {
	row := ModelUsageRow{
		InputTokens:         1,
		OutputTokens:        2,
		CacheCreationTokens: 3,
		CacheReadTokens:     4,
	}

	if got, want := row.TotalTokens(), int64(10); got != want {
		t.Fatalf("TotalTokens() = %d, want %d", got, want)
	}
}

func TestAccountUsageRowTotalTokens(t *testing.T) {
	row := AccountUsageRow{
		InputTokens:         5,
		OutputTokens:        6,
		CacheCreationTokens: 7,
		CacheReadTokens:     8,
	}

	if got, want := row.TotalTokens(), int64(26); got != want {
		t.Fatalf("TotalTokens() = %d, want %d", got, want)
	}
}
