package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/dyammarcano/gin-nats-starter/internal/model"
	"github.com/nats-io/nats.go"

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

	_, err = cfg.nc.QueueSubscribe("service.identity", "identity-workers", identityWorkers(db))
	if err != nil {
		return err
	}

	log.Println("Identity service listening")
	select {}
}

func identityWorkers(db *gorm.DB) func(m *nats.Msg) {
	return func(m *nats.Msg) {
		log.Printf("[identity] headers=%v data=%s", m.Header, string(m.Data))

		var ident model.Identity

		if _, ok := m.Header["scheme"]; ok {
			b, _ := json.Marshal(ident)
			_ = m.RespondMsg(&nats.Msg{Data: b, Header: nats.Header{"schema": {string(b)}}})
			return
		}

		var req map[string]string
		if err := json.Unmarshal(m.Data, &req); err != nil {
			_ = m.Respond([]byte(`{"error":"bad request"}`))
			return
		}

		docQ := req["document"]
		valid := false
		if len(docQ) == 11 || len(docQ) == 14 {
			valid = true
		}

		_ = db.First(&ident, "cpf = ? OR cnpj = ?", docQ, docQ)
		b, _ := json.Marshal(map[string]any{"valid": valid, "found_name": ident.Name})
		_ = m.Respond(b)
	}
}
