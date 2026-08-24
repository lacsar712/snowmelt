package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/snowmelt/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Heatbank.PreheatStartedAt.IsZero() {
		return false, "preheat not started"
	}
	if !a.preheatWindow.Ready(snap.Heatbank.PreheatStartedAt) {
		return false, "preheat window open"
	}
	if !snap.Heatbank.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Heatbank.IgnitionAt) {
		return false, "heatbank warmup window open"
	}
	if !snap.Glycol.LastSwellAt.IsZero() {
		if err := a.glycol.RequireSettled(snap.Glycol); err != nil {
			return false, "glycol swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) PreheatRemaining() string {
	snap := a.Snapshot()
	if snap.Heatbank.PreheatStartedAt.IsZero() {
		return "not started"
	}
	if a.preheatWindow.Ready(snap.Heatbank.PreheatStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) HeatbankWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Heatbank.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Heatbank.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
