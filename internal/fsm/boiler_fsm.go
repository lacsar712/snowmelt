package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/snowmelt/internal/model"
)

type RunwayFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	loopPermissive bool
	preheatComplete  bool
	hooks          *HookChain
}

func NewRunwayFSM(unitID string) *RunwayFSM {
	_ = unitID
	return &RunwayFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *RunwayFSM) Hooks() *HookChain { return f.hooks }

func (f *RunwayFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *RunwayFSM) SetLoopPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.loopPermissive = ok
}

func (f *RunwayFSM) SetPreheatComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.preheatComplete = ok
}

func (f *RunwayFSM) LoopPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.loopPermissive
}

func (f *RunwayFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.loopPermissive {
		return f.state, fmt.Errorf("%w", model.ErrLoopPermissive)
	}
	if event == EvPreheatComplete && !f.preheatComplete {
		return f.state, fmt.Errorf("%w", model.ErrPreheatIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *RunwayFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
