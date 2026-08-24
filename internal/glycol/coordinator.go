package glycol

import (
	"context"
	"fmt"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

type Coordinator struct {
	clk       clock.ProcessClock
	level     *LevelController
	carryover *CarryoverMonitor
	swell     *SwellDetector
	settle    *clock.GlycolSwellWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:       clk,
		level:     NewLevelController(clk),
		carryover: NewCarryoverMonitor(clk),
		swell:     NewSwellDetector(clk),
		settle:    clock.NewGlycolSwellWindow(clk),
	}
}

func (c *Coordinator) Level() *LevelController       { return c.level }
func (c *Coordinator) Carryover() *CarryoverMonitor  { return c.carryover }
func (c *Coordinator) Swell() *SwellDetector         { return c.swell }

func (c *Coordinator) Tick(ctx context.Context, snap model.PlantSnapshot, firing bool) (model.GlycolReading, error) {
	select {
	case <-ctx.Done():
		return snap.Glycol, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Glycol
	level, cond := c.level.Compute(snap, firing)
	out.LevelPercent = level
	out.Condition = cond
	out.FeedwaterTPH = snap.Settings.FeedwaterFlowTPH
	if firing {
		out.SteamFlowTPH = snap.Runway.MainSteamFlowTPH
	}
	out.CarryoverPPM = c.carryover.Estimate(out, snap.Runway.SteamPressurePSI)
	if cond == model.GlycolSwell {
		out.LastSwellAt = c.clk.Now()
	}
	return out, nil
}

func (c *Coordinator) SettledAfterSwell(snap model.GlycolReading) bool {
	if snap.LastSwellAt.IsZero() {
		return true
	}
	return c.settle.Settled(snap.LastSwellAt)
}

func (c *Coordinator) RequireSettled(snap model.GlycolReading) error {
	if snap.LastSwellAt.IsZero() {
		return nil
	}
	return c.settle.RequireSettled(snap.LastSwellAt)
}

func (c *Coordinator) TripRequired(snap model.GlycolReading) bool {
	return snap.LevelPercent < model.TripGlycolLowPercent || snap.LevelPercent > model.TripGlycolHighPercent
}

func (c *Coordinator) CoordinateFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	return c.level.RecommendFeedwater(snap, firing)
}
