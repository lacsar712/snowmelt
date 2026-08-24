package clock

import (
	"context"
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

func (s *Scheduler) Cancel(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if cancel, ok := s.tasks[id]; ok {
		cancel()
		delete(s.tasks, id)
	}
}

func (s *Scheduler) CancelAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, cancel := range s.tasks {
		cancel()
		delete(s.tasks, id)
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
