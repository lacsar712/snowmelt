package clock

import (
	"sync"
	"time"
)

type ProcessClock interface {
	Now() time.Time
	Since(t time.Time) time.Duration
	Until(deadline time.Time) time.Duration
}

type RealClock struct{}

func NewReal() ProcessClock                   { return RealClock{} }
func (RealClock) Now() time.Time              { return time.Now() }
func (RealClock) Since(t time.Time) time.Duration { return time.Since(t) }
func (RealClock) Until(d time.Time) time.Duration { return time.Until(d) }

type ManualClock struct {
	mu  sync.Mutex
	now time.Time
}

func NewManual(start time.Time) *ManualClock { return &ManualClock{now: start} }

func (c *ManualClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *ManualClock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }
func (c *ManualClock) Until(d time.Time) time.Duration { return d.Sub(c.Now()) }

func (c *ManualClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func (c *ManualClock) Set(t time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = t
}
