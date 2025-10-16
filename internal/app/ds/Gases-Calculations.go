package ds

type GasCalculation struct {
	ID uint `gorm:"primaryKey"`
	// здесь создаем Unique key, указывая общий uniqueIndex
	CalculationID uint `gorm:"not null;uniqueIndex:idx_calculation_gas"`
	GasID         uint `gorm:"not null;uniqueIndex:idx_calculation_gas"`

	Sound    bool `gorm:"default:true"`
	Quantity int  `gorm:"default:1"`
	Position int  `gorm:"default:0"`

	Calculation Calculation `gorm:"foreignKey:CalculationID"`
	Gas         Gas         `gorm:"foreignKey:GasID"`
}
