package app

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/config"
	"github.com/lacsar712/snowmelt/internal/fsm"
	"github.com/lacsar712/snowmelt/internal/model"
)

func Bootstrap(cfgPath string) (*App, error) {
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, err
	}
	if err := config.Validate(cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	clk := clock.NewReal()
	app := New(cfg, clk)
	if err := app.EnsureDataDir(); err != nil {
		return nil, err
	}
	if err := app.journal.Load(); err != nil {
		return nil, fmt.Errorf("journal: %w", err)
	}
	snap := model.DefaultSnapshot(cfg.UnitID)
	snap.Settings = cfg.Settings
	if err := app.store.Register(cfg.UnitID, snap); err != nil {
		return nil, err
	}
	fsm.RegisterLoggingHooks(app.fsm.Hooks())
	fsm.RegisterZonelockDriveHook(app.fsm.Hooks())
	app.journalEvent("bootstrap", storePayload("unit", cfg.UnitID))
	return app, nil
}

func BootstrapWithClock(cfg config.Config, clk clock.ProcessClock) (*App, error) {
	if err := config.Validate(cfg); err != nil {
		return nil, err
	}
	app := New(cfg, clk)
	snap := model.DefaultSnapshot(cfg.UnitID)
	snap.Settings = cfg.Settings
	if err := app.store.Register(cfg.UnitID, snap); err != nil {
		return nil, err
	}
	fsm.RegisterZonelockDriveHook(app.fsm.Hooks())
	return app, nil
}

func (a *App) EnsureDataDir() error {
	if a.cfg.DataDir == "" {
		return nil
	}
	return os.MkdirAll(a.cfg.DataDir, 0o755)
}

func (a *App) DataPath(name string) string {
	return filepath.Join(a.cfg.DataDir, name)
}

func storePayload(k, v string) string {
	return fmt.Sprintf(`{"%s":"%s"}`, k, v)
}
