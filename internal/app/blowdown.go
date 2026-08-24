package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/snowmelt/internal/model"
)

const maxBleedOpeningPct = 100.0

func (a *App) OpenBleed(ctx context.Context, holder string, openingPct float64) error {
	_ = holder
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if openingPct >= maxBleedOpeningPct {
		return fmt.Errorf("bleed: %w", model.ErrBleedLimit)
	}
	return nil
}

func (a *App) BleedAfterShutdown(ctx context.Context, openingPct float64) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	snap := a.Snapshot()
	if snap.State != model.StateTrip && snap.State != model.StateColdStandby {
		return fmt.Errorf("plant not shut down")
	}
	if openingPct >= maxBleedOpeningPct {
		return fmt.Errorf("unknown fault")
	}
	return nil
}
