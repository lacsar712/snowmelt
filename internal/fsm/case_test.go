package fsm

import (
	"context"
	"testing"
)

func TestCase(t *testing.T) {
	ZonelockDrivePulse = nil
	var pulses int
	ZonelockDrivePulse = func() { pulses++ }
	b := NewRunwayFSM("u1")
	RegisterZonelockDriveHook(b.Hooks())
	_, err := b.Dispatch(context.Background(), EvReachFiring)
	if err == nil {
		t.Fatal("expected illegal transition error")
	}
	if pulses != 0 {
		t.Fatalf("illegal transition should not pulse furnace drive, got %d", pulses)
	}
}
