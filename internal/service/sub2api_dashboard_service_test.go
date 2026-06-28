package service

import (
	"context"
	"math"
	"testing"
	"time"

	"sub2api-usage-keeper/internal/sub2api"
)

type fakeSub2APIReader struct {
	accounts             []sub2api.AccountRow
	accountUsage         []sub2api.AccountUsageRow
	dailyOverview        []sub2api.UsageOverviewRow
	hourlyOverview       []sub2api.UsageOverviewRow
	modelUsage           []sub2api.ModelUsageRow
	rankings             []sub2api.RankingRow
	events               []sub2api.UsageEventRow
	eventsTotal          int64
	lastDailyDays        int
	lastHourlyHours      int
	lastAccountSince     time.Time
	lastModelSince       time.Time
	lastModelLimit       int
	lastRankingDimension string
	lastRankingSince     time.Time
	lastRankingLimit     int
	lastEventsPage       int
	lastEventsLimit      int
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

func (f *fakeSub2APIReader) GetRankings(_ context.Context, dimension string, since time.Time, limit int) ([]sub2api.RankingRow, error) {
	f.lastRankingDimension = dimension
	f.lastRankingSince = since
	f.lastRankingLimit = limit
	return f.rankings, nil
}

func (f *fakeSub2APIReader) GetEvents(_ context.Context, page int, limit int) ([]sub2api.UsageEventRow, int64, error) {
	f.lastEventsPage = page
	f.lastEventsLimit = limit
	return f.events, f.eventsTotal, nil
}

func (f *fakeSub2APIReader) GetFiveHourAccountUsage(_ context.Context) ([]sub2api.AccountUsageRow, error) {
	return f.accountUsage, nil
}

func (f *fakeSub2APIReader) GetHealthBlocks(_ context.Context, _ int) ([]sub2api.HealthBlockRow, error) {
	return nil, nil
}

func (f *fakeSub2APIReader) GetRankingTrend(_ context.Context, _ string, _ time.Time, _ string, _ int) ([]sub2api.RankingTrendRow, error) {
	return nil, nil
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

func TestSub2APIDashboardAccountQuotasDefaultsToSevenDays(t *testing.T) {
	reader := &fakeSub2APIReader{}
	fixedNow := time.Date(2026, time.May, 17, 12, 30, 0, 0, time.UTC)
	service := NewSub2APIDashboardService(reader)
	service.now = func() time.Time { return fixedNow }

	_, err := service.AccountQuotas(context.Background(), 0)
	if err != nil {
		t.Fatalf("AccountQuotas returned error: %v", err)
	}
	if !reader.lastAccountSince.Equal(fixedNow.AddDate(0, 0, -7)) {
		t.Fatalf("expected account usage since %v, got %v", fixedNow.AddDate(0, 0, -7), reader.lastAccountSince)
	}
}

func TestSub2APIDashboardRankingsNormalizesDefaults(t *testing.T) {
	reader := &fakeSub2APIReader{
		rankings: []sub2api.RankingRow{{
			Name:                "user@example.com",
			TotalRequests:       2,
			InputTokens:         10,
			OutputTokens:        5,
			CacheCreationTokens: 3,
			CacheReadTokens:     2,
		}},
	}
	fixedNow := time.Date(2026, time.May, 17, 12, 30, 0, 0, time.UTC)
	service := NewSub2APIDashboardService(reader)
	service.now = func() time.Time { return fixedNow }

	rankings, err := service.Rankings(context.Background(), "", 0, 0)
	if err != nil {
		t.Fatalf("Rankings returned error: %v", err)
	}
	if reader.lastRankingDimension != "user" {
		t.Fatalf("expected ranking dimension user, got %q", reader.lastRankingDimension)
	}
	if !reader.lastRankingSince.Equal(fixedNow.AddDate(0, 0, -7)) {
		t.Fatalf("expected ranking since %v, got %v", fixedNow.AddDate(0, 0, -7), reader.lastRankingSince)
	}
	if reader.lastRankingLimit != 20 {
		t.Fatalf("expected ranking limit 20, got %d", reader.lastRankingLimit)
	}
	if len(rankings) != 1 || rankings[0].Dimension != "user" || rankings[0].TotalTokens != 20 {
		t.Fatalf("unexpected rankings: %#v", rankings)
	}
}

func TestSub2APIDashboardEventsNormalizesDefaults(t *testing.T) {
	reader := &fakeSub2APIReader{
		events: []sub2api.UsageEventRow{{
			ID:                  42,
			User:                "user@example.com",
			InputTokens:         10,
			OutputTokens:        5,
			CacheCreationTokens: 3,
			CacheReadTokens:     2,
		}},
		eventsTotal: 5,
	}

	events, err := NewSub2APIDashboardService(reader).Events(context.Background(), 0, 0)
	if err != nil {
		t.Fatalf("Events returned error: %v", err)
	}
	if reader.lastEventsPage != 1 {
		t.Fatalf("expected events page 1, got %d", reader.lastEventsPage)
	}
	if reader.lastEventsLimit != 100 {
		t.Fatalf("expected events limit 100, got %d", reader.lastEventsLimit)
	}
	if events.Page != 1 || events.Limit != 100 || events.Total != 5 || len(events.Events) != 1 || events.Events[0].TotalTokens != 20 {
		t.Fatalf("unexpected events response: %#v", events)
	}
}

func TestSub2APIDashboardClampsMaximums(t *testing.T) {
	reader := &fakeSub2APIReader{}
	fixedNow := time.Date(2026, time.May, 17, 12, 30, 0, 0, time.UTC)
	service := NewSub2APIDashboardService(reader)
	service.now = func() time.Time { return fixedNow }

	if _, err := service.AccountQuotas(context.Background(), 9999); err != nil {
		t.Fatalf("AccountQuotas returned error: %v", err)
	}
	if !reader.lastAccountSince.Equal(fixedNow.AddDate(0, 0, -90)) {
		t.Fatalf("expected account usage since %v, got %v", fixedNow.AddDate(0, 0, -90), reader.lastAccountSince)
	}

	if _, err := service.Overview(context.Background(), 9999); err != nil {
		t.Fatalf("Overview returned error: %v", err)
	}
	if reader.lastDailyDays != 90 {
		t.Fatalf("expected daily overview days 90, got %d", reader.lastDailyDays)
	}

	if _, err := service.Hourly(context.Background(), 9999); err != nil {
		t.Fatalf("Hourly returned error: %v", err)
	}
	if reader.lastHourlyHours != 168 {
		t.Fatalf("expected hourly overview hours 168, got %d", reader.lastHourlyHours)
	}

	if _, err := service.Models(context.Background(), 9999, 9999); err != nil {
		t.Fatalf("Models returned error: %v", err)
	}
	if !reader.lastModelSince.Equal(fixedNow.AddDate(0, 0, -90)) {
		t.Fatalf("expected model usage since %v, got %v", fixedNow.AddDate(0, 0, -90), reader.lastModelSince)
	}
	if reader.lastModelLimit != 100 {
		t.Fatalf("expected model limit 100, got %d", reader.lastModelLimit)
	}

	if _, err := service.Rankings(context.Background(), "user", 9999, 9999); err != nil {
		t.Fatalf("Rankings returned error: %v", err)
	}
	if !reader.lastRankingSince.Equal(fixedNow.AddDate(0, 0, -90)) {
		t.Fatalf("expected ranking since %v, got %v", fixedNow.AddDate(0, 0, -90), reader.lastRankingSince)
	}
	if reader.lastRankingLimit != 100 {
		t.Fatalf("expected ranking limit 100, got %d", reader.lastRankingLimit)
	}

	if _, err := service.Events(context.Background(), 9999, 9999); err != nil {
		t.Fatalf("Events returned error: %v", err)
	}
	if reader.lastEventsPage != 1000 {
		t.Fatalf("expected events page 1000, got %d", reader.lastEventsPage)
	}
	if reader.lastEventsLimit != 100 {
		t.Fatalf("expected events limit 100, got %d", reader.lastEventsLimit)
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
