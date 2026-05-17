package sub2api

import (
	"context"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func Open(databaseURL string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open sub2api postgres: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("configure sub2api postgres: %w", err)
	}

	sqlDB.SetMaxOpenConns(3)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	return NewRepository(db), nil
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) database() (*gorm.DB, error) {
	if r == nil {
		return nil, fmt.Errorf("sub2api repository is nil")
	}
	if r.db == nil {
		return nil, fmt.Errorf("sub2api repository database is nil")
	}
	return r.db, nil
}

func (r *Repository) Close() error {
	if r == nil || r.db == nil {
		return nil
	}

	sqlDB, err := r.db.DB()
	if err != nil {
		return fmt.Errorf("configure sub2api postgres close: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("close sub2api postgres: %w", err)
	}
	return nil
}

func (r *Repository) ListAccounts(ctx context.Context) ([]AccountRow, error) {
	db, err := r.database()
	if err != nil {
		return nil, fmt.Errorf("list sub2api accounts: %w", err)
	}

	var rows []AccountRow
	err = db.WithContext(ctx).Raw(`
		SELECT
			id,
			name,
			platform,
			type,
			status,
			schedulable,
			credentials,
			rate_limited_at,
			rate_limit_reset_at,
			session_window_start,
			session_window_end,
			session_window_status,
			expires_at,
			rate_multiplier,
			load_factor,
			last_used_at,
			updated_at
		FROM accounts
		WHERE deleted_at IS NULL
		ORDER BY platform, id
	`).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("list sub2api accounts: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetDailyOverview(ctx context.Context, days int) ([]UsageOverviewRow, error) {
	if days <= 0 {
		days = 7
	}

	db, err := r.database()
	if err != nil {
		return nil, fmt.Errorf("get sub2api daily overview: %w", err)
	}

	var rows []UsageOverviewRow
	err = db.WithContext(ctx).Raw(`
		SELECT
			bucket_date,
			total_requests,
			input_tokens,
			output_tokens,
			cache_creation_tokens,
			cache_read_tokens,
			total_cost,
			actual_cost,
			account_cost,
			total_duration_ms,
			active_users
		FROM usage_dashboard_daily
		WHERE bucket_date >= current_date - (?::int - 1)
		ORDER BY bucket_date ASC
	`, days).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get sub2api daily overview: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetHourlyOverview(ctx context.Context, hours int) ([]UsageOverviewRow, error) {
	if hours <= 0 {
		hours = 24
	}

	db, err := r.database()
	if err != nil {
		return nil, fmt.Errorf("get sub2api hourly overview: %w", err)
	}

	var rows []UsageOverviewRow
	err = db.WithContext(ctx).Raw(`
		SELECT
			bucket_start,
			total_requests,
			input_tokens,
			output_tokens,
			cache_creation_tokens,
			cache_read_tokens,
			total_cost,
			actual_cost,
			account_cost,
			total_duration_ms,
			active_users
		FROM usage_dashboard_hourly
		WHERE bucket_start >= now() - (?::int * interval '1 hour')
		ORDER BY bucket_start ASC
	`, hours).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get sub2api hourly overview: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetModelUsage(ctx context.Context, since time.Time, limit int) ([]ModelUsageRow, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	db, err := r.database()
	if err != nil {
		return nil, fmt.Errorf("get sub2api model usage: %w", err)
	}

	var rows []ModelUsageRow
	err = db.WithContext(ctx).Raw(`
		SELECT
			model,
			requested_model,
			upstream_model,
			COUNT(*) AS total_requests,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) AS cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) AS cache_read_tokens,
			COALESCE(SUM(total_cost), 0) AS total_cost,
			COALESCE(SUM(actual_cost), 0) AS actual_cost,
			COALESCE(AVG(duration_ms), 0) AS average_duration_ms
		FROM usage_logs
		WHERE created_at >= ?
		GROUP BY model, requested_model, upstream_model
		ORDER BY total_requests DESC
		LIMIT ?
	`, since, limit).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get sub2api model usage: %w", err)
	}
	return rows, nil
}

func (r *Repository) GetAccountUsage(ctx context.Context, since time.Time) ([]AccountUsageRow, error) {
	db, err := r.database()
	if err != nil {
		return nil, fmt.Errorf("get sub2api account usage: %w", err)
	}

	var rows []AccountUsageRow
	err = db.WithContext(ctx).Raw(`
		SELECT
			account_id,
			COUNT(*) AS total_requests,
			COALESCE(SUM(input_tokens), 0) AS input_tokens,
			COALESCE(SUM(output_tokens), 0) AS output_tokens,
			COALESCE(SUM(cache_creation_tokens), 0) AS cache_creation_tokens,
			COALESCE(SUM(cache_read_tokens), 0) AS cache_read_tokens,
			COALESCE(SUM(total_cost), 0) AS total_cost,
			COALESCE(SUM(actual_cost), 0) AS actual_cost,
			COALESCE(AVG(duration_ms), 0) AS average_duration_ms,
			MAX(created_at) AS last_request_at
		FROM usage_logs
		WHERE created_at >= ?
			AND account_id IS NOT NULL
		GROUP BY account_id
		ORDER BY total_requests DESC
	`, since).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("get sub2api account usage: %w", err)
	}
	return rows, nil
}
