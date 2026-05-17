package sub2api

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
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

func TestRepositoryNilDatabaseReturnsError(t *testing.T) {
	_, err := NewRepository(nil).ListAccounts(context.Background())
	if err == nil {
		t.Fatal("ListAccounts() error = nil, want database nil error")
	}
	if !strings.Contains(err.Error(), "database is nil") {
		t.Fatalf("ListAccounts() error = %q, want database is nil", err.Error())
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
