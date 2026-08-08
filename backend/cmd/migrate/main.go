// Command migrate applies GORM migrations (AutoMigrate) to the database.
package main

import (
	"fmt"
	"os"

	"github.com/aeroxe/sign-flow/backend/internal/config"
	"github.com/aeroxe/sign-flow/backend/internal/database"
	"github.com/aeroxe/sign-flow/backend/internal/pkg/logger"
)

func main() {
	cfg := config.Load()
	logger.Init(cfg.Env, cfg.LogLevel)
	db, err := database.Init(cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database connect failed: %v\n", err)
		os.Exit(1)
	}
	if err := database.Migrate(db, database.AllModels()...); err != nil {
		fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("migrations applied successfully")
}
