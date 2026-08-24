package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/snowmelt/internal/runway"
	"github.com/lacsar712/snowmelt/internal/clock"
	"github.com/lacsar712/snowmelt/internal/heatbank"
	"github.com/lacsar712/snowmelt/internal/config"
	"github.com/lacsar712/snowmelt/internal/glycol"
	"github.com/lacsar712/snowmelt/internal/fsm"
	"github.com/lacsar712/snowmelt/internal/interlock"
	"github.com/lacsar712/snowmelt/internal/model"
	"github.com/lacsar712/snowmelt/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.RunwayFSM
	runway        *runway.Controller
	heatbank    *heatbank.Coordinator
	glycol          *glycol.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	preheatWindow   *clock.PreheatWindow
	warmupWindow  *clock.HeatbankWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	loopLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewRunwayFSM(cfg.UnitID),
		runway:       runway.NewController(clk),
		heatbank:   heatbank.NewCoordinator(clk),
		glycol:         glycol.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		preheatWindow:  clock.NewPreheatWindow(clk),
		warmupWindow: clock.NewHeatbankWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		loopLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.RunwayFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetGlycol(a.glycol.Level().WithinLimits(snap.Glycol.LevelPercent))
	a.permissives.SetPressure(a.runway.Pressure().WithinTripLimits(snap.Runway.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetHeatbank(a.heatbank.Burner().MeltStable(snap.Heatbank))
	a.permissives.SetLoop(snap.Heatbank.LoopFlowTPH > 0 || snap.State == model.StatePreheat)
	a.permissives.SetIgnition(snap.Heatbank.BurnerPhase == model.BurnerStable || snap.Heatbank.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetLoopPermissive(a.permissives.LoopOK())
	a.fsm.SetPreheatComplete(a.preheatWindow.Ready(snap.Heatbank.PreheatStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
