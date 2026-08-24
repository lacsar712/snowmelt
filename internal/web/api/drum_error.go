package api

import (
	"errors"

	"github.com/lacsar712/snowmelt/internal/model"
)

func classifyGlycolError(err error) (string, bool) {
	if errors.Is(err, model.ErrGlycolLevelLow) {
		return "glycol_level_low", true
	}
	return "", false
}
