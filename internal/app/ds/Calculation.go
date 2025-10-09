package ds

import (
	"time"
)

type Calculation struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null"`
	GasID        uint      `gorm:"not null"`
	Volume       float64   `gorm:"type:decimal(10,4);not null"` // Объем сосуда (л)
	Temperature1 float64   `gorm:"type:decimal(10,2);not null"` // Начальная температура (°C)
	Temperature2 float64   `gorm:"type:decimal(10,2);not null"` // Конечная температура (°C)
	Pressure1    float64   `gorm:"type:decimal(10,2);not null"` // Начальное давление (атм)
	Pressure2    float64   `gorm:"type:decimal(10,2)"`          // Рассчитанное конечное давление (атм)
	Mass         float64   `gorm:"type:decimal(10,4)"`          // Масса газа (г)
	Moles        float64   `gorm:"type:decimal(10,4)"`          // Количество вещества (моль)
	DateCreate   time.Time `gorm:"not null"`
	DateUpdate   time.Time
	IsDeleted    bool `gorm:"type:boolean;default:false"`

	User Users `gorm:"foreignKey:UserID"`
	Gas  Gas   `gorm:"foreignKey:GasID"`
}
