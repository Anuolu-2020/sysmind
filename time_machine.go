package main

import (
	"context"

	"sysmind/internal/models"
)

// GetTimeMachine returns persisted historical samples, annotations, and forecasts.
func (a *App) GetTimeMachine(hours int) models.TimeMachineView {
	if a.timeMachine == nil {
		return models.TimeMachineView{
			WindowHours:        hours,
			RetentionHours:     0,
			SamplingSeconds:    0,
			Samples:            []models.TimeMachineSample{},
			Annotations:        []models.TimeMachineAnnotation{},
			Forecasts:          []models.TimeMachineForecast{},
			Summary:            "Time machine storage is unavailable.",
			PersistenceEnabled: false,
		}
	}
	return a.timeMachine.GetView(hours)
}

// shutdown flushes persisted stores before exit.
func (a *App) shutdown(ctx context.Context) {
	_ = ctx
	if a.timeMachine != nil {
		a.timeMachine.Save()
	}
}
