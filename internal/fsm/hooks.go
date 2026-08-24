package fsm

import (
	"context"

	"github.com/lacsar712/snowmelt/internal/model"
)

type TransitionHook func(ctx context.Context, from, to model.PlantState, event PlantEvent) error

type HookChain struct {
	before []TransitionHook
	after  []TransitionHook
}

func NewHookChain() *HookChain { return &HookChain{} }

func (h *HookChain) OnBefore(fn TransitionHook) { h.before = append(h.before, fn) }
func (h *HookChain) OnAfter(fn TransitionHook)  { h.after = append(h.after, fn) }

func (h *HookChain) RunBefore(ctx context.Context, from, to model.PlantState, event PlantEvent) error {
	for _, fn := range h.before {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

func (h *HookChain) RunAfter(ctx context.Context, from, to model.PlantState, event PlantEvent) error {
	for _, fn := range h.after {
		if err := fn(ctx, from, to, event); err != nil {
			return err
		}
	}
	return nil
}

func (h *HookChain) Count() int { return len(h.before) + len(h.after) }

type LoggingHook struct{}

func (LoggingHook) Before(ctx context.Context, from, to model.PlantState, event PlantEvent) error {
	_ = ctx
	_ = from
	_ = to
	_ = event
	return nil
}

func (LoggingHook) After(ctx context.Context, from, to model.PlantState, event PlantEvent) error {
	_ = ctx
	_ = from
	_ = to
	_ = event
	return nil
}

// ZonelockDrivePulse counts zonelock-side effects from FSM after hooks (acceptance).
var ZonelockDrivePulse func()

func RegisterZonelockDriveHook(chain *HookChain) {
	chain.OnAfter(func(ctx context.Context, from, to model.PlantState, event PlantEvent) error {
		_ = ctx
		_ = from
		_ = to
		_ = event
		if ZonelockDrivePulse != nil {
			ZonelockDrivePulse()
		}
		return nil
	})
}

func RegisterLoggingHooks(chain *HookChain) {
	h := LoggingHook{}
	chain.OnBefore(h.Before)
	chain.OnAfter(h.After)
}
