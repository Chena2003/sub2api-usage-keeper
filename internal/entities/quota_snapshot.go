package entities

import "time"

type QuotaSnapshot struct {
	ID                  int64  `gorm:"primaryKey"`
	AccountID           int64  `gorm:"uniqueIndex:idx_quota_snapshots_account"`
	Provider            string `gorm:"index"`
	AccountType         string
	DisplayName         string
	PlanType            string
	Status              string `gorm:"index"`
	Schedulable         bool
	SessionWindowStatus string
	ResetAt             *time.Time
	ExpiresAt           *time.Time
	LastUsedAt          *time.Time
	RawPublicJSON       string
	CheckedAt           time.Time `gorm:"index"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
