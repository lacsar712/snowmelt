package interlock

import (
	"fmt"

	"github.com/lacsar712/snowmelt/internal/model"
)

type PermissiveSet struct {
	loopOK       bool
	ignitionOK   bool
	glycolOK       bool
	pressureOK   bool
	heatbankOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetLoop(ok bool)       { p.loopOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetGlycol(ok bool)       { p.glycolOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetHeatbank(ok bool) { p.heatbankOK = ok }

func (p *PermissiveSet) LoopOK() bool       { return p.loopOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) GlycolOK() bool       { return p.glycolOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) HeatbankOK() bool { return p.heatbankOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.loopOK && p.ignitionOK && p.glycolOK && p.pressureOK && p.heatbankOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.loopOK {
		return fmt.Errorf("%w", model.ErrLoopPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckMeltLoss(reading model.HeatbankReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.ZonelockTempF < 600 {
		return fmt.Errorf("%w", model.ErrMeltLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.glycolOK {
		return fmt.Errorf("%w", model.ErrGlycolLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.heatbankOK {
		return fmt.Errorf("%w", model.ErrHeatbankTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
