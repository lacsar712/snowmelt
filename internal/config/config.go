package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lacsar712/snowmelt/internal/model"
)

type Config struct {
	UnitID          string             `json:"unit_id"`
	ListenAddr      string             `json:"listen_addr"`
	DataDir         string             `json:"data_dir"`
	JournalPath     string             `json:"journal_path"`
	JournalCapacity int                `json:"journal_capacity"`
	LeaseTTL        time.Duration      `json:"lease_ttl"`
	Settings        model.PlantSettings `json:"settings"`
	TickInterval    time.Duration      `json:"tick_interval"`
}

func Default(unitID string) Config {
	return Config{
		UnitID:          unitID,
		ListenAddr:      ":8080",
		DataDir:         "./data",
		JournalCapacity: model.DefaultJournalCapacity,
		LeaseTTL:        model.DefaultLeaseTTL,
		Settings:        model.DefaultSnapshot(unitID).Settings,
		TickInterval:    time.Second,
	}
}

func Load(path string) (Config, error) {
	cfg := Default("UNIT-1")
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, fmt.Errorf("read config: %w", err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse config: %w", err)
	}
	if cfg.UnitID == "" {
		cfg.UnitID = "UNIT-1"
	}
	if cfg.ListenAddr == "" {
		cfg.ListenAddr = ":8080"
	}
	if cfg.JournalPath == "" && cfg.DataDir != "" {
		cfg.JournalPath = cfg.DataDir + "/journal.jsonl"
	}
	return cfg, nil
}

func (c Config) Save(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
