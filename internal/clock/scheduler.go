package clock

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/lacsar712/snowmelt/internal/model"
)

type ScheduledFunc func(ctx context.Context) error

type Scheduler struct {
	clk       ProcessClock
	mu        sync.Mutex
	tasks     map[string]context.CancelFunc
	planItems map[string]bool
	running   bool
}

func NewScheduler(clk ProcessClock) *Scheduler {
	return &Scheduler{clk: clk, tasks: make(map[string]context.CancelFunc), planItems: make(map[string]bool)}
}

func (s *Scheduler) Clock() ProcessClock { return s.clk }

func (s *Scheduler) Schedule(parent context.Context, id string, interval time.Duration, fn ScheduledFunc) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.tasks[id]; ok {
		cancel()
	}
	ctx, cancel := context.WithCancel(parent)
	s.tasks[id] = cancel
	go s.runLoop(ctx, id, interval, fn)
	return nil
}

func (s *Scheduler) runLoop(ctx context.Context, id string, interval time.Duration, fn ScheduledFunc) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := fn(ctx); err != nil {
				_ = id
			}
		}
	}
}

// Cancel tears down a single running task. It does not touch plan items,
// because plan items are keyed by plan id, not by task id.
func (s *Scheduler) Cancel(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.tasks[id]; ok {
		cancel()
		delete(s.tasks, id)
	}
}

// CancelBurnPlan withdraws every heat-sequence step belonging to planID from the
// backend queue. A withdraw must clear the queue so the planner no longer sees
// the post-preheat heat segments; otherwise the stale steps linger and get
// re-appended on the next install.
func (s *Scheduler) CancelBurnPlan(planID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prefix := planID + ":"
	for key := range s.planItems {
		if strings.HasPrefix(key, prefix) {
			delete(s.planItems, key)
		}
	}
}

func (s *Scheduler) CancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, cancel := range s.tasks {
		cancel()
		delete(s.tasks, id)
	}
	// A full teardown must also drain the backend heat-sequence queue; leaving
	// planItems behind means a withdraw only changes the screen, not the queue.
	for key := range s.planItems {
		delete(s.planItems, key)
	}
}

func (s *Scheduler) After(ctx context.Context, d time.Duration, fn func() error) {
	go func() {
		deadline := s.clk.Now().Add(d)
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if s.clk.Now().After(deadline) || s.clk.Now().Equal(deadline) {
					_ = fn()
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()
}

func (s *Scheduler) ActiveCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.tasks)
}

type BurnScheduleEntry struct {
	Name     string
	StartsAt time.Time
}

func PlanBurnSchedule(clk ProcessClock, settings model.PlantSettings, planID string) []BurnScheduleEntry {
	_ = settings
	now := clk.Now()
	return []BurnScheduleEntry{
		{Name: "ignite-prep", StartsAt: now},
		{Name: "ignite-main", StartsAt: now.Add(5 * time.Second)},
		{Name: "ignite-confirm", StartsAt: now.Add(10 * time.Second)},
	}
}

func (s *Scheduler) InstallBurnPlan(settings model.PlantSettings, planID string) error {
	return s.InstallBurnPlanCtx(context.Background(), settings, planID)
}

func (s *Scheduler) InstallBurnPlanCtx(ctx context.Context, settings model.PlantSettings, planID string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, e := range PlanBurnSchedule(s.clk, settings, planID) {
		if err := ctx.Err(); err != nil {
			return err
		}
		s.mu.Lock()
		s.planItems[planID+":"+e.Name] = true
		s.mu.Unlock()
		time.Sleep(time.Millisecond)
	}
	return nil
}

func (s *Scheduler) ItemCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.planItems)
}
