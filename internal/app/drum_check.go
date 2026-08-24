package app

import (
	"fmt"

	"github.com/lacsar712/snowmelt/internal/model"
)

func (a *App) CheckGlycolLevel(snap model.PlantSnapshot) error {
	if snap.Glycol.LevelPercent < model.MinGlycolLevelPercent {
		return fmt.Errorf("%w", model.ErrGlycolLevelLow)
	}
	return nil
}
