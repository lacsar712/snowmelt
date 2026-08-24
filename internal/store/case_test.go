package store_test

import (
	"testing"
	"time"

	"github.com/lacsar712/snowmelt/internal/model"
	"github.com/lacsar712/snowmelt/internal/store"
)

func TestCase(t *testing.T) {
	snap := model.PlantSnapshot{
		UnitID: "EXP-1",
		Glycol: model.GlycolReading{LevelPercent: 55},
		Alarms: []model.AlarmEvent{{
			Code:     "LVL-HI",
			Severity: "warning",
			Message:  "level high",
			RaisedAt: time.Now(),
			Active:   true,
		}},
	}
	clone := store.CloneGlycolSnapshot(snap)
	if len(clone.Alarms) != 1 {
		t.Fatal("expected one alarm in clone")
	}
	clone.Alarms[0].Code = "MUTATED"
	if snap.Alarms[0].Code == "MUTATED" {
		t.Fatal("mutating clone alarms should not affect source snapshot")
	}
}
