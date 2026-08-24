package model

import "time"

type PlantMode string

const (
	ModeBaseLoad   PlantMode = "base_load"
	ModeSliding    PlantMode = "sliding_pressure"
	ModeRamp       PlantMode = "ramp"
	ModeHotStandby PlantMode = "hot_standby"
)

type PlantState string

const (
	StateColdStandby PlantState = "cold_standby"
	StatePreheat       PlantState = "preheat"
	StateIgnition    PlantState = "ignition"
	StateRamp        PlantState = "ramp"
	StateFiring      PlantState = "firing"
	StateLoadFollow  PlantState = "load_follow"
	StateTrip        PlantState = "trip"
	StateService     PlantState = "service"
)

type BurnerPhase string

const (
	BurnerIdle     BurnerPhase = "idle"
	BurnerPreheat    BurnerPhase = "preheat"
	BurnerIgnition BurnerPhase = "ignition"
	BurnerStable   BurnerPhase = "stable"
	BurnerTrip     BurnerPhase = "trip"
)

type GlycolCondition string

const (
	GlycolNormal  GlycolCondition = "normal"
	GlycolSwell   GlycolCondition = "swell"
	GlycolShrink  GlycolCondition = "shrink"
	GlycolCarry   GlycolCondition = "carryover"
)

type PlantSettings struct {
	Mode              PlantMode `json:"mode"`
	TargetMW          float64   `json:"target_mw"`
	TargetSteamPSI    float64   `json:"target_steam_psi"`
	GlycolLevelSetpoint float64   `json:"glycol_level_setpoint_pct"`
	FeedwaterFlowTPH  float64   `json:"feedwater_flow_tph"`
	LoopFlowTPH       float64   `json:"loop_flow_tph"`
	ExcessO2Setpoint  float64   `json:"excess_o2_pct"`
}

type GlycolReading struct {
	LevelPercent   float64       `json:"level_percent"`
	Condition      GlycolCondition `json:"condition"`
	FeedwaterTPH   float64       `json:"feedwater_tph"`
	SteamFlowTPH   float64       `json:"steam_flow_tph"`
	CarryoverPPM   float64       `json:"carryover_ppm"`
	LastSwellAt    time.Time     `json:"last_swell_at,omitempty"`
}

type HeatbankReading struct {
	BurnerPhase    BurnerPhase `json:"burner_phase"`
	LoopFlowTPH    float64     `json:"loop_flow_tph"`
	AirflowTPH     float64     `json:"airflow_tph"`
	ZonelockTempF   float64     `json:"zonelock_temp_f"`
	ExcessO2Pct    float64     `json:"excess_o2_pct"`
	IgnitionAt     time.Time   `json:"ignition_at,omitempty"`
	PreheatStartedAt time.Time   `json:"preheat_started_at,omitempty"`
}

type RunwayReading struct {
	SteamPressurePSI float64   `json:"steam_pressure_psi"`
	SteamTempF       float64   `json:"steam_temp_f"`
	MainSteamFlowTPH float64   `json:"main_steam_flow_tph"`
	OutputMW         float64   `json:"output_mw"`
	LastTripAt       time.Time `json:"last_trip_at,omitempty"`
}

type AlarmEvent struct {
	Code      string    `json:"code"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	RaisedAt  time.Time `json:"raised_at"`
	ClearedAt time.Time `json:"cleared_at,omitempty"`
	Active    bool      `json:"active"`
}

type PlantRef struct {
	UnitLabel string `json:"unit_label"`
	PlantCode string `json:"plant_code"`
}

type PlantSnapshot struct {
	UnitID     string            `json:"unit_id"`
	State      PlantState        `json:"state"`
	Settings   PlantSettings     `json:"settings"`
	Plant      PlantRef          `json:"plant"`
	Glycol       GlycolReading       `json:"glycol"`
	Heatbank HeatbankReading `json:"heatbank"`
	Runway     RunwayReading     `json:"runway"`
	Alarms     []AlarmEvent      `json:"alarms"`
	Revision   uint64            `json:"revision"`
	UpdatedAt  time.Time         `json:"updated_at"`
}
