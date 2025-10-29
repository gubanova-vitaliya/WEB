package ds

import (
	"database/sql"
	"time"

	"gorm.io/gorm"
)

type Calculation struct {
	ID     uint           `gorm:"primaryKey"`
	Status string         `gorm:"type:varchar(15);not null;default:'draft'"`
	Text   sql.NullString `gorm:"type:text;default:null"`

	// Даты
	DateCreate time.Time    `gorm:"not null"`
	DateForm   sql.NullTime `gorm:"default:null"`
	DateFinish sql.NullTime `gorm:"default:null"`

	// Пользователи
	CreatorID   uint  `gorm:"not null"`
	ModeratorID *uint `gorm:"default:null"`

	// Параметры расчета (теперь на уровне расчета, а не газа)
	InitialPressure    sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	InitialTemperature sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	FinalTemperature   sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	Volume             sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	GasAmount          sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`
	FinalPressure      sql.NullFloat64 `gorm:"type:decimal(10,4);default:null"`

	// Связи
	Creator   Users  `gorm:"foreignKey:CreatorID"`
	Moderator *Users `gorm:"foreignKey:ModeratorID"`

	// Газы в расчете
	Gases []GasCalculation `gorm:"foreignKey:CalculationID"`

	DeletedAt gorm.DeletedAt `gorm:"index"`
}
