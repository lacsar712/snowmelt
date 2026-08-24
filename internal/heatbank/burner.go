package heatbank

import (
	"math"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateZonelockTemp(reading model.HeatbankReading) float64 {
	base := 300.0
	loopHeat := reading.LoopFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + loopHeat - airCool
}

func (b *BurnerController) MeltStable(reading model.HeatbankReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.ZonelockTempF > 800 && reading.ExcessO2Pct >= model.MinZonelockO2Percent
}

func (b *BurnerController) TripRequired(reading model.HeatbankReading) bool {
	if reading.ExcessO2Pct > model.MaxZonelockO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.ZonelockTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerPreheat:
		return "Preheat"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Melt"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.HeatbankReading) float64 {
	return reading.LoopFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentLoop float64) float64 {
	if settings.LoopFlowTPH <= 0 {
		return 0
	}
	return currentLoop / settings.LoopFlowTPH
}

func (b *BurnerController) MinStableLoop(settings model.PlantSettings) float64 {
	return settings.LoopFlowTPH * 0.25
}

func (b *BurnerController) NormalizeLoop(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
