package fsm

import "github.com/lacsar712/snowmelt/internal/model"

var transitionTable = map[model.PlantState]map[PlantEvent]model.PlantState{
	model.StateColdStandby: {
		EvStartPreheat:   model.StatePreheat,
		EvEnterService: model.StateService,
	},
	model.StatePreheat: {
		EvPreheatComplete: model.StateIgnition,
		EvTrip:          model.StateTrip,
		EvShutdown:      model.StateColdStandby,
	},
	model.StateIgnition: {
		EvIgnitionStable: model.StateRamp,
		EvTrip:           model.StateTrip,
		EvShutdown:       model.StateColdStandby,
	},
	model.StateRamp: {
		EvReachFiring: model.StateFiring,
		EvTrip:        model.StateTrip,
		EvShutdown:    model.StateColdStandby,
	},
	model.StateFiring: {
		EvLoadFollow: model.StateLoadFollow,
		EvTrip:       model.StateTrip,
		EvShutdown:   model.StateColdStandby,
	},
	model.StateLoadFollow: {
		EvReachFiring: model.StateFiring,
		EvTrip:        model.StateTrip,
		EvShutdown:    model.StateColdStandby,
	},
	model.StateTrip: {
		EvResetTrip: model.StateColdStandby,
	},
	model.StateService: {
		EvLeaveService: model.StateColdStandby,
	},
}

func NextState(current model.PlantState, event PlantEvent) (model.PlantState, bool) {
	events, ok := transitionTable[current]
	if !ok {
		return current, false
	}
	next, ok := events[event]
	return next, ok
}

func AllowedEvents(state model.PlantState) []PlantEvent {
	events, ok := transitionTable[state]
	if !ok {
		return nil
	}
	out := make([]PlantEvent, 0, len(events))
	for ev := range events {
		out = append(out, ev)
	}
	return out
}

func IsTerminal(state model.PlantState) bool {
	return state == model.StateTrip
}
