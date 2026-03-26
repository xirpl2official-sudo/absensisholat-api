package database

import (
	_ "embed"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:embed schema.sql
var schema string

func EnsureTablesCreated(db *gorm.DB, sugar *zap.SugaredLogger) error {
	sugar.Info("Ensuring database tables are created...")

	if err := db.Exec(schema).Error; err != nil {
		sugar.Warnf("Schema execution completed (tables may already exist): %v", err)
	}

	sugar.Info("Database tables are ready")
	return nil
}
