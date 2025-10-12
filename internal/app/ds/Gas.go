package ds

type Gas struct {
	ID          int
	Title       string
	Formula     string
	MolarMass   float64
	ImageURL    string
	Description string
}

func (Gas) TableName() string {
	return "gas"
}
