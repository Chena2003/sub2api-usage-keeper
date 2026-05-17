package service

import (
	"context"
	"fmt"
	"time"

	"sub2api-usage-keeper/internal/quota"
	"sub2api-usage-keeper/internal/sub2api"
)

type Sub2APIReader interface {
	ListAccounts(context.Context) ([]sub2api.AccountRow, error)
	GetAccountUsage(context.Context, time.Time) ([]sub2api.AccountUsageRow, error)
	GetDailyOverview(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	GetHourlyOverview(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	GetModelUsage(context.Context, time.Time, int) ([]sub2api.ModelUsageRow, error)
}

type Sub2APIDashboardService struct {
	reader Sub2APIReader
	now    func() time.Time
}

func NewSub2APIDashboardService(reader Sub2APIReader) *Sub2APIDashboardService {
	return &Sub2APIDashboardService{
		reader: reader,
		now:    time.Now,
	}
}

func (s *Sub2APIDashboardService) Accounts(ctx context.Context, days int) ([]quota.Sub2APIAccountQuota, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}

	accounts, err := s.reader.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	usageRows, err := s.reader.GetAccountUsage(ctx, sinceDays(s.currentTime(), days))
	if err != nil {
		return nil, err
	}

	usageByAccountID := make(map[int64]sub2api.AccountUsageRow, len(usageRows))
	for _, usage := range usageRows {
		usageByAccountID[usage.AccountID] = usage
	}

	result := make([]quota.Sub2APIAccountQuota, 0, len(accounts))
	for _, account := range accounts {
		var usage *sub2api.AccountUsageRow
		if row, ok := usageByAccountID[account.ID]; ok {
			usage = &row
		}
		result = append(result, quota.NormalizeSub2APIAccount(account, usage))
	}
	return result, nil
}

func (s *Sub2APIDashboardService) Overview(ctx context.Context, days int) (quota.Sub2APIOverview, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIOverview{}, err
	}

	accounts, err := s.reader.ListAccounts(ctx)
	if err != nil {
		return quota.Sub2APIOverview{}, err
	}
	rows, err := s.reader.GetDailyOverview(ctx, normalizeDays(days))
	if err != nil {
		return quota.Sub2APIOverview{}, err
	}

	overview := quota.Sub2APIOverview{AccountCount: int64(len(accounts))}
	for _, account := range accounts {
		if account.Status == "active" {
			overview.ActiveAccountCount++
		}
	}
	for _, row := range rows {
		cacheTokens := row.CacheCreationTokens + row.CacheReadTokens
		overview.TotalRequests += row.TotalRequests
		overview.InputTokens += row.InputTokens
		overview.OutputTokens += row.OutputTokens
		overview.CacheTokens += cacheTokens
		overview.ActualCost += row.ActualCost
		overview.AccountCost += row.AccountCost
		if row.ActiveUsers > overview.ActiveUsers {
			overview.ActiveUsers = row.ActiveUsers
		}
	}
	overview.TotalTokens = overview.InputTokens + overview.OutputTokens + overview.CacheTokens
	return overview, nil
}

func (s *Sub2APIDashboardService) Hourly(ctx context.Context, hours int) ([]sub2api.UsageOverviewRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s.reader.GetHourlyOverview(ctx, normalizeHours(hours))
}

func (s *Sub2APIDashboardService) Models(ctx context.Context, days int, limit int) ([]sub2api.ModelUsageRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s.reader.GetModelUsage(ctx, sinceDays(s.currentTime(), days), limit)
}

func (s *Sub2APIDashboardService) validate() error {
	if s == nil {
		return fmt.Errorf("sub2api dashboard service is nil")
	}
	if s.reader == nil {
		return fmt.Errorf("sub2api dashboard reader is nil")
	}
	return nil
}

func (s *Sub2APIDashboardService) currentTime() time.Time {
	if s.now == nil {
		return time.Now()
	}
	return s.now()
}

func sinceDays(now time.Time, days int) time.Time {
	return now.AddDate(0, 0, -normalizeDays(days))
}

func normalizeDays(days int) int {
	if days <= 0 {
		return 7
	}
	return days
}

func normalizeHours(hours int) int {
	if hours <= 0 {
		return 24
	}
	return hours
}
