package repository

import (
	"fmt"

	"sub2api-usage-keeper/internal/entities"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func UpsertQuotaSnapshots(db *gorm.DB, snapshots []entities.QuotaSnapshot) error {
	if db == nil {
		return fmt.Errorf("database is nil")
	}
	if len(snapshots) == 0 {
		return nil
	}

	if err := db.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "account_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"provider",
			"account_type",
			"display_name",
			"plan_type",
			"status",
			"schedulable",
			"session_window_status",
			"reset_at",
			"expires_at",
			"last_used_at",
			"raw_public_json",
			"checked_at",
			"updated_at",
		}),
	}).Create(&snapshots).Error; err != nil {
		return fmt.Errorf("upsert quota snapshots: %w", err)
	}

	return nil
}
