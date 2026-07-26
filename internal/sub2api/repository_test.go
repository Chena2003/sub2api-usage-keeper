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
			name: "empty credentials returns empty slice",
		},
		{
			name:        "invalid JSON returns empty slice",
			credentials: json.RawMessage(`{invalid-json`),
		},
		{
			name:        "empty JSON object returns empty slice",
			credentials: json.RawMessage(`{}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := AccountRow{Credentials: tt.credentials}

			if got := row.CredentialKeys(); len(got) != 0 {
				t.Fatalf("CredentialKeys() = %v, want empty slice", got)
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

	rankings, err := repository.GetRankings(context.Background(), "user", time.Now().Add(-24*time.Hour), time.Time{}, 5)
	if err != nil {
		t.Fatalf("GetRankings() error = %v", err)
	}
	if len(rankings) != 1 || rankings[0].Name != "user42@example.com" || rankings[0].TotalRequests != 1 {
		t.Fatalf("GetRankings() = %+v, want one user_id ranking", rankings)
	}

	accountRankings, err := repository.GetRankings(context.Background(), "account", time.Now().Add(-24*time.Hour), time.Time{}, 5)
	if err != nil {
		t.Fatalf("GetRankings(account) error = %v", err)
	}
	if len(accountRankings) != 1 || accountRankings[0].Name != "test-account-3" {
		t.Fatalf("GetRankings(account) = %+v, want account name 'test-account-3'", accountRankings)
	}

	events, total, err := repository.GetEvents(context.Background(), time.Now().Add(-24*time.Hour), time.Time{}, 1, 5)
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

func TestGetModelUsageAggregatesByModelOnly(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE usage_logs (
			id INTEGER PRIMARY KEY,
			created_at DATETIME NOT NULL,
			model TEXT,
			requested_model TEXT,
			upstream_model TEXT,
			input_tokens INTEGER,
			output_tokens INTEGER,
			cache_creation_tokens INTEGER,
			cache_read_tokens INTEGER,
			total_cost REAL,
			actual_cost REAL,
			duration_ms INTEGER
		)
	`).Error; err != nil {
		t.Fatalf("create usage_logs table: %v", err)
	}

	createdAt := time.Now().Add(-time.Hour)
	// Same model, differing requested_model / upstream_model. Previously this
	// split into 3 rows; now it must collapse into 1 aggregated row.
	insert := func(id int, requested, upstream string, in int) {
		if err := db.Exec(`
			INSERT INTO usage_logs (id, created_at, model, requested_model, upstream_model, input_tokens, output_tokens, cache_creation_tokens, cache_read_tokens, total_cost, actual_cost, duration_ms)
			VALUES (?, ?, 'claude-sonnet', ?, ?, ?, 0, 0, 0, 0, 0, 100)
		`, id, createdAt, requested, upstream, in).Error; err != nil {
			t.Fatalf("insert usage log %d: %v", id, err)
		}
	}
	insert(1, "sonnet", "claude-3-5-sonnet", 10)
	insert(2, "", "claude-3-5-sonnet", 20)
	insert(3, "sonnet", "", 30)

	repository := NewRepository(db)
	rows, err := repository.GetModelUsage(context.Background(), time.Now().Add(-24*time.Hour), time.Time{}, 20)
	if err != nil {
		t.Fatalf("GetModelUsage() error = %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("GetModelUsage() = %+v, want a single aggregated model row", rows)
	}
	if rows[0].Model != "claude-sonnet" || rows[0].TotalRequests != 3 || rows[0].InputTokens != 60 {
		t.Fatalf("unexpected aggregated row: %+v", rows[0])
	}
}

func TestGetOverviewByRangeFiltersBucketColumns(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE usage_dashboard_hourly (
			bucket_start DATETIME,
			total_requests INTEGER,
			input_tokens INTEGER,
			output_tokens INTEGER,
			cache_creation_tokens INTEGER,
			cache_read_tokens INTEGER,
			total_cost REAL,
			actual_cost REAL,
			account_cost REAL,
			total_duration_ms INTEGER,
			active_users INTEGER
		)
	`).Error; err != nil {
		t.Fatalf("create usage_dashboard_hourly: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE usage_dashboard_daily (
			bucket_date DATETIME,
			total_requests INTEGER,
			input_tokens INTEGER,
			output_tokens INTEGER,
			cache_creation_tokens INTEGER,
			cache_read_tokens INTEGER,
			total_cost REAL,
			actual_cost REAL,
			account_cost REAL,
			total_duration_ms INTEGER,
			active_users INTEGER
		)
	`).Error; err != nil {
		t.Fatalf("create usage_dashboard_daily: %v", err)
	}

	base := time.Date(2026, 5, 17, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 5; i++ {
		if err := db.Exec(`INSERT INTO usage_dashboard_hourly (bucket_start, total_requests) VALUES (?, ?)`, base.Add(time.Duration(i)*time.Hour), int64(i+1)).Error; err != nil {
			t.Fatalf("insert hourly: %v", err)
		}
	}
	for i := 0; i < 5; i++ {
		if err := db.Exec(`INSERT INTO usage_dashboard_daily (bucket_date, total_requests) VALUES (?, ?)`, base.AddDate(0, 0, i), int64(i+1)).Error; err != nil {
			t.Fatalf("insert daily: %v", err)
		}
	}

	repository := NewRepository(db)

	// Hourly: [base+1h, base+3h] must yield buckets at +1,+2,+3 => 3 rows.
	hourly, err := repository.GetHourlyOverviewByRange(context.Background(), base.Add(time.Hour), base.Add(3*time.Hour))
	if err != nil {
		t.Fatalf("GetHourlyOverviewByRange() error = %v", err)
	}
	if len(hourly) != 3 {
		t.Fatalf("GetHourlyOverviewByRange() returned %d rows, want 3", len(hourly))
	}

	// Zero until means no upper bound => all 5 rows from +1h onward except first.
	hourlyOpen, err := repository.GetHourlyOverviewByRange(context.Background(), base.Add(time.Hour), time.Time{})
	if err != nil {
		t.Fatalf("GetHourlyOverviewByRange(open) error = %v", err)
	}
	if len(hourlyOpen) != 4 {
		t.Fatalf("GetHourlyOverviewByRange(open) returned %d rows, want 4", len(hourlyOpen))
	}

	// Daily: [base+1d, base+2d] => 2 rows.
	daily, err := repository.GetDailyOverviewByRange(context.Background(), base.AddDate(0, 0, 1), base.AddDate(0, 0, 2))
	if err != nil {
		t.Fatalf("GetDailyOverviewByRange() error = %v", err)
	}
	if len(daily) != 2 {
		t.Fatalf("GetDailyOverviewByRange() returned %d rows, want 2", len(daily))
	}
}

func TestWithDSNTimeZone(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
		tz   string
		want string
	}{
		{
			name: "url dsn without timezone gets one",
			dsn:  "postgres://u:p@h:5432/db?sslmode=disable",
			tz:   "Asia/Shanghai",
			want: "TimeZone=Asia%2FShanghai",
		},
		{
			name: "keyword dsn without timezone gets one",
			dsn:  "host=h port=5432 dbname=db",
			tz:   "Asia/Shanghai",
			want: "TimeZone=Asia/Shanghai",
		},
		{
			name: "empty timezone leaves dsn untouched",
			dsn:  "host=h port=5432 dbname=db",
			tz:   "",
			want: "host=h port=5432 dbname=db",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := withDSNTimeZone(tt.dsn, tt.tz)
			if !strings.Contains(got, tt.want) {
				t.Fatalf("withDSNTimeZone(%q, %q) = %q, want to contain %q", tt.dsn, tt.tz, got, tt.want)
			}
		})
	}

	// Existing TimeZone must not be duplicated / overridden.
	dsn := "postgres://u:p@h:5432/db?TimeZone=UTC"
	if got := withDSNTimeZone(dsn, "Asia/Shanghai"); got != dsn {
		t.Fatalf("withDSNTimeZone should not override existing TimeZone, got %q", got)
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
