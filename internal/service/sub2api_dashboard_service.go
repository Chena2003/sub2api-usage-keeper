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
	GetAccountUsage(context.Context, time.Time, time.Time) ([]sub2api.AccountUsageRow, error)
	GetFiveHourAccountUsage(context.Context) ([]sub2api.AccountUsageRow, error)
	GetDailyOverview(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	GetHourlyOverview(context.Context, int) ([]sub2api.UsageOverviewRow, error)
	GetDailyOverviewByRange(context.Context, time.Time, time.Time) ([]sub2api.UsageOverviewRow, error)
	GetHourlyOverviewByRange(context.Context, time.Time, time.Time) ([]sub2api.UsageOverviewRow, error)
	GetModelUsage(context.Context, time.Time, time.Time, int) ([]sub2api.ModelUsageRow, error)
	GetEvents(context.Context, time.Time, time.Time, int, int) ([]sub2api.UsageEventRow, int64, error)
	GetRankings(context.Context, string, time.Time, time.Time, int) ([]sub2api.RankingRow, error)
	GetRankingTrend(context.Context, string, time.Time, time.Time, string, int) ([]sub2api.RankingTrendRow, error)
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

func (s *Sub2APIDashboardService) Accounts(ctx context.Context, since time.Time, until time.Time) ([]quota.Sub2APIAccountQuota, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}

	accounts, err := s.reader.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	usageRows, err := s.reader.GetAccountUsage(ctx, normalizedSince, normalizedUntil)
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

	return accumulateOverview(accounts, rows), nil
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

	return accumulateOverview(accounts, rows), nil
}

func (s *Sub2APIDashboardService) Hourly(ctx context.Context, hours int) ([]sub2api.UsageOverviewRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	return s.reader.GetHourlyOverview(ctx, normalizeHours(hours))
}

// OverviewByRange aggregates the overview for an explicit [since, until] window.
// It picks the hourly aggregation table for short windows (<= 48h) and the daily
// table otherwise, mirroring how RankingTrend chooses its granularity.
func (s *Sub2APIDashboardService) OverviewByRange(ctx context.Context, since time.Time, until time.Time) (quota.Sub2APIOverview, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIOverview{}, err
	}

	accounts, err := s.reader.ListAccounts(ctx)
	if err != nil {
		return quota.Sub2APIOverview{}, err
	}

	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	hours := int(s.currentTime().Sub(normalizedSince).Hours())

	var rows []sub2api.UsageOverviewRow
	if hours <= 48 {
		rows, err = s.reader.GetHourlyOverviewByRange(ctx, normalizedSince, normalizedUntil)
	} else {
		rows, err = s.reader.GetDailyOverviewByRange(ctx, normalizedSince, normalizedUntil)
	}
	if err != nil {
		return quota.Sub2APIOverview{}, err
	}

	return accumulateOverview(accounts, rows), nil
}

// HourlyByRange returns hourly overview points for an explicit [since, until] window.
func (s *Sub2APIDashboardService) HourlyByRange(ctx context.Context, since time.Time, until time.Time) ([]sub2api.UsageOverviewRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	return s.reader.GetHourlyOverviewByRange(ctx, normalizedSince, normalizedUntil)
}

// accumulateOverview sums per-bucket usage rows into a single overview, taking the
// MAX active users across buckets and recomputing TotalTokens.
func accumulateOverview(accounts []sub2api.AccountRow, rows []sub2api.UsageOverviewRow) quota.Sub2APIOverview {
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
	return overview
}

func (s *Sub2APIDashboardService) Models(ctx context.Context, since time.Time, until time.Time, limit int) ([]sub2api.ModelUsageRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	return s.reader.GetModelUsage(ctx, normalizedSince, normalizedUntil, normalizeLimit(limit, 20))
}

func (s *Sub2APIDashboardService) Rankings(ctx context.Context, dimension string, since time.Time, until time.Time, limit int) ([]quota.Sub2APIRankingRow, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	dimension = normalizeSub2APIRankingDimension(dimension)
	limit = normalizeLimit(limit, 20)
	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	rows, err := s.reader.GetRankings(ctx, dimension, normalizedSince, normalizedUntil, limit)
	if err != nil {
		return nil, err
	}
	return quota.NormalizeSub2APIRankings(dimension, rows), nil
}

func (s *Sub2APIDashboardService) RankingTrend(ctx context.Context, dimension string, since time.Time, until time.Time, limit int) ([]quota.Sub2APIRankingTrendPoint, string, error) {
	if err := s.validate(); err != nil {
		return nil, "", err
	}
	dimension = normalizeSub2APIRankingDimension(dimension)
	limit = normalizeLimit(limit, 12)

	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	hours := int(s.currentTime().Sub(normalizedSince).Hours())

	granularity := "YYYY-MM-DD"
	granularityLabel := "day"
	if hours <= 48 {
		granularity = "YYYY-MM-DD HH24:00"
		granularityLabel = "hour"
	}

	rows, err := s.reader.GetRankingTrend(ctx, dimension, normalizedSince, normalizedUntil, granularity, limit)
	if err != nil {
		return nil, "", err
	}
	return quota.NormalizeSub2APIRankingTrend(dimension, rows), granularityLabel, nil
}

func (s *Sub2APIDashboardService) Events(ctx context.Context, since time.Time, until time.Time, page int, limit int) (quota.Sub2APIEventsResponse, error) {
	if err := s.validate(); err != nil {
		return quota.Sub2APIEventsResponse{}, err
	}
	page = normalizePage(page)
	limit = normalizeLimit(limit, 100)
	normalizedSince, normalizedUntil := normalizeTimeRange(s.currentTime(), since, until)
	rows, total, err := s.reader.GetEvents(ctx, normalizedSince, normalizedUntil, page, limit)
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
	// The health grid is always "today + 6 prior complete calendar days" = 168h,
	// 7 rows × 96 cols × 15min. It is intentionally decoupled from the page range
	// (the incoming hours is ignored) so the grid is fully occupied and 15-min SQL
	// buckets align 1:1 with grid slots.
	const healthWindowHours = 168
	blocks, err := s.reader.GetHealthBlocks(ctx, healthWindowHours)
	if err != nil {
		return quota.Sub2APIServiceHealth{}, err
	}
	return quota.BuildSub2APIServiceHealth(blocks, healthWindowHours), nil
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
// A since slightly in the future (within 1 hour) is tolerated to account for
// timezone differences between client and server (e.g. "today" at local midnight
// may be slightly ahead of server now). Anything further in the future falls
// back to the 7-day default.
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

// normalizeTimeRange clamps both since and until timestamps. until is returned
// as-is when non-zero and valid; a zero until means "no upper bound".
func normalizeTimeRange(now time.Time, since time.Time, until time.Time) (time.Time, time.Time) {
	// Allow up to 1 hour of timezone skew before falling back to default.
	if since.IsZero() || since.After(now.Add(time.Hour)) {
		since = now.AddDate(0, 0, -7)
	}
	maxSince := now.AddDate(0, 0, -sub2api.MaxDashboardDays)
	if since.Before(maxSince) {
		since = maxSince
	}
	// Clamp until to not be before since; leave zero as "no upper bound".
	if !until.IsZero() && until.Before(since) {
		until = time.Time{}
	}
	return since, until
}

func normalizeSub2APIRankingDimension(dimension string) string {
	switch dimension {
	case "api_key", "model", "account":
		return dimension
	default:
		return "user"
	}
}
