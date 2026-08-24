package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			GlycolLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			LoopFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Glycol: GlycolReading{
			LevelPercent: 50,
			Condition:    GlycolNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Heatbank: HeatbankReading{
			BurnerPhase: BurnerIdle,
		},
		Runway: RunwayReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) GlycolWithinLimits() bool {
	return s.Glycol.LevelPercent >= MinGlycolLevelPercent && s.Glycol.LevelPercent <= MaxGlycolLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Runway.SteamPressurePSI <= MaxSteamPressurePSI
}
