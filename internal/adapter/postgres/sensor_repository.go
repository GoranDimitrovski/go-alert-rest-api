package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

type sensorRow struct {
	ID        uint   `gorm:"primaryKey;autoIncrement:false"`
	Status    string `gorm:"type:varchar(16);not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (sensorRow) TableName() string { return "sensor" }

func (r sensorRow) toDomain() domain.Sensor {
	return domain.Sensor{ID: r.ID, Status: domain.Status(r.Status)}
}

var _ domain.SensorRepository = (*SensorRepository)(nil)

type SensorRepository struct{ db *gorm.DB }

func NewSensorRepository(db *gorm.DB) *SensorRepository { return &SensorRepository{db: db} }

func (s *SensorRepository) ByID(ctx context.Context, id uint) (domain.Sensor, error) {
	var row sensorRow
	if err := conn(ctx, s.db).First(&row, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Sensor{}, domain.ErrNotFound
		}
		return domain.Sensor{}, err
	}
	return row.toDomain(), nil
}

// Save inserts the sensor, or updates the status of one already registered.
func (s *SensorRepository) Save(ctx context.Context, sensor domain.Sensor) error {
	row := sensorRow{ID: sensor.ID, Status: string(sensor.Status)}
	return conn(ctx, s.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"status", "updated_at"}),
	}).Create(&row).Error
}
