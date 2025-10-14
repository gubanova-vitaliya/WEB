package ds

type Gas struct {
	ID          int     `gorm:"primaryKey"`
	Title       string  `gorm:"type:varchar(255)"`
	Formula     string  `gorm:"type:varchar(50)"`
	MolarMass   float64 `gorm:"type:decimal(10,4)"`
	ImageURL    string  `gorm:"type:varchar(255)"`
	Description string  `gorm:"type:text"`
}

func (Gas) TableName() string {
	return "gas"
}
