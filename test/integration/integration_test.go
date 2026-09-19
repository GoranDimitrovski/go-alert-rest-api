//go:build integration

// Package integration exercises the real adapters against a real
// PostgreSQL instance. Run it with the compose stack up:
//
//	docker compose run --rm go test -tags=integration ./...
package integration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	httpadapter "github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/http"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/postgres"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/config"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

const testDatabase = "alarm_test"

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// openTestDB migrates and empties a database kept separate from the one the
// running api uses, so tests never destroy development data.
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := postgres.Open(ctx, config.Load().DSN)
	if err != nil {
		t.Fatalf("connecting to postgres: %v (is the compose stack up?)", err)
	}
	// Errors here mean the database already exists, which is the normal case.
	admin.Exec("CREATE DATABASE " + testDatabase)
	postgres.Close(admin)

	dsn, err := url.Parse(config.Load().DSN)
	if err != nil {
		t.Fatalf("parsing dsn: %v", err)
	}
	dsn.Path = "/" + testDatabase

	db, err := postgres.Open(ctx, dsn.String())
	if err != nil {
		t.Fatalf("connecting to %s: %v", testDatabase, err)
	}
	t.Cleanup(func() { postgres.Close(db) })

	if err := postgres.Migrate(db); err != nil {
		t.Fatalf("migrating: %v", err)
	}
	if err := db.Exec("TRUNCATE alert, measurement, sensor RESTART IDENTITY CASCADE").Error; err != nil {
		t.Fatalf("truncating: %v", err)
	}
	return db
}

func TestSensorRepository(t *testing.T) {
	ctx := context.Background()
	repo := postgres.NewSensorRepository(openTestDB(t))

	if _, err := repo.ByID(ctx, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("ByID() on an empty table = %v, want %v", err, domain.ErrNotFound)
	}

	if err := repo.Save(ctx, domain.NewSensor(1)); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	sensor, err := repo.ByID(ctx, 1)
	if err != nil {
		t.Fatalf("ByID() error = %v", err)
	}
	if sensor.Status != domain.StatusOK {
		t.Errorf("status = %q, want %q", sensor.Status, domain.StatusOK)
	}

	// Saving again must update the row, not fail on the primary key.
	if err := repo.Save(ctx, domain.Sensor{ID: 1, Status: domain.StatusAlert}); err != nil {
		t.Fatalf("Save() on an existing sensor error = %v", err)
	}
	if sensor, _ := repo.ByID(ctx, 1); sensor.Status != domain.StatusAlert {
		t.Errorf("status = %q, want %q after update", sensor.Status, domain.StatusAlert)
	}
}

func TestMeasurementRepositoryOrdersAndLimits(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	sensors := postgres.NewSensorRepository(db)
	measurements := postgres.NewMeasurementRepository(db)

	if err := sensors.Save(ctx, domain.NewSensor(1)); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	base := time.Now().UTC().Truncate(time.Millisecond)
	for i, level := range []uint{100, 200, 300, 400} {
		m := domain.Measurement{SensorID: 1, Level: level, RecordedAt: base.Add(time.Duration(i) * time.Minute)}
		if err := measurements.Add(ctx, &m); err != nil {
			t.Fatalf("Add() error = %v", err)
		}
		if m.ID == 0 {
			t.Fatal("Add() did not set the id")
		}
	}

	recent, err := measurements.Recent(ctx, 1, domain.AlertStreak)
	if err != nil {
		t.Fatalf("Recent() error = %v", err)
	}
	if len(recent) != domain.AlertStreak {
		t.Fatalf("Recent() returned %d, want %d", len(recent), domain.AlertStreak)
	}
	if got := []uint{recent[0].Level, recent[1].Level, recent[2].Level}; got[0] != 400 || got[1] != 300 || got[2] != 200 {
		t.Errorf("Recent() levels = %v, want [400 300 200] (newest first)", got)
	}
	if !recent[0].RecordedAt.Equal(base.Add(3 * time.Minute)) {
		t.Errorf("RecordedAt = %v, want %v back from postgres unchanged", recent[0].RecordedAt, base.Add(3*time.Minute))
	}

	all, err := measurements.BySensor(ctx, 1)
	if err != nil {
		t.Fatalf("BySensor() error = %v", err)
	}
	if len(all) != 4 {
		t.Errorf("BySensor() returned %d, want 4", len(all))
	}
	if other, _ := measurements.BySensor(ctx, 2); len(other) != 0 {
		t.Errorf("BySensor(2) returned %d, want 0", len(other))
	}
}

