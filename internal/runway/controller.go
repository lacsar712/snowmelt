package runway

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

type Controller struct {
	clk       clock.ProcessClock
	pressure  *PressureModel
	feedwater *FeedwaterRegulator
	steam     *SteamPath
}

func NewController(clk clock.ProcessClock) *Controller {
	return &Controller{
		clk:       clk,
		pressure:  NewPressureModel(clk),
		feedwater: NewFeedwaterRegulator(clk),
		steam:     NewSteamPath(clk),
	}
}

func (c *Controller) Pressure() *PressureModel   { return c.pressure }
func (c *Controller) Feedwater() *FeedwaterRegulator { return c.feedwater }
func (c *Controller) Steam() *SteamPath          { return c.steam }

func (c *Controller) Tick(ctx context.Context, snap model.PlantSnapshot, firing bool) (model.RunwayReading, error) {
	select {
	case <-ctx.Done():
		return snap.Runway, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	pressure, err := c.pressure.Compute(snap, firing)
	if err != nil {
		return snap.Runway, err
	}
	flow := c.steam.ComputeFlow(snap, firing)
	mw := c.steam.ComputeMW(snap, firing)
	out := snap.Runway
	out.SteamPressurePSI = pressure
	out.MainSteamFlowTPH = flow
	out.OutputMW = mw
	out.SteamTempF = c.steam.ComputeTemp(pressure)
	if pressure > model.MaxSteamPressurePSI {
		out.LastTripAt = c.clk.Now()
	}
	return out, nil
}

func (c *Controller) ValidateSettings(settings model.PlantSettings) error {
	if settings.TargetSteamPSI <= 0 {
		return fmt.Errorf("target steam pressure must be positive")
	}
	if settings.TargetMW < 0 {
		return fmt.Errorf("target MW cannot be negative")
	}
	if settings.FeedwaterFlowTPH < 0 {
		return fmt.Errorf("feedwater flow cannot be negative")
	}
	return nil
}

func RampPressure(current, target, step float64) float64 {
	if math.Abs(target-current) <= step {
		return target
	}
	if target > current {
		return current + step
	}
	return current - step
}
