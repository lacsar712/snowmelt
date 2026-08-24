package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

func (a *App) bindLoopLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if cancel, ok := a.loopLoopCancels[holder]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.loopLoopCancels[holder] = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelLoopLoop(holder string) {
	a.mu.Lock()
	if cancel, ok := a.loopLoopCancels[holder]; ok {
		cancel()
		delete(a.loopLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) cancelAllLoopLoops() {
	a.mu.Lock()
	for holder, cancel := range a.loopLoopCancels {
		cancel()
		delete(a.loopLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Heatbank.LoopFlowTPH
}

func (a *App) RunLoopRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindLoopLoop(holder, ctx)
	defer a.cancelLoopLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Heatbank.LoopFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Heatbank
		comb.LoopFlowTPH = current + 1.0
		_ = a.store.UpdateHeatbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.LoopFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindLoopLoop(holder, ctx)
	defer a.cancelLoopLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Heatbank
		comb.LoopFlowTPH += 0.5
		_ = a.store.UpdateHeatbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.LoopFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
