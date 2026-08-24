package heatbank

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	loop    *LoopRegulator
	preheat   *clock.PreheatWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.HeatbankWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		loop:     NewLoopRegulator(clk),
		preheat:    clock.NewPreheatWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewHeatbankWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Loop() *LoopRegulator     { return c.loop }

func (c *Coordinator) StartPreheat(ctx context.Context, snap model.PlantSnapshot) (model.HeatbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Heatbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Heatbank
	out.BurnerPhase = model.BurnerPreheat
	out.PreheatStartedAt = c.clk.Now()
	out.LoopFlowTPH = 0
	out.AirflowTPH = c.airflow.PreheatRate()
	return out, nil
}

func (c *Coordinator) CompletePreheat(snap model.HeatbankReading) error {
	return c.preheat.Require(snap.PreheatStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.HeatbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Heatbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.preheat.Require(snap.Heatbank.PreheatStartedAt); err != nil {
		return snap.Heatbank, err
	}
	out := snap.Heatbank
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.LoopFlowTPH = c.loop.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.ZonelockTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.HeatbankReading, error) {
	if err := c.ignition.Require(snap.Heatbank.IgnitionAt); err != nil {
		return snap.Heatbank, err
	}
	out := snap.Heatbank
	out.BurnerPhase = model.BurnerStable
	out.LoopFlowTPH = snap.Settings.LoopFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.ZonelockTempF = c.burner.EstimateZonelockTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.HeatbankReading {
	out := snap.Heatbank
	out.LoopFlowTPH = snap.Settings.LoopFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.ZonelockTempF = c.burner.EstimateZonelockTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.HeatbankReading) model.HeatbankReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.LoopFlowTPH = 0
	out.ZonelockTempF = math.Max(200, out.ZonelockTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.HeatbankReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
