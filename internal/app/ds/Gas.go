package ds

type Gas struct {
	ID          uint    `gorm:"primaryKey"`
	Title       string  `gorm:"type:varchar(100);not null"`
	Formula     string  `gorm:"type:varchar(50);not null"`
	MolarMass   float64 `gorm:"type:decimal(8,4);not null"`
	ImageURL    string  `gorm:"type:varchar(200)"`
	Description string  `gorm:"type:text"`
	IsDeleted   bool    `gorm:"type:boolean;default:false"`
}
