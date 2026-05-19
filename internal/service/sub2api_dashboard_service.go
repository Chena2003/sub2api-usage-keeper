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
	GetEvents(context.Context, int, int) ([]sub2api.UsageEventRow, int64, error)
	GetRankings(context.Context, string, time.Time, int) ([]sub2api.RankingRow, error)
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

func (s *Sub2APIDashboardService) AccountQuotas(ctx context.Context, days int) ([]quota.Sub2APIAccountQuota, error) {
	return s.Accounts(ctx, days)
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

func (s *Sub2APIDashboardService) Rankings(ctx context.Context, dimension string, days int, limit int) ([]quota.Sub2APIRankingRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	dimension = normalizeSub2APIRankingDimension(dimension)
	limit = normalizeLimit(limit, 20)
	rows, err := s.reader.GetRankings(ctx, dimension, sinceDays(s.currentTime(), days), limit)
	if err != nil {
		return nil, err
	}
	return quota.NormalizeSub2APIRankings(dimension, rows), nil
}

func (s *Sub2APIDashboardService) Events(ctx context.Context, page int, limit int) (quota.Sub2APIEventsResponse, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIEventsResponse{}, err
	}
	page = normalizePage(page)
	limit = normalizeLimit(limit, 100)
	rows, total, err := s.reader.GetEvents(ctx, page, limit)
	if err != nil {
		return quota.Sub2APIEventsResponse{}, err
	}
	return quota.Sub2APIEventsResponse{
		Events: quota.NormalizeSub2APIEvents(rows),
		Total:  total,
		Page:   page,
		Limit:  limit,
	}, nil
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

func normalizePage(page int) int {
	if page <= 0 {
		return 1
	}
	return page
}

func normalizeLimit(limit int, defaultValue int) int {
	if limit <= 0 {
		return defaultValue
	}
	return limit
}

func normalizeSub2APIRankingDimension(dimension string) string {
	switch dimension {
	case "api_key", "model", "account":
		return dimension
	default:
		return "user"
	}
}
