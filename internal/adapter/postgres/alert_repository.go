package postgres

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/GoranDimitrovski/go-alert-rest-api/internal/domain"
)

type alertRow struct {
	ID        uint      `gorm:"primaryKey"`
	SensorID  uint      `gorm:"not null;index"`
	Sensor    sensorRow `gorm:"foreignKey:SensorID;constraint:OnDelete:CASCADE"`
	Level1    uint      `gorm:"not null"`
	Level2    uint      `gorm:"not null"`
	Level3    uint      `gorm:"not null"`
	StartTime time.Time `gorm:"not null"`
	EndTime   time.Time `gorm:"not null"`
}

func (alertRow) TableName() string { return "alert" }

func (r alertRow) toDomain() domain.Alert {
	return domain.Alert{
		ID:        r.ID,
		SensorID:  r.SensorID,
		Levels:    [domain.AlertStreak]uint{r.Level1, r.Level2, r.Level3},
		StartTime: r.StartTime,
		EndTime:   r.EndTime,
	}
}

var _ domain.AlertRepository = (*AlertRepository)(nil)

type AlertRepository struct{ db *gorm.DB }

func NewAlertRepository(db *gorm.DB) *AlertRepository { return &AlertRepository{db: db} }

func (a *AlertRepository) Add(ctx context.Context, alert *domain.Alert) error {
	row := alertRow{
		SensorID:  alert.SensorID,
		Level1:    alert.Levels[0],
		Level2:    alert.Levels[1],
		Level3:    alert.Levels[2],
		StartTime: alert.StartTime,
		EndTime:   alert.EndTime,
	}
	if err := conn(ctx, a.db).Omit("Sensor").Create(&row).Error; err != nil {
		return err
	}
	alert.ID = row.ID
	return nil
}

func (a *AlertRepository) BySensor(ctx context.Context, sensorID uint) ([]domain.Alert, error) {
	var rows []alertRow
	err := conn(ctx, a.db).
		Where("sensor_id = ?", sensorID).
		Order("start_time DESC, id DESC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	alerts := make([]domain.Alert, len(rows))
	for i, row := range rows {
		alerts[i] = row.toDomain()
	}
	return alerts, nil
}
