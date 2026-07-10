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
	GetFiveHourAccountUsage(context.Context) ([]sub2api.AccountUsageRow, error)
	GetDailyOverview(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	GetHourlyOverview(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	GetModelUsage(context.Context, time.Time, int) ([]sub2api.ModelUsageRow, error)
	GetEvents(context.Context, time.Time, int, int) ([]sub2api.UsageEventRow, int64, error)
	GetRankings(context.Context, string, time.Time, int) ([]sub2api.RankingRow, error)
	GetRankingTrend(context.Context, string, time.Time, string, int) ([]sub2api.RankingTrendRow, error)
	GetHealthBlocks(context.Context, int) ([]sub2api.HealthBlockRow, error)
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

func (s *Sub2APIDashboardService) Accounts(ctx context.Context, since time.Time) ([]quota.Sub2APIAccountQuota, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}

	accounts, err := s.reader.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	usageRows, err := s.reader.GetAccountUsage(ctx, normalizeSince(s.currentTime(), since))
	if err != nil {
		return nil, err
	}
	fiveHourRows, err := s.reader.GetFiveHourAccountUsage(ctx)
	if err != nil {
		return nil, err
	}

	usageByAccountID := make(map[int64]sub2api.AccountUsageRow, len(usageRows))
	for _, usage := range usageRows {
		usageByAccountID[usage.AccountID] = usage
	}
	fiveHourByAccountID := make(map[int64]sub2api.AccountUsageRow, len(fiveHourRows))
	for _, usage := range fiveHourRows {
		fiveHourByAccountID[usage.AccountID] = usage
	}

	result := make([]quota.Sub2APIAccountQuota, 0, len(accounts))
	for _, account := range accounts {
		var usage *sub2api.AccountUsageRow
		if row, ok := usageByAccountID[account.ID]; ok {
			usage = &row
		}
		var fiveHourUsage *sub2api.AccountUsageRow
		if row, ok := fiveHourByAccountID[account.ID]; ok {
			fiveHourUsage = &row
		}
		result = append(result, quota.NormalizeSub2APIAccount(account, usage, fiveHourUsage))
	}
	return result, nil
}

func (s *Sub2APIDashboardService) AccountQuotas(ctx context.Context, since time.Time) ([]quota.Sub2APIAccountQuota, error) {
	return s.Accounts(ctx, since)
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

func (s *Sub2APIDashboardService) OverviewByHours(ctx context.Context, hours int) (quota.Sub2APIOverview, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIOverview{}, err
	}

	accounts, err := s.reader.ListAccounts(ctx)
	if err != nil {
		return quota.Sub2APIOverview{}, err
	}
	rows, err := s.reader.GetHourlyOverview(ctx, normalizeHours(hours))
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

func (s *Sub2APIDashboardService) Models(ctx context.Context, since time.Time, limit int) ([]sub2api.ModelUsageRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s.reader.GetModelUsage(ctx, normalizeSince(s.currentTime(), since), normalizeLimit(limit, 20))
}

func (s *Sub2APIDashboardService) Rankings(ctx context.Context, dimension string, since time.Time, limit int) ([]quota.Sub2APIRankingRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	dimension = normalizeSub2APIRankingDimension(dimension)
	limit = normalizeLimit(limit, 20)
	rows, err := s.reader.GetRankings(ctx, dimension, normalizeSince(s.currentTime(), since), limit)
	if err != nil {
		return nil, err
	}
	return quota.NormalizeSub2APIRankings(dimension, rows), nil
}

func (s *Sub2APIDashboardService) RankingTrend(ctx context.Context, dimension string, since time.Time, limit int) ([]quota.Sub2APIRankingTrendPoint, string, error) {
	if err := s.validate(); err != nil {
		return nil, "", err
	}
	dimension = normalizeSub2APIRankingDimension(dimension)
	limit = normalizeLimit(limit, 12)

	normalizedSince := normalizeSince(s.currentTime(), since)
	hours := int(s.currentTime().Sub(normalizedSince).Hours())

	granularity := "YYYY-MM-DD"
	granularityLabel := "day"
	if hours <= 48 {
		granularity = "YYYY-MM-DD HH24:00"
		granularityLabel = "hour"
	}

	rows, err := s.reader.GetRankingTrend(ctx, dimension, normalizedSince, granularity, limit)
	if err != nil {
		return nil, "", err
	}
	return quota.NormalizeSub2APIRankingTrend(dimension, rows), granularityLabel, nil
}

func (s *Sub2APIDashboardService) Events(ctx context.Context, since time.Time, page int, limit int) (quota.Sub2APIEventsResponse, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIEventsResponse{}, err
	}
	page = normalizePage(page)
	limit = normalizeLimit(limit, 100)
	rows, total, err := s.reader.GetEvents(ctx, normalizeSince(s.currentTime(), since), page, limit)
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

func (s *Sub2APIDashboardService) ServiceHealth(ctx context.Context, hours int) (quota.Sub2APIServiceHealth, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIServiceHealth{}, err
	}
	hours = sub2api.ClampDashboardHours(hours)
	blocks, err := s.reader.GetHealthBlocks(ctx, hours)
	if err != nil {
		return quota.Sub2APIServiceHealth{}, err
	}
	return quota.BuildSub2APIServiceHealth(blocks, hours), nil
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

func normalizeDays(days int) int {
	return sub2api.ClampDashboardDays(days)
}

func normalizeHours(hours int) int {
	return sub2api.ClampDashboardHours(hours)
}

func normalizePage(page int) int {
	return sub2api.ClampDashboardPage(page)
}

func normalizeLimit(limit int, defaultValue int) int {
	return sub2api.ClampDashboardLimit(limit, defaultValue)
}

// normalizeSince clamps a since timestamp to the allowed dashboard range.
// A zero since defaults to 7 days ago; the maximum lookback is 90 days.
func normalizeSince(now time.Time, since time.Time) time.Time {
	if since.IsZero() || since.After(now) {
		return now.AddDate(0, 0, -7)
	}
	maxSince := now.AddDate(0, 0, -sub2api.MaxDashboardDays)
	if since.Before(maxSince) {
		return maxSince
	}
	return since
}

func normalizeSub2APIRankingDimension(dimension string) string {
	switch dimension {
	case "api_key", "model", "account":
		return dimension
	default:
		return "user"
	}
}
