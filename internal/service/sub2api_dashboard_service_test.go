package service

import (
	"context"
	"math"
	"testing"
	"time"

	"sub2api-usage-keeper/internal/sub2api"
)

type fakeSub2APIReader struct {
	accounts          []sub2api.AccountRow
	accountUsage      []sub2api.AccountUsageRow
	dailyOverview     []sub2api.UsageOverviewRow
	hourlyOverview    []sub2api.UsageOverviewRow
	modelUsage        []sub2api.ModelUsageRow
	lastDailyDays     int
	lastHourlyHours   int
	lastAccountSince  time.Time
	lastModelSince    time.Time
	lastModelLimit    int
}

func (f *fakeSub2APIReader) ListAccounts(context.Context) ([]sub2api.AccountRow, error) {
	return f.accounts, nil
}

func (f *fakeSub2APIReader) GetAccountUsage(_ context.Context, since time.Time) ([]sub2api.AccountUsageRow, error) {
	f.lastAccountSince = since
	return f.accountUsage, nil
}

func (f *fakeSub2APIReader) GetDailyOverview(_ context.Context, days int) ([]sub2api.UsageOverviewRow, error) {
	f.lastDailyDays = days
	return f.dailyOverview, nil
}

func (f *fakeSub2APIReader) GetHourlyOverview(_ context.Context, hours int) ([]sub2api.UsageOverviewRow, error) {
	f.lastHourlyHours = hours
	return f.hourlyOverview, nil
}

func (f *fakeSub2APIReader) GetModelUsage(_ context.Context, since time.Time, limit int) ([]sub2api.ModelUsageRow, error) {
	f.lastModelSince = since
	f.lastModelLimit = limit
	return f.modelUsage, nil
}

func TestSub2APIDashboardAccountsMergeUsage(t *testing.T) {
	reader := &fakeSub2APIReader{
		accounts: []sub2api.AccountRow{{
			ID:          4,
			Platform:    "openai",
			Type:        "oauth",
			Status:      "active",
			Schedulable: true,
		}},
		accountUsage: []sub2api.AccountUsageRow{{
			AccountID:     4,
			TotalRequests: 10,
			InputTokens:   20,
			OutputTokens:  30,
		}},
	}

	accounts, err := NewSub2APIDashboardService(reader).Accounts(context.Background(), 7)
	if err != nil {
		t.Fatalf("Accounts returned error: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(accounts))
	}
	if accounts[0].ID != 4 || accounts[0].Provider != "openai" || accounts[0].AccountType != "oauth" {
		t.Fatalf("unexpected account: %#v", accounts[0])
	}
	if accounts[0].Usage.TotalRequests != 10 || accounts[0].Usage.TotalTokens != 50 {
		t.Fatalf("unexpected account usage: %#v", accounts[0].Usage)
	}
}

func TestSub2APIDashboardOverviewSumsDailyRows(t *testing.T) {
	reader := &fakeSub2APIReader{
		accounts: []sub2api.AccountRow{
			{ID: 1, Status: "active"},
			{ID: 2, Status: "error"},
		},
		dailyOverview: []sub2api.UsageOverviewRow{
			{
				TotalRequests:       3,
				InputTokens:         10,
				OutputTokens:        20,
				CacheCreationTokens: 1,
				CacheReadTokens:     2,
				ActualCost:          0.5,
				AccountCost:         0.3,
				ActiveUsers:         2,
			},
			{
				TotalRequests:       7,
				InputTokens:         30,
				OutputTokens:        40,
				CacheCreationTokens: 3,
				CacheReadTokens:     4,
				ActualCost:          1.5,
				AccountCost:         1.2,
				ActiveUsers:         5,
			},
		},
	}

	overview, err := NewSub2APIDashboardService(reader).Overview(context.Background(), 7)
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if overview.AccountCount != 2 {
		t.Fatalf("expected account count 2, got %d", overview.AccountCount)
	}
	if overview.ActiveAccountCount != 1 {
		t.Fatalf("expected active account count 1, got %d", overview.ActiveAccountCount)
	}
	if overview.TotalRequests != 10 {
		t.Fatalf("expected total requests 10, got %d", overview.TotalRequests)
	}
	if overview.TotalTokens != 110 {
		t.Fatalf("expected total tokens 110, got %d", overview.TotalTokens)
	}
	if overview.InputTokens != 40 || overview.OutputTokens != 60 || overview.CacheTokens != 10 {
		t.Fatalf("unexpected token totals: %#v", overview)
	}
	assertFloatEqual(t, overview.ActualCost, 2.0)
	assertFloatEqual(t, overview.AccountCost, 1.5)
	if overview.ActiveUsers != 5 {
		t.Fatalf("expected active users max 5, got %d", overview.ActiveUsers)
	}
}

func TestSub2APIDashboardOverviewDefaultsToSevenDays(t *testing.T) {
	reader := &fakeSub2APIReader{}

	_, err := NewSub2APIDashboardService(reader).Overview(context.Background(), 0)
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if reader.lastDailyDays != 7 {
		t.Fatalf("expected daily overview days 7, got %d", reader.lastDailyDays)
	}
}

func TestSub2APIDashboardHourlyDefaultsToTwentyFourHours(t *testing.T) {
	reader := &fakeSub2APIReader{}

	_, err := NewSub2APIDashboardService(reader).Hourly(context.Background(), 0)
	if err != nil {
		t.Fatalf("Hourly returned error: %v", err)
	}
	if reader.lastHourlyHours != 24 {
		t.Fatalf("expected hourly overview hours 24, got %d", reader.lastHourlyHours)
	}
}

func TestSub2APIDashboardModelsDefaultsToSevenDays(t *testing.T) {
	reader := &fakeSub2APIReader{}
	fixedNow := time.Date(2026, time.May, 17, 12, 30, 0, 0, time.UTC)
	service := NewSub2APIDashboardService(reader)
	service.now = func() time.Time { return fixedNow }

	_, err := service.Models(context.Background(), 0, 20)
	if err != nil {
		t.Fatalf("Models returned error: %v", err)
	}
	if !reader.lastModelSince.Equal(fixedNow.AddDate(0, 0, -7)) {
		t.Fatalf("expected model usage since %v, got %v", fixedNow.AddDate(0, 0, -7), reader.lastModelSince)
	}
	if reader.lastModelLimit != 20 {
		t.Fatalf("expected model usage limit 20, got %d", reader.lastModelLimit)
	}
}

func TestSub2APIDashboardValidateDoesNotSetDefaultNow(t *testing.T) {
	service := &Sub2APIDashboardService{reader: &fakeSub2APIReader{}}

	_, err := service.Overview(context.Background(), 7)
	if err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if service.now != nil {
		t.Fatal("expected validate to leave now unset")
	}
}

func assertFloatEqual(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
