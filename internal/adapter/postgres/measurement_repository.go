package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

type measurementRow struct {
	ID         uint      `gorm:"primaryKey"`
	SensorID   uint      `gorm:"not null;index:idx_measurement_sensor_recorded,priority:1"`
	Sensor     sensorRow `gorm:"foreignKey:SensorID;constraint:OnDelete:CASCADE"`
	Level      uint      `gorm:"not null"`
	RecordedAt time.Time `gorm:"not null;index:idx_measurement_sensor_recorded,priority:2,sort:desc"`
}

func (measurementRow) TableName() string { return "measurement" }

func (r measurementRow) toDomain() domain.Measurement {
	return domain.Measurement{ID: r.ID, SensorID: r.SensorID, Level: r.Level, RecordedAt: r.RecordedAt}
}

func toMeasurements(rows []measurementRow) []domain.Measurement {
	out := make([]domain.Measurement, len(rows))
	for i, row := range rows {
		out[i] = row.toDomain()
	}
	return out
}

var _ domain.MeasurementRepository = (*MeasurementRepository)(nil)

type MeasurementRepository struct{ db *gorm.DB }

func NewMeasurementRepository(db *gorm.DB) *MeasurementRepository {
	return &MeasurementRepository{db: db}
}

func (m *MeasurementRepository) Add(ctx context.Context, measurement *domain.Measurement) error {
	row := measurementRow{
		SensorID:   measurement.SensorID,
		Level:      measurement.Level,
		RecordedAt: measurement.RecordedAt,
	}
	if err := conn(ctx, m.db).Omit("Sensor").Create(&row).Error; err != nil {
		return err
	}
	measurement.ID = row.ID
	return nil
}

// Recent returns the newest measurements first, the order the domain evaluates
// streaks in.
func (m *MeasurementRepository) Recent(ctx context.Context, sensorID uint, limit int) ([]domain.Measurement, error) {
	var rows []measurementRow
	err := conn(ctx, m.db).
		Where("sensor_id = ?", sensorID).
		Order("recorded_at DESC, id DESC").
		Limit(limit).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toMeasurements(rows), nil
}

func (m *MeasurementRepository) BySensor(ctx context.Context, sensorID uint) ([]domain.Measurement, error) {
	var rows []measurementRow
	err := conn(ctx, m.db).
		Where("sensor_id = ?", sensorID).
		Order("recorded_at DESC, id DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	return toMeasurements(rows), nil
}
