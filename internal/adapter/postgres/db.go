// Package postgres is the persistence adapter: it implements the domain
// repository ports on top of GORM and keeps every SQL detail behind them.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

func Open(ctx context.Context, dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.New(zerologWriter{}, gormlogger.Config{
			SlowThreshold: time.Second,
			LogLevel:      gormlogger.Warn,
		}),
		TranslateError:         true,
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}

// ponytail: AutoMigrate is enough while the schema is append-only; move to
// versioned migrations (goose/atlas) the first time a column has to change.
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&sensorRow{}, &measurementRow{}, &alertRow{})
}

type zerologWriter struct{}

func (zerologWriter) Printf(format string, args ...any) { log.Warn().Msgf(format, args...) }

type txKey struct{}

var _ domain.UnitOfWork = UnitOfWork{}

// UnitOfWork runs work inside a single database transaction.
type UnitOfWork struct{ db *gorm.DB }

func NewUnitOfWork(db *gorm.DB) UnitOfWork { return UnitOfWork{db: db} }

func (u UnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return fn(ctx) // already inside a transaction
	}
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// conn returns the transaction carried by ctx, or the plain connection pool.
func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

func Ping(db *gorm.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
