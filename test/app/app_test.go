package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/memory"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

func newMeasurementService(store *memory.Store) *app.MeasurementService {
	return app.NewMeasurementService(store.Measurements(), store.Sensors(), store.Alerts(), store.UnitOfWork())
}

// record fails the test if the reading is rejected.
func record(t *testing.T, service *app.MeasurementService, sensorID, level uint) app.MeasurementDTO {
	t.Helper()
	m, err := service.Record(context.Background(), app.RecordMeasurementCommand{SensorID: sensorID, Level: level})
	if err != nil {
		t.Fatalf("Record(%d, %d) error = %v", sensorID, level, err)
	}
	return m
}

func statusOf(t *testing.T, store *memory.Store, sensorID uint) domain.Status {
	t.Helper()
	sensor, err := store.Sensors().ByID(context.Background(), sensorID)
	if err != nil {
		t.Fatalf("sensor %d: %v", sensorID, err)
	}
	return sensor.Status
}

func TestRecordRegistersUnknownSensor(t *testing.T) {
	store := memory.New()
	service := newMeasurementService(store)

	m := record(t, service, 7, 500)

	if m.ID == 0 {
		t.Error("stored measurement has no id")
	}
	if m.RecordedAt.IsZero() {
		t.Error("stored measurement has no timestamp")
	}
	if got := statusOf(t, store, 7); got != domain.StatusOK {
		t.Errorf("status = %q, want %q", got, domain.StatusOK)
	}
	if alerts, _ := store.Alerts().BySensor(context.Background(), 7); len(alerts) != 0 {
		t.Errorf("alerts = %d, want 0", len(alerts))
	}
}

func TestRecordRaisesOneAlertPerStreak(t *testing.T) {
	store := memory.New()
	service := newMeasurementService(store)
	ctx := context.Background()

	record(t, service, 1, 2100)
	record(t, service, 1, 2200)
	if got := statusOf(t, store, 1); got != domain.StatusWarn {
		t.Fatalf("after two high readings status = %q, want %q", got, domain.StatusWarn)
	}

	record(t, service, 1, 2300)
	if got := statusOf(t, store, 1); got != domain.StatusAlert {
		t.Fatalf("after three high readings status = %q, want %q", got, domain.StatusAlert)
	}

	alerts, _ := store.Alerts().BySensor(ctx, 1)
	if len(alerts) != 1 {
		t.Fatalf("alerts = %d, want 1", len(alerts))
	}
	if want := ([domain.AlertStreak]uint{2100, 2200, 2300}); alerts[0].Levels != want {
		t.Errorf("alert levels = %v, want %v (oldest first)", alerts[0].Levels, want)
	}
	if alerts[0].StartTime.After(alerts[0].EndTime) {
		t.Errorf("start %v is after end %v", alerts[0].StartTime, alerts[0].EndTime)
	}

	// Staying unhealthy must not raise the same alert again.
	record(t, service, 1, 2400)
	if alerts, _ := store.Alerts().BySensor(ctx, 1); len(alerts) != 1 {
		t.Errorf("alerts = %d, want 1 while still in alert", len(alerts))
	}

	// A healthy reading clears the sensor, and a new streak alerts again.
	record(t, service, 1, 400)
	if got := statusOf(t, store, 1); got != domain.StatusOK {
		t.Fatalf("status = %q, want %q after recovery", got, domain.StatusOK)
	}
	record(t, service, 1, 2500)
	record(t, service, 1, 2600)
	record(t, service, 1, 2700)
	if alerts, _ := store.Alerts().BySensor(ctx, 1); len(alerts) != 2 {
		t.Errorf("alerts = %d, want 2 after a second streak", len(alerts))
	}
}

func TestRecordKeepsSensorsApart(t *testing.T) {
	store := memory.New()
	service := newMeasurementService(store)

	for _, level := range []uint{2100, 2200, 2300} {
		record(t, service, 1, level)
		record(t, service, 2, 100)
	}

	if got := statusOf(t, store, 1); got != domain.StatusAlert {
		t.Errorf("sensor 1 status = %q, want %q", got, domain.StatusAlert)
	}
	if got := statusOf(t, store, 2); got != domain.StatusOK {
		t.Errorf("sensor 2 status = %q, want %q", got, domain.StatusOK)
	}
	if alerts, _ := store.Alerts().BySensor(context.Background(), 2); len(alerts) != 0 {
		t.Errorf("sensor 2 alerts = %d, want 0", len(alerts))
	}
}

func TestRecordRejectsInvalidInput(t *testing.T) {
	cases := map[string]struct {
		sensorID, level uint
		want            error
	}{
		"sensor id zero":   {0, 100, domain.ErrInvalidSensor},
		"level zero":       {1, 0, domain.ErrInvalidLevel},
		"level too high":   {1, domain.MaxLevel + 1, domain.ErrInvalidLevel},
		"level at the max": {1, domain.MaxLevel, nil},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			store := memory.New()
			_, err := newMeasurementService(store).Record(context.Background(), app.RecordMeasurementCommand{SensorID: tc.sensorID, Level: tc.level})
			if !errors.Is(err, tc.want) {
				t.Fatalf("Record() error = %v, want %v", err, tc.want)
			}

			stored, _ := store.Measurements().BySensor(context.Background(), tc.sensorID)
			if tc.want != nil && len(stored) != 0 {
				t.Errorf("measurements = %d, want 0 after a rejected reading", len(stored))
			}
		})
	}
}

func TestRecordPropagatesRepositoryFailure(t *testing.T) {
	store := memory.New()
	store.FailWith = errors.New("database is down")

	_, err := newMeasurementService(store).Record(context.Background(), app.RecordMeasurementCommand{SensorID: 1, Level: 100})
	if err == nil || err.Error() != "database is down" {
		t.Fatalf("Record() error = %v, want the repository error", err)
	}
}

func TestMeasurementsBySensorNewestFirst(t *testing.T) {
	store := memory.New()
	service := newMeasurementService(store)
	for _, level := range []uint{100, 200, 300} {
		record(t, service, 1, level)
	}

	got, err := service.BySensor(context.Background(), 1)
	if err != nil {
		t.Fatalf("BySensor() error = %v", err)
	}
	if len(got) != 3 {
		t.Fatalf("measurements = %d, want 3", len(got))
	}
	if got[0].Level != 300 {
		t.Errorf("first level = %d, want 300 (newest first)", got[0].Level)
	}
}

func TestQueryServices(t *testing.T) {
	store := memory.New()
	service := newMeasurementService(store)
	for _, level := range []uint{2100, 2200, 2300} {
		record(t, service, 5, level)
	}
	ctx := context.Background()

	sensor, err := app.NewSensorService(store.Sensors()).ByID(ctx, 5)
	if err != nil {
		t.Fatalf("SensorService.ByID() error = %v", err)
	}
	if sensor.Status != string(domain.StatusAlert) {
		t.Errorf("status = %q, want %q", sensor.Status, domain.StatusAlert)
	}

	if _, err := app.NewSensorService(store.Sensors()).ByID(ctx, 404); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("unknown sensor error = %v, want %v", err, domain.ErrNotFound)
	}

	alerts, err := app.NewAlertService(store.Alerts()).BySensor(ctx, 5)
	if err != nil {
		t.Fatalf("AlertService.BySensor() error = %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("alerts = %d, want 1", len(alerts))
	}
}
