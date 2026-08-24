package store

import "github.com/lacsar712/snowmelt/internal/model"

type GlycolSnapshotView struct {
	UnitID   string
	Glycol     model.GlycolReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneGlycolSnapshot(s model.PlantSnapshot) GlycolSnapshotView {
	out := GlycolSnapshotView{
		UnitID:   s.UnitID,
		Glycol:     s.Glycol,
		Revision: s.Revision,
	}
	out.Alarms = s.Alarms[:len(s.Alarms):len(s.Alarms)]
	return out
}
