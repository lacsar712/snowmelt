package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/snowmelt/internal/model"
)

type PlantStore struct {
	mu    sync.RWMutex
	units map[string]model.PlantSnapshot
	seq   uint64
}

func NewPlantStore() *PlantStore {
	return &PlantStore{units: make(map[string]model.PlantSnapshot)}
}

func (s *PlantStore) Register(unitID string, snap model.PlantSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.units[unitID]; ok {
		return fmt.Errorf("unit %s already registered", unitID)
	}
	snap.UnitID = unitID
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) Get(unitID string) (model.PlantSnapshot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap, ok := s.units[unitID]
	return model.CloneSnapshot(snap), ok
}

func (s *PlantStore) Require(unitID string) (model.PlantSnapshot, error) {
	snap, ok := s.Get(unitID)
	if !ok {
		return model.PlantSnapshot{}, fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	return snap, nil
}

func (s *PlantStore) UpdateState(unitID string, state model.PlantState) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	snap.State = state
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) UpdateGlycol(unitID string, glycol model.GlycolReading) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	snap.Glycol = glycol
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) UpdateHeatbank(unitID string, comb model.HeatbankReading) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	snap.Heatbank = comb
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) UpdateRunway(unitID string, runway model.RunwayReading) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	snap.Runway = runway
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) UpdateSettings(unitID string, settings model.PlantSettings) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	snap.Settings = settings
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) SetAlarms(unitID string, alarms []model.AlarmEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return
	}
	snap.Alarms = append([]model.AlarmEvent(nil), alarms...)
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
}

func (s *PlantStore) CompareAndSwap(unitID string, expectedRevision uint64, update func(*model.PlantSnapshot) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap, ok := s.units[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrPlantNotFound)
	}
	if snap.Revision != expectedRevision {
		return fmt.Errorf("%w", model.ErrSnapshotStale)
	}
	if err := update(&snap); err != nil {
		return err
	}
	snap.UpdatedAt = time.Now()
	s.seq++
	snap.Revision = s.seq
	s.units[unitID] = snap
	return nil
}

func (s *PlantStore) Revision() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.seq
}

func (s *PlantStore) ListUnits() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]string, 0, len(s.units))
	for id := range s.units {
		out = append(out, id)
	}
	return out
}
