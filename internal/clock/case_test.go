package clock

import (
	"testing"
	"time"
)

func TestCase(t *testing.T) {
	clk := NewManual(time.Unix(0, 0))
	p := NewPreheatWindow(clk)
	anchor := clk.Now()
	if p.Ready(anchor) {
		t.Fatal("purge window should not be satisfied before process clock advances")
	}
}
