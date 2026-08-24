package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lacsar712/snowmelt/internal/app"
	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/config"
	"github.com/lacsar712/snowmelt/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("CAL-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	holder := "cal-tech"
	app.CalibrateProbe = func(ctx context.Context) error {
		return fmt.Errorf("scale fault")
	}
	defer func() { app.CalibrateProbe = nil }()
	if err := a.Calibrate(context.Background(), holder); err == nil {
		t.Fatal("expected calibrate error")
	}
	now := clk.Now()
	leaseErr := a.Interlock().Leases().Require(cfg.UnitID, holder, now)
	if !errors.Is(leaseErr, model.ErrLeaseMissing) {
		t.Fatalf("lease should be released after failed calibrate, got %v", leaseErr)
	}
}
