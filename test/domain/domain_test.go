package domain_test

import (
	"testing"
	"time"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

// newest returns measurements in repository order (newest first).
func newest(levels ...uint) []domain.Measurement {
	out := make([]domain.Measurement, len(levels))
	base := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	for i, level := range levels {
		out[i] = domain.Measurement{SensorID: 1, Level: level, RecordedAt: base.Add(-time.Duration(i) * 30 * time.Minute)}
	}
	return out
}

func TestEvaluateStatus(t *testing.T) {
	cases := map[string]struct {
		recent []domain.Measurement
		want   domain.Status
	}{
		"no measurements":       {nil, domain.StatusOK},
		"below threshold":       {newest(1500), domain.StatusOK},
		"exactly at threshold":  {newest(2000, 2001, 2002), domain.StatusOK},
		"one above":             {newest(2001), domain.StatusWarn},
		"two above":             {newest(2001, 2002), domain.StatusWarn},
		"three above":           {newest(2001, 2002, 2003), domain.StatusAlert},
		"streak broken":         {newest(2001, 1000, 2003), domain.StatusWarn},
		"older reading ignored": {newest(2001, 2002, 2003, 100), domain.StatusAlert},
		"recovered":             {newest(100, 2002, 2003), domain.StatusOK},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			if got := domain.EvaluateStatus(tc.recent); got != tc.want {
				t.Fatalf("EvaluateStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewAlert(t *testing.T) {
	recent := newest(2003, 2002, 2001) // newest first

	alert, err := domain.NewAlert(recent)
	if err != nil {
		t.Fatalf("NewAlert() error = %v", err)
	}
	if want := ([domain.AlertStreak]uint{2001, 2002, 2003}); alert.Levels != want {
		t.Errorf("Levels = %v, want %v (oldest first)", alert.Levels, want)
	}
	if !alert.StartTime.Equal(recent[domain.AlertStreak-1].RecordedAt) {
		t.Errorf("StartTime = %v, want the oldest reading", alert.StartTime)
	}
	if !alert.EndTime.Equal(recent[0].RecordedAt) {
		t.Errorf("EndTime = %v, want the newest reading", alert.EndTime)
	}

	if _, err := domain.NewAlert(newest(2001, 2002)); err == nil {
		t.Error("NewAlert() with a short streak: expected an error")
	}
}

func TestNewMeasurementRejectsNonsense(t *testing.T) {
	now := time.Now()
	if _, err := domain.NewMeasurement(0, 100, now); err == nil {
		t.Error("expected an error for sensor id 0")
	}
	if _, err := domain.NewMeasurement(1, 0, now); err == nil {
		t.Error("expected an error for level 0")
	}
	if _, err := domain.NewMeasurement(1, domain.MaxLevel+1, now); err == nil {
		t.Error("expected an error for an impossible level")
	}
}
