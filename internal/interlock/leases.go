package interlock

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lacsar712/snowmelt/internal/model"
)

type Lease struct {
	unitID, holder string
	registry       *LeaseRegistry
	released       bool
}

func (l *Lease) Release() {
	if l.registry == nil || l.released {
		return
	}
	l.registry.release(l.unitID, l.holder)
	l.released = true
}

type LeaseRegistry struct {
	mu      sync.Mutex
	leases  map[string]string
	expires map[string]time.Time
	ttl     time.Duration
}

func NewLeaseRegistry(ttl time.Duration) *LeaseRegistry {
	if ttl <= 0 {
		ttl = model.DefaultLeaseTTL
	}
	return &LeaseRegistry{
		leases:  make(map[string]string),
		expires: make(map[string]time.Time),
		ttl:     ttl,
	}
}

func (r *LeaseRegistry) Acquire(unitID, holder string, now time.Time) (*Lease, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.leases[unitID]; ok {
		if exp, ok := r.expires[unitID]; ok && now.Before(exp) {
			return nil, fmt.Errorf("%s held by %s: %w", unitID, h, model.ErrLeaseHeld)
		}
	}
	r.leases[unitID] = holder
	r.expires[unitID] = now.Add(r.ttl)
	return &Lease{unitID: unitID, holder: holder, registry: r}, nil
}

func (r *LeaseRegistry) release(unitID, holder string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.leases[unitID]; !ok || h != holder {
		return
	}
	delete(r.leases, unitID)
	delete(r.expires, unitID)
}

func (r *LeaseRegistry) Require(unitID, holder string, now time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	h, ok := r.leases[unitID]
	if !ok {
		return fmt.Errorf("%w", model.ErrLeaseMissing)
	}
	if h != holder {
		return fmt.Errorf("%w", model.ErrLeaseHeld)
	}
	if exp, ok := r.expires[unitID]; ok && !now.Before(exp) {
		delete(r.leases, unitID)
		return fmt.Errorf("%w", model.ErrLeaseMissing)
	}
	return nil
}

func (r *LeaseRegistry) ReleaseHolder(unitID, holder string) { r.release(unitID, holder) }

func (r *LeaseRegistry) Renew(unitID, holder string, now time.Time) error {
	if err := r.Require(unitID, holder, now); err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.expires[unitID] = now.Add(r.ttl)
	return nil
}

func (r *LeaseRegistry) WithLease(ctx context.Context, unitID, holder string, now time.Time, fn func(context.Context) error) error {
	lease, err := r.Acquire(unitID, holder, now)
	if err != nil {
		return err
	}
	defer lease.Release()
	return fn(ctx)
}

type GateReason string

const (
	ReasonTrip        GateReason = "plant_trip"
	ReasonService     GateReason = "service_mode"
	ReasonGlycolTrip    GateReason = "glycol_level_trip"
	ReasonPressure    GateReason = "pressure_trip"
	ReasonHeatbank  GateReason = "heatbank_trip"
	ReasonPreheat       GateReason = "preheat_incomplete"
)

type SafetyGate struct {
	mu      sync.RWMutex
	blocked map[string]GateReason
}

func NewSafetyGate() *SafetyGate { return &SafetyGate{blocked: make(map[string]GateReason)} }

func (g *SafetyGate) Block(id string, r GateReason) {
	g.mu.Lock()
	g.blocked[id] = r
	g.mu.Unlock()
}

func (g *SafetyGate) Unblock(id string) {
	g.mu.Lock()
	delete(g.blocked, id)
	g.mu.Unlock()
}

func (g *SafetyGate) Allow(id string, state model.PlantState) error {
	g.mu.RLock()
	_, blocked := g.blocked[id]
	g.mu.RUnlock()
	if blocked {
		return fmt.Errorf("%w", model.ErrGateBlocked)
	}
	if state == model.StateTrip || state == model.StateService {
		return fmt.Errorf("%w", model.ErrGateBlocked)
	}
	return nil
}

func (g *SafetyGate) Reason(id string) (GateReason, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	r, ok := g.blocked[id]
	return r, ok
}

type Interlock struct {
	leases *LeaseRegistry
	gate   *SafetyGate
}

func NewInterlock(ttl time.Duration) *Interlock {
	return &Interlock{leases: NewLeaseRegistry(ttl), gate: NewSafetyGate()}
}

func (i *Interlock) Leases() *LeaseRegistry { return i.leases }
func (i *Interlock) Gate() *SafetyGate      { return i.gate }

func (i *Interlock) AuthorizeFiring(unitID, holder string, state model.PlantState, now time.Time) error {
	if err := i.leases.Require(unitID, holder, now); err != nil {
		return err
	}
	return i.gate.Allow(unitID, state)
}
