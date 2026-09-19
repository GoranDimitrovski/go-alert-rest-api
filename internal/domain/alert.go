package domain

import (
	"errors"
	"time"
)

// ErrShortStreak guards Alert's invariant: it always records the full streak
// of readings that triggered it.
var ErrShortStreak = errors.New("alert needs a full streak of measurements")

// Alert is the record of a sensor crossing into StatusAlert.
type Alert struct {
	ID        uint
	SensorID  uint
	Levels    [AlertStreak]uint // oldest first
	StartTime time.Time
	EndTime   time.Time
}

// NewAlert builds the alert for streak, the triggering measurements ordered
// newest first (as repositories return them).
func NewAlert(streak []Measurement) (Alert, error) {
	if len(streak) < AlertStreak {
		return Alert{}, ErrShortStreak
	}
	streak = streak[:AlertStreak]

	alert := Alert{
		SensorID:  streak[0].SensorID,
		StartTime: streak[AlertStreak-1].RecordedAt,
		EndTime:   streak[0].RecordedAt,
	}
	for i, m := range streak {
		alert.Levels[AlertStreak-1-i] = m.Level
	}
	return alert, nil
}
