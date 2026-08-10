// Package database bootstraps the GORM connection and runs migrations.
// All persistence goes through GORM — no raw SQL is used anywhere.
package database

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/aeroxe/sign-flow/backend/internal/pkg/cryptoser"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
)

// DB is the shared GORM connection.
var DB *gorm.DB

// Init opens the PostgreSQL connection and registers the encryption serializer.
// Slow queries (>= 200ms) are logged at Warn so production latency regressions
// surface in the application log.
func Init(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.New(zapWriter{}, gormLogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormLogger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true, // never log raw values (no PII in logs)
		}),
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
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // retire idle conns before the DB drops them
	schema.RegisterSerializer(cryptoser.Name, cryptoser.Enc{})
	DB = db
	return db, nil
}

// Migrate runs AutoMigrate for every model. Safe to call on every start.
func Migrate(db *gorm.DB, models ...any) error {
	return db.AutoMigrate(models...)
}

// zapWriter routes GORM's log output into the application's structured zap
// logger (Warn level), so slow queries appear in the same pipeline as every
// other log line.
type zapWriter struct{}

func (zapWriter) Printf(format string, args ...any) {
	logger.L().Sugar().Warnf(format, args...)
}

// Transaction runs fn inside a DB transaction with retry on serialisation errors.
func Transaction(fn func(tx *gorm.DB) error) error {
	return DB.Transaction(fn)
}