func TestAlertRepositoryKeepsLevelOrder(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	if err := postgres.NewSensorRepository(db).Save(ctx, domain.NewSensor(1)); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	repo := postgres.NewAlertRepository(db)
	start := time.Now().UTC().Truncate(time.Millisecond)
	alert := domain.Alert{
		SensorID:  1,
		Levels:    [domain.AlertStreak]uint{2100, 2200, 2300},
		StartTime: start,
		EndTime:   start.Add(time.Hour),
	}
	if err := repo.Add(ctx, &alert); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	stored, err := repo.BySensor(ctx, 1)
	if err != nil {
		t.Fatalf("BySensor() error = %v", err)
	}
	if len(stored) != 1 {
		t.Fatalf("alerts = %d, want 1", len(stored))
	}
	if stored[0].Levels != alert.Levels {
		t.Errorf("levels = %v, want %v", stored[0].Levels, alert.Levels)
	}
	if !stored[0].StartTime.Equal(start) {
		t.Errorf("start = %v, want %v", stored[0].StartTime, start)
	}
}

func TestUnitOfWorkRollsBack(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)
	sensors := postgres.NewSensorRepository(db)
	measurements := postgres.NewMeasurementRepository(db)
	uow := postgres.NewUnitOfWork(db)

	wanted := errors.New("something went wrong later in the use case")
	err := uow.Do(ctx, func(ctx context.Context) error {
		if err := sensors.Save(ctx, domain.NewSensor(1)); err != nil {
			return err
		}
		m := domain.Measurement{SensorID: 1, Level: 100, RecordedAt: time.Now().UTC()}
		if err := measurements.Add(ctx, &m); err != nil {
			return err
		}
		return wanted
	})
	if !errors.Is(err, wanted) {
		t.Fatalf("Do() error = %v, want %v", err, wanted)
	}

	if _, err := sensors.ByID(ctx, 1); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("sensor survived the rollback: %v", err)
	}
	if stored, _ := measurements.BySensor(ctx, 1); len(stored) != 0 {
		t.Errorf("measurements = %d, want 0 after a rollback", len(stored))
	}
}

func TestAPIOnPostgres(t *testing.T) {
	db := openTestDB(t)
	sensors := postgres.NewSensorRepository(db)
	measurements := postgres.NewMeasurementRepository(db)
	alerts := postgres.NewAlertRepository(db)

	router := httpadapter.NewRouter(
		app.NewMeasurementService(measurements, sensors, alerts, postgres.NewUnitOfWork(db)),
		app.NewAlertService(alerts),
		app.NewSensorService(sensors),
		postgres.Ping(db),
	)

	call := func(method, path, body string) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	status := func() string {
		t.Helper()
		rec := call(http.MethodGet, "/v1/sensors/3", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("GET sensor = %d, want 200 (%s)", rec.Code, rec.Body)
		}
		var sensor httpadapter.SensorResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &sensor); err != nil {
			t.Fatalf("decoding sensor: %v", err)
		}
		return sensor.Status
	}

	if rec := call(http.MethodGet, "/healthz", ""); rec.Code != http.StatusOK {
		t.Fatalf("healthz = %d, want 200", rec.Code)
	}

	// An unknown sensor registers itself on its first reading.
	if rec := call(http.MethodPost, "/v1/measurements", `{"sensor_id":3,"level":1000}`); rec.Code != http.StatusCreated {
		t.Fatalf("POST = %d, want 201 (%s)", rec.Code, rec.Body)
	}
	if got := status(); got != string(domain.StatusOK) {
		t.Fatalf("status = %q, want OK", got)
	}

	for _, level := range []string{"2100", "2200"} {
		call(http.MethodPost, "/v1/measurements", `{"sensor_id":3,"level":`+level+`}`)
	}
	if got := status(); got != string(domain.StatusWarn) {
		t.Fatalf("status = %q, want WARN", got)
	}

	call(http.MethodPost, "/v1/measurements", `{"sensor_id":3,"level":2300}`)
	if got := status(); got != string(domain.StatusAlert) {
		t.Fatalf("status = %q, want ALERT", got)
	}

	var raised []httpadapter.AlertResponse
	rec := call(http.MethodGet, "/v1/sensors/3/alerts", "")
	if err := json.Unmarshal(rec.Body.Bytes(), &raised); err != nil {
		t.Fatalf("decoding alerts: %v", err)
	}
	if len(raised) != 1 {
		t.Fatalf("alerts = %d, want 1", len(raised))
	}
	if want := ([domain.AlertStreak]uint{2100, 2200, 2300}); raised[0].Levels != want {
		t.Errorf("levels = %v, want %v", raised[0].Levels, want)
	}

	// Everything the API reported is actually in the database.
	var rows int64
	if err := db.Table("measurement").Where("sensor_id = ?", 3).Count(&rows).Error; err != nil {
		t.Fatalf("counting measurements: %v", err)
	}
	if rows != 4 {
		t.Errorf("measurement rows = %d, want 4", rows)
	}
}

func TestMeasurementRequiresAnExistingSensorRow(t *testing.T) {
	// The foreign key is the database's job, not the application's.
	db := openTestDB(t)
	err := postgres.NewMeasurementRepository(db).Add(context.Background(), &domain.Measurement{
		SensorID: 999, Level: 100, RecordedAt: time.Now().UTC(),
	})
	if err == nil {
		t.Fatal("Add() for an unknown sensor succeeded, want a foreign key violation")
	}
}
