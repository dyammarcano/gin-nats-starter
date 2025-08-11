package service

import (
	"context"
	"fmt"
	"time"

	"github.com/inovacc/config"
	"github.com/nats-io/nats.go"
)

type ConfigService struct {
	nc         *nats.Conn
	ctx        context.Context
	cancel     context.CancelFunc
	BaseConfig *config.Config `yaml:"-"`
	Port       int            `yaml:"port"`
	Nats       NatsConfig     `yaml:"nats"`
	Database   Database       `yaml:"database"`
}

func (c *ConfigService) Close() error {
	c.cancel()
	c.nc.Close()
	return nil
}

type Database struct {
	DBPath string `yaml:"db_path"`
}

type NatsConfig struct {
	Url           string        `yaml:"url"`
	Name          string        `yaml:"name"`
	ReconnectWait time.Duration `yaml:"reconnectWait"`
	MaxReconnects int           `yaml:"maxReconnects"`
}

func serviceCommon(ctx context.Context, configPath string) (*ConfigService, error) {
	if err := config.InitServiceConfig(&ConfigService{}, configPath); err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	cfg, err := config.GetServiceConfig[*ConfigService]()
	if err != nil {
		return nil, fmt.Errorf("failed to get service config: %v", err)
	}

	cfg.BaseConfig = config.GetBaseConfig()

	if cfg.Nats.Url == "" {
		return nil, fmt.Errorf("nats url is required")
	}

	cfg.nc, err = nats.Connect(
		cfg.Nats.Url,
		nats.Name(cfg.Nats.Name),
		nats.MaxReconnects(cfg.Nats.MaxReconnects),
		nats.ReconnectWait(cfg.Nats.ReconnectWait*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cfg.ctx = ctx
	cfg.cancel = cancel

	return cfg, nil
}
