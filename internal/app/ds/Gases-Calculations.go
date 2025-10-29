package ds

type GasCalculation struct {
	ID uint `gorm:"primaryKey"`

	CalculationID uint `gorm:"not null;uniqueIndex:idx_calculation_gas"`
	GasID         uint `gorm:"not null;uniqueIndex:idx_calculation_gas"`

	Sound    bool `gorm:"default:true"`
	Quantity int  `gorm:"default:1"`
	Position int  `gorm:"default:0"`

	// Параметры расчета для конкретного газа
	InitialPressure    float64 `gorm:"type:decimal(10,4);default:0"`
	InitialTemperature float64 `gorm:"type:decimal(10,4);default:0"`
	FinalTemperature   float64 `gorm:"type:decimal(10,4);default:0"`
	Volume             float64 `gorm:"type:decimal(10,4);default:0"`
	GasAmount          float64 `gorm:"type:decimal(10,4);default:0"`
	FinalPressure      float64 `gorm:"type:decimal(10,4);default:0"`

	Calculation Calculation `gorm:"foreignKey:CalculationID"`
	Gas         Gas         `gorm:"foreignKey:GasID"`
}
