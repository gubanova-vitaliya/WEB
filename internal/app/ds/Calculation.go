package ds

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

// internal/ds/Calculation.go
type Calculation struct {
	ID          uint           `gorm:"primaryKey"`
	Status      string         `gorm:"type:varchar(15);not null"`
	Text        sql.NullString `gorm:"type:text;default:null"`
	DateCreate  time.Time      `gorm:"not null"`
	DateUpdate  time.Time
	DateForm    sql.NullTime `gorm:"default:null"`
	DateFinish  sql.NullTime `gorm:"default:null"`
	CreatorID   uint         `gorm:"not null"`
	ModeratorID *uint        `gorm:"default:null"` // Измените на указатель

	// Остальные поля...
	InitialPressure    sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	InitialTemperature sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	FinalTemperature   sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	Volume             sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	GasAmount          sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	FinalPressure      sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`

	Creator   Users  `gorm:"foreignKey:CreatorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"` // Также измените здесь

	DeletedAt gorm.DeletedAt `gorm:"index"`
}
