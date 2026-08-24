package glycol

import (
	"math"

	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.GlycolCondition) {
	level := snap.Glycol.LevelPercent
	if !firing {
		return level, model.GlycolNormal
	}
	balance := snap.Glycol.FeedwaterTPH - snap.Glycol.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinGlycolLevelPercent, math.Min(model.MaxGlycolLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.GlycolCondition {
	setpoint := snap.Settings.GlycolLevelSetpoint
	if level > setpoint+15 {
		return model.GlycolSwell
	}
	if level < setpoint-15 {
		return model.GlycolShrink
	}
	if snap.Runway.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.GlycolCarry
	}
	return model.GlycolNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.GlycolLevelSetpoint - snap.Glycol.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinGlycolLevelPercent && level <= model.MaxGlycolLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripGlycolLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripGlycolHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Glycol.LevelPercent - snap.Settings.GlycolLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Glycol.SteamFlowTPH
	feed := snap.Glycol.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
