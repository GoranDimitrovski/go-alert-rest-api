# go-alert-rest-api

REST API that collects CO2 readings and alerts when a sensor stays unhealthy.
Go 1.27 · gin · GORM · PostgreSQL 18, all in Docker.

## Setup

- Docker is the only prerequisite; `make` just shortens the commands.
- Every command runs the Go toolchain in a container — nothing else is installed.

### Get the code

```bash
git clone git@github.com:GoranDimitrovski/go-alert-rest-api.git
cd go-alert-rest-api
cp .env.example .env    # optional, same values as the defaults
```

### Start it — api on `:8080`, postgres on `:5432`

```make
make up
```

```bash
docker compose up --build -d
```

### Check it

```bash
curl localhost:8080/healthz   # {"status":"ok"}
```

### Day to day

```make
make          # list every target
make logs     # follow the api logs
make psql     # a psql shell on the database
make down     # stop (ARGS=-v also drops the data)
make clean    # stop, remove volumes, images and coverage.out
```

```bash
docker compose logs -f api
docker compose exec postgres psql -U myuser -d alarm
docker compose down            # -v also drops the data
docker compose down -v --rmi local && rm -f coverage.out
```

## Tests

```make
make test               # unit + API, no database
make test-integration   # + the PostgreSQL suite
make cover              # coverage.out and a total
make check              # fmt, vet, lint and test
```

```bash
docker compose run --rm --no-deps go test ./test/...
docker compose run --rm go test -tags=integration -count=1 ./test/...
docker compose run --rm go test -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./...
docker compose run --rm --no-deps go tool cover -func=coverage.out
```

`make test` prints one line per test and a summary:

```
PASS test/domain.TestEvaluateStatus/streak_broken (0.00s)
PASS test/api.TestPostMeasurementRejectsBadRequests/level_zero (0.00s)
...
DONE 46 tests in 0.018s
```

- `FORMAT=pkgname make test` collapses it to one line per package.
- A failure prints its output, then repeats it under a `=== Failed` summary.
- `--no-deps` skips starting PostgreSQL for suites that do not need it.
- The integration suite sits behind the `integration` build tag.
- It creates and truncates its own `alarm_test` database, never your data.

All tests live under `test/`, one folder per suite:

| Folder | Covers | |
| --- | --- | --- |
| `test/domain` | thresholds, streaks, recovery, alert building | 96% |
| `test/app` | the use cases over in-memory repositories | 89% |
| `test/api` | every route: status codes, JSON shapes, bad payloads | 93% |
| `test/config` | defaults, overrides, escaping, redaction | 93% |
| `test/integration` | repositories and API on real PostgreSQL | 89% |

Total across `internal/`: **92.9%**. Because the tests sit outside the packages
they exercise, coverage is measured with `-coverpkg=./internal/...`, which is
what `make cover` does.

### Every target and its equivalent

| Target | Without `make` |
| --- | --- |
| `make up` | `docker compose up --build -d` |
| `make down` | `docker compose down` |
| `make restart` | `docker compose down && docker compose up --build -d` |
| `make logs` | `docker compose logs -f api` |
| `make psql` | `docker compose exec postgres psql -U myuser -d alarm` |
| `make build` | `docker compose run --rm --no-deps go build ./...` |
| `make test` | `docker compose run --rm --no-deps go run gotest.tools/gotestsum@v1.13.0 --format testname -- ./test/...` |
| `make test-integration` | the same, plus `-tags=integration -count=1` |
| `make cover` | `docker compose run --rm go test -tags=integration -coverpkg=./internal/... -coverprofile=coverage.out ./...` |
| `make vet` | `docker compose run --rm --no-deps go vet -tags=integration ./...` |
| `make lint` | `docker compose run --rm --no-deps go run honnef.co/go/tools/cmd/staticcheck@latest -tags=integration ./...` |
| `make fmt` | `docker compose run --rm --no-deps go fmt ./...` |
| `make tidy` | `docker compose run --rm --no-deps go mod tidy` |
| `make clean` | `docker compose down -v --rmi local && rm -f coverage.out` |

## Behaviour

- A reading above **2000 ppm** is unhealthy.
- One unhealthy reading → `WARN`; three in a row → `ALERT`.
- Entering `ALERT` records an alert with those three readings and their time
  span — one per streak.
- Any healthy reading resets the sensor to `OK`.
- Sensors register themselves on their first reading.

## Endpoints

| | | |
| --- | --- | --- |
| `POST` | `/v1/measurements` | `{"sensor_id":42,"level":2100}` → `201` |
| `GET` | `/v1/sensors/:id` | `{"id":42,"status":"ALERT"}`, `404` if unknown |
| `GET` | `/v1/sensors/:id/measurements` | readings, newest first |
| `GET` | `/v1/sensors/:id/alerts` | `[{"levels":[2100,2200,2300],"start_time":…}]` |
| `GET` | `/healthz` | `200`, or `503` if the database is unreachable |

- Rejected requests: `{"error":"…"}` with a `4xx`.
- Unexpected failures: `500`, with the cause in the log only.

## Configuration

- Compose reads `.env` if present; `docker-compose.yml` holds the same defaults.
- `POSTGRES_DB` / `POSTGRES_USER` / `POSTGRES_PASSWORD` — `alarm`, `myuser`, `mypassword`
- `API_PORT` `8080` · `POSTGRES_PORT` `5432`
- `LOG_LEVEL` `info` · `DB_SSLMODE` `disable` · `SHUTDOWN_TIMEOUT` `10s`

## Layout

- `cmd/api` — configuration, wiring, graceful shutdown
- `internal/domain` — entities, rules, repository ports; no I/O
- `internal/app` — use cases
- `internal/adapter/http` — handlers, request/response types, error mapping
- `internal/adapter/postgres` — GORM repositories, unit of work
- `internal/adapter/memory` — the same ports in memory, so tests can skip the database
- `internal/config` — the only place that reads the environment
- `test/` — every test suite, one folder each (see [Tests](#tests))

- Dependencies point inwards, and the ports live in `domain`, so the rules never
  see GORM.
- Writes that must agree — reading, status, alert — share one transaction via
  `domain.UnitOfWork`.
