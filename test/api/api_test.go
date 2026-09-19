package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	httpadapter "github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/http"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/adapter/memory"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/app"
	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	m.Run()
}

// testAPI builds the real router over in-memory repositories.
func testAPI(t *testing.T, pingErr error) (*gin.Engine, *memory.Store) {
	t.Helper()

	store := memory.New()
	router := httpadapter.NewRouter(
		app.NewMeasurementService(store.Measurements(), store.Sensors(), store.Alerts(), store.UnitOfWork()),
		app.NewAlertService(store.Alerts()),
		app.NewSensorService(store.Sensors()),
		func(context.Context) error { return pingErr },
	)
	return router, store
}

func call(t *testing.T, router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

// decode reads the body into the DTO the endpoint promises.
func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()

	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decoding %q: %v", rec.Body.String(), err)
	}
	return out
}

func assertStatus(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, want, rec.Body.String())
	}
}

// post records a reading and fails the test if the API rejects it.
func post(t *testing.T, router *gin.Engine, body string) httpadapter.MeasurementResponse {
	t.Helper()

	rec := call(t, router, http.MethodPost, "/v1/measurements", body)
	assertStatus(t, rec, http.StatusCreated)
	return decode[httpadapter.MeasurementResponse](t, rec)
}

func TestHealthz(t *testing.T) {
	router, _ := testAPI(t, nil)

	rec := call(t, router, http.MethodGet, "/healthz", "")
	assertStatus(t, rec, http.StatusOK)
	if got := decode[httpadapter.HealthResponse](t, rec).Status; got != "ok" {
		t.Errorf("status = %q, want %q", got, "ok")
	}

	unhealthy, _ := testAPI(t, errors.New("connection refused"))
	rec = call(t, unhealthy, http.MethodGet, "/healthz", "")
	assertStatus(t, rec, http.StatusServiceUnavailable)
	if got := decode[httpadapter.ErrorResponse](t, rec).Error; got != "database unavailable" {
		t.Errorf("error = %q, want %q", got, "database unavailable")
	}
}

func TestPostMeasurement(t *testing.T) {
	router, _ := testAPI(t, nil)

	got := post(t, router, `{"sensor_id":42,"level":1500}`)

	if got.SensorID != 42 {
		t.Errorf("sensor_id = %d, want 42", got.SensorID)
	}
	if got.Level != 1500 {
		t.Errorf("level = %d, want 1500", got.Level)
	}
	if got.ID == 0 {
		t.Error("id is missing from the response")
	}
	if got.RecordedAt.IsZero() {
		t.Error("recorded_at is missing from the response")
	}
}

func TestPostMeasurementRejectsBadRequests(t *testing.T) {
	cases := map[string]string{
		"empty body":         ``,
		"not json":           `not json`,
		"missing level":      `{"sensor_id":1}`,
		"missing sensor id":  `{"level":100}`,
		"level zero":         `{"sensor_id":1,"level":0}`,
		"level out of range": `{"sensor_id":1,"level":500000}`,
		"wrong type":         `{"sensor_id":"one","level":100}`,
	}

	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			router, store := testAPI(t, nil)

			rec := call(t, router, http.MethodPost, "/v1/measurements", body)
			assertStatus(t, rec, http.StatusBadRequest)
			if decode[httpadapter.ErrorResponse](t, rec).Error == "" {
				t.Error("no error message in the response")
			}

			stored, _ := store.Measurements().BySensor(context.Background(), 1)
			if len(stored) != 0 {
				t.Errorf("measurements = %d, want 0 after a rejected request", len(stored))
			}
		})
	}
}

func TestPostMeasurementReportsRepositoryFailure(t *testing.T) {
	router, store := testAPI(t, nil)
	store.FailWith = errors.New("database is down")

	rec := call(t, router, http.MethodPost, "/v1/measurements", `{"sensor_id":1,"level":100}`)
	assertStatus(t, rec, http.StatusInternalServerError)
	if got := decode[httpadapter.ErrorResponse](t, rec).Error; got != "internal server error" {
		t.Errorf("error = %q, want the internal details to stay hidden", got)
	}
}

