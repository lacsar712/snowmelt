package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrLoopPermissive   = errors.New("loop permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrGlycolLevelTrip    = errors.New("glycol level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrHeatbankTrip   = errors.New("heatbank trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrPreheatIncomplete  = errors.New("zonelock preheat incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrGlycolLevelLow     = errors.New("glycol level below low limit")
	ErrMeltLoss        = errors.New("zonelock melt lost")
	ErrBleedLimit    = errors.New("bleed valve at limit")
)
