package heatbank

import (
	"math"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

type LoopRegulator struct {
	clk clock.ProcessClock
}

func NewLoopRegulator(clk clock.ProcessClock) *LoopRegulator {
	return &LoopRegulator{clk: clk}
}

func (f *LoopRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.LoopFlowTPH * 0.08
}

func (f *LoopRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.LoopFlowTPH * loadPct
}

func (f *LoopRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *LoopRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *LoopRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *LoopRegulator) ValidatePermissive(settings model.PlantSettings, glycolOK, preheatOK bool) error {
	if !preheatOK {
		return model.ErrPreheatIncomplete
	}
	if !glycolOK {
		return model.ErrGlycolLevelTrip
	}
	if settings.LoopFlowTPH <= 0 {
		return model.ErrLoopPermissive
	}
	return nil
}

func (f *LoopRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.LoopFlowTPH * 0.2
}

func (f *LoopRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.LoopFlowTPH * 1.1
}