func TestSensorEndpoints(t *testing.T) {
	router, _ := testAPI(t, nil)
	for _, level := range []string{"2100", "2200", "2300"} {
		post(t, router, `{"sensor_id":9,"level":`+level+`}`)
	}

	t.Run("sensor", func(t *testing.T) {
		rec := call(t, router, http.MethodGet, "/v1/sensors/9", "")
		assertStatus(t, rec, http.StatusOK)

		got := decode[httpadapter.SensorResponse](t, rec)
		if got.ID != 9 {
			t.Errorf("id = %d, want 9", got.ID)
		}
		if got.Status != string(domain.StatusAlert) {
			t.Errorf("status = %q, want %q", got.Status, domain.StatusAlert)
		}
	})

	t.Run("measurements", func(t *testing.T) {
		rec := call(t, router, http.MethodGet, "/v1/sensors/9/measurements", "")
		assertStatus(t, rec, http.StatusOK)

		got := decode[[]httpadapter.MeasurementResponse](t, rec)
		if len(got) != 3 {
			t.Fatalf("measurements = %d, want 3", len(got))
		}
		if got[0].Level != 2300 {
			t.Errorf("first level = %d, want 2300 (newest first)", got[0].Level)
		}
	})

	t.Run("alerts", func(t *testing.T) {
		rec := call(t, router, http.MethodGet, "/v1/sensors/9/alerts", "")
		assertStatus(t, rec, http.StatusOK)

		got := decode[[]httpadapter.AlertResponse](t, rec)
		if len(got) != 1 {
			t.Fatalf("alerts = %d, want 1", len(got))
		}
		if want := ([domain.AlertStreak]uint{2100, 2200, 2300}); got[0].Levels != want {
			t.Errorf("levels = %v, want %v (oldest first)", got[0].Levels, want)
		}
		if got[0].SensorID != 9 {
			t.Errorf("sensor_id = %d, want 9", got[0].SensorID)
		}
		if got[0].StartTime.After(got[0].EndTime) {
			t.Errorf("start %v is after end %v", got[0].StartTime, got[0].EndTime)
		}
	})
}

func TestUnknownSensor(t *testing.T) {
	router, _ := testAPI(t, nil)

	rec := call(t, router, http.MethodGet, "/v1/sensors/404", "")
	assertStatus(t, rec, http.StatusNotFound)
	if got := decode[httpadapter.ErrorResponse](t, rec).Error; got != "not found" {
		t.Errorf("error = %q, want %q", got, "not found")
	}

	// Collections of an unknown sensor are empty, not missing, and serialise as
	// [] rather than null.
	for _, path := range []string{"/v1/sensors/404/measurements", "/v1/sensors/404/alerts"} {
		rec := call(t, router, http.MethodGet, path, "")
		assertStatus(t, rec, http.StatusOK)
		if body := strings.TrimSpace(rec.Body.String()); body != "[]" {
			t.Errorf("GET %s body = %s, want []", path, body)
		}
	}
}

func TestInvalidSensorIDInPath(t *testing.T) {
	router, _ := testAPI(t, nil)

	for _, path := range []string{"/v1/sensors/abc", "/v1/sensors/0", "/v1/sensors/-1", "/v1/sensors/99999999999999999999/alerts"} {
		rec := call(t, router, http.MethodGet, path, "")
		assertStatus(t, rec, http.StatusBadRequest)
	}
}

func TestUnknownRoute(t *testing.T) {
	router, _ := testAPI(t, nil)

	assertStatus(t, call(t, router, http.MethodGet, "/v1/nope", ""), http.StatusNotFound)
	assertStatus(t, call(t, router, http.MethodDelete, "/v1/measurements", ""), http.StatusNotFound)
}
