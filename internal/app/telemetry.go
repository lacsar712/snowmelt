package app

import (
	"sync"
	"time"
)

type TelemetrySnapshot struct {
	UnitID      string    `json:"unit_id"`
	TickCount   uint64    `json:"tick_count"`
	FiringTicks uint64    `json:"firing_ticks"`
	LastTickAt  time.Time `json:"last_tick_at"`
	StartedAt   time.Time `json:"started_at"`
}

type Telemetry struct {
	mu          sync.RWMutex
	unitID      string
	tickCount   uint64
	firingTicks uint64
	coalFeedTPH float64
	lastTickAt  time.Time
	startedAt   time.Time
}

func NewTelemetry(unitID string) *Telemetry {
	return &Telemetry{unitID: unitID, startedAt: time.Now()}
}

func (t *Telemetry) RecordCoalFeed(tph float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.coalFeedTPH = tph
}

func (t *Telemetry) CoalFeedTPH() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.coalFeedTPH
}

func (t *Telemetry) RecordTick(firing bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tickCount++
	if firing {
		t.firingTicks++
	}
	t.lastTickAt = time.Now()
}

func (t *Telemetry) Snapshot() TelemetrySnapshot {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return TelemetrySnapshot{
		UnitID:      t.unitID,
		TickCount:   t.tickCount,
		FiringTicks: t.firingTicks,
		LastTickAt:  t.lastTickAt,
		StartedAt:   t.startedAt,
	}
}

func (t *Telemetry) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.tickCount = 0
	t.firingTicks = 0
	t.startedAt = time.Now()
}

func (t *Telemetry) Uptime() time.Duration {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return time.Since(t.startedAt)
}

func (t *Telemetry) FiringRatio() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.tickCount == 0 {
		return 0
	}
	return float64(t.firingTicks) / float64(t.tickCount)
}
