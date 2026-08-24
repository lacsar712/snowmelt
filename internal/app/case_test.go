package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/snowmelt/internal/app"
	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/config"
	"github.com/lacsar712/snowmelt/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("FLAME-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	comb := a.Snapshot().Heatbank
	comb.BurnerPhase = model.BurnerStable
	comb.ZonelockTempF = 400
	if err := a.Store().UpdateHeatbank(cfg.UnitID, comb); err != nil {
		t.Fatal(err)
	}
	err = a.OnMeltLoss(context.Background(), "maint-op")
	if err == nil {
		t.Fatal("expected melt loss error")
	}
	if !errors.Is(err, model.ErrMeltLoss) {
		t.Fatalf("expected ErrMeltLoss, got %v", err)
	}
}
