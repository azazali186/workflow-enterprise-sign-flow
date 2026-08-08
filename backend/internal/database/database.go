// Package database bootstraps the GORM connection and runs migrations.
// All persistence goes through GORM — no raw SQL is used anywhere.
package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/cryptoser"
)

// DB is the shared GORM connection.
var DB *gorm.DB

// Init opens the PostgreSQL connection and registers the encryption serializer.
func Init(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(50)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	schema.RegisterSerializer(cryptoser.Name, cryptoser.Enc{})
	DB = db
	return db, nil
}

// Migrate runs AutoMigrate for every model. Safe to call on every start.
func Migrate(db *gorm.DB, models ...any) error {
	return db.AutoMigrate(models...)
}

// Transaction runs fn inside a DB transaction with retry on serialisation errors.
func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}
