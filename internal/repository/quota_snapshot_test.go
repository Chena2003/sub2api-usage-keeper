package repository

import (
	"testing"
	"time"

	"sub2api-usage-keeper/internal/entities"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpsertQuotaSnapshotsUpdatesExistingAccountSnapshot(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite database: %v", err)
	}
	if err := db.AutoMigrate(&entities.QuotaSnapshot{}); err != nil {
		t.Fatalf("auto migrate quota snapshots: %v", err)
	}

	checkedAt := time.Date(2026, 5, 17, 10, 0, 0, 0, time.UTC)
	if err := UpsertQuotaSnapshots(db, []entities.QuotaSnapshot{{
		AccountID:   4,
		Provider:    "openai",
		AccountType: "oauth",
		Status:      "active",
		PlanType:    "pro",
		CheckedAt:   checkedAt,
	}}); err != nil {
		t.Fatalf("first upsert returned error: %v", err)
	}
	if err := UpsertQuotaSnapshots(db, []entities.QuotaSnapshot{{
		AccountID: 4,
		Status:    "error",
		CheckedAt: checkedAt.Add(time.Minute),
	}}); err != nil {
		t.Fatalf("second upsert returned error: %v", err)
	}

	var count int64
	if err := db.Model(&entities.QuotaSnapshot{}).Count(&count).Error; err != nil {
		t.Fatalf("count quota snapshots: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one quota snapshot row, got %d", count)
	}

	var row entities.QuotaSnapshot
	if err := db.Where("account_id = ?", 4).First(&row).Error; err != nil {
		t.Fatalf("load quota snapshot: %v", err)
	}
	if row.Status != "error" {
		t.Fatalf("expected status error after upsert, got %q", row.Status)
	}
}
