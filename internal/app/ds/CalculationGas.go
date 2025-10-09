package ds

type CalculationGas struct {
	ID            uint `gorm:"primaryKey"`
	CalculationID uint `gorm:"not null;uniqueIndex:idx_calculation_gas"`
	GasID         uint `gorm:"not null;uniqueIndex:idx_calculation_gas"`

	// Дополнительные поля связи если нужны
	Comment string `gorm:"type:text"`

	Calculation Calculation `gorm:"foreignKey:CalculationID"`
	Gas         Gas         `gorm:"foreignKey:GasID"`
}
