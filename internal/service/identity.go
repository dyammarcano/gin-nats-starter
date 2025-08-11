package service

import (
	"fmt"

	"awesomeProject6/internal/model"

	"github.com/spf13/cobra"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newDatabase(cfg *ConfigService) (*gorm.DB, error) {
	if cfg.Database.DBPath == "" {
		return nil, fmt.Errorf("database path required")
	}

	db, err := gorm.Open(sqlite.Open(cfg.Database.DBPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

func Identity(cmd *cobra.Command, args []string) error {
	configStr := cmd.Flag("config").Value.String()
	cfg, err := serviceCommon(cmd.Context(), configStr)
	if err != nil {
		return err
	}

	db, err := newDatabase(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := db.AutoMigrate(
		&model.Identity{},
		&model.CEP{},
	); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}
