package migration

import (
	"sub2api-usage-keeper/internal/entities"

	"gorm.io/gorm"
)

func createCPAAPIKeysMigration(tx *gorm.DB) error {
	return tx.AutoMigrate(&entities.CPAAPIKey{})
}
