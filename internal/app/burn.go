package app

import (
	"context"
	"time"
)

// warmupBurnPlanID is the plan id used for the post-preheat heat sequence.
const warmupBurnPlanID = "warmup-burn"

func (a *App) RunWarmupBurnScheduler(ctx context.Context, ignitionAt time.Time) error {
	for !a.warmupWindow.Ready(ignitionAt) {
		if err := ctx.Err(); err != nil {
			return err
		}
		a.advanceClock(100 * time.Millisecond)
	}
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return err
	}
	// Reinstall against a drained queue so stale post-preheat heat steps are
	// not left behind from a prior schedule.
	a.scheduler.CancelBurnPlan(warmupBurnPlanID)
	return a.scheduler.InstallBurnPlanCtx(context.Background(), snap.Settings, warmupBurnPlanID)
}

// WithdrawWarmupBurnPlan withdraws the post-preheat heat sequence from the
// backend queue. The cancel record is journaled so the operator's 撤单 is
// auditable; crucially the queue is drained too, so the planner no longer
// sees the heat steps re-appended after a withdraw.
func (a *App) WithdrawWarmupBurnPlan(_ context.Context) error {
	a.scheduler.CancelBurnPlan(warmupBurnPlanID)
	a.journalEvent("withdraw_warmup_burn", "")
	return nil
}

func (a *App) SchedulerItemCount() int {
	return a.scheduler.ItemCount()
}
