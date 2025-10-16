package ds

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Calculation struct {
	ID          uint           `gorm:"primaryKey"`
	Status      string         `gorm:"type:varchar(15);not null"`
	Text        sql.NullString `gorm:"type:text;default:null"`
	DateCreate  time.Time      `gorm:"not null"`
	DateUpdate  time.Time
	DateForm    sql.NullTime `gorm:"default:null"`
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID uint

	// Поля для расчета давления в сосуде при нагреве идеального газа
	InitialPressure    sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"` // Начальное давление, Па
	InitialTemperature sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"` // Начальная температура, К
	FinalTemperature   sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"` // Конечная температура, К
	Volume             sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"` // Объем сосуда, м³
	GasAmount          sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"` // Количество вещества, моль
	FinalPressure      sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"` // Расчетное конечное давление, Па

	Creator   Users `gorm:"foreignKey:CreatorID"`
	Moderator Users `gorm:"foreignKey:ModeratorID"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
}
