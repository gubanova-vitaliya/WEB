package repository

import (
	"fmt"
	"strings"
	"sync"
)

type Repository struct {
	mu sync.RWMutex
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Gas struct {
	ID          int
	Title       string
	Formula     string
	MolarMass   float64
	ImageURL    string
	Description string
}

// Структура для расчета давления
type PressureCalculation struct {
	ID                 int
	GasID              int
	GasTitle           string
	Formula            string
	InitialPressure    float64 // Па
	InitialTemperature float64 // K
	FinalTemperature   float64 // K
	FinalPressure      float64 // Па
}

// Журнал расчетов
type Journal struct {
	Calculations []PressureCalculation
}

var (
	journalStorage = &Journal{
		Calculations: []PressureCalculation{
			{
				ID:                 1,
				GasID:              1,
				GasTitle:           "Азот",
				Formula:            "N₂",
				InitialPressure:    101325,
				InitialTemperature: 293,
				FinalTemperature:   373,
				FinalPressure:      128857,
			},
			{
				ID:                 2,
				GasID:              2,
				GasTitle:           "Кислород",
				Formula:            "O₂",
				InitialPressure:    100000,
				InitialTemperature: 273,
				FinalTemperature:   323,
				FinalPressure:      118315,
			},
			{
				ID:                 3,
				GasID:              3,
				GasTitle:           "Гелий",
				Formula:            "He",
				InitialPressure:    95000,
				InitialTemperature: 283,
				FinalTemperature:   353,
				FinalPressure:      118551,
			},
		},
	}
	journalMutex = &sync.RWMutex{}
)

func (r *Repository) GetGases() ([]Gas, error) {
	Gases := []Gas{
		{
			ID:          1,
			Title:       "Азот",
			Formula:     "N₂",
			MolarMass:   28.02,
			ImageURL:    "http://127.0.0.1:9000/gase/azot.webp",
			Description: "Азот — инертный газ, составляющий около 78% атмосферы Земли. Широко используется в промышленности и медицине.",
		},
		{
			ID:          2,
			Title:       "Кислород",
			Formula:     "O₂",
			MolarMass:   32.00,
			ImageURL:    "http://localhost:9000/gase/kislorod.webp",
			Description: "Кислород необходим для дыхания и горения. Составляет около 21% атмосферы Земли.",
		},
		{
			ID:          3,
			Title:       "Гелий",
			Formula:     "He",
			MolarMass:   4.00,
			ImageURL:    "http://localhost:9000/gase/geliy.png",
			Description: "Гелий — лёгкий инертный газ, второй по распространённости во Вселенной. Используется в баллонах и охлаждающих системах.",
		},
		{
			ID:          4,
			Title:       "Водород",
			Formula:     "H₂",
			MolarMass:   2.016,
			ImageURL:    "http://localhost:9000/gase/vodolod.webp",
			Description: "Самый легкий газ во Вселенной с высокой диффузионной способностью. Используется как топливо и в химической промышленности.",
		},
		{
			ID:          5,
			Title:       "Углекислый газ",
			Formula:     "CO₂",
			MolarMass:   44.01,
			ImageURL:    "http://localhost:9000/gase/uglekisliy_gas.webp",
			Description: "Важный компонент атмосферы и круговорота углерода. Широко применяется в пищевой промышленности и пожаротушении.",
		},
	}
	if len(Gases) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}
	return Gases, nil
}

func (r *Repository) GetGas(id int) (Gas, error) {
	gases, err := r.GetGases()
	if err != nil {
		return Gas{}, err
	}

	for _, gas := range gases {
		if gas.ID == id {
			return gas, nil
		}
	}
	return Gas{}, fmt.Errorf("газ не найден")
}

func (r *Repository) GetGasesByTitle(title string) ([]Gas, error) {
	gases, err := r.GetGases()
	if err != nil {
		return []Gas{}, err
	}

	var result []Gas
	for _, gase := range gases {
		if strings.Contains(strings.ToLower(gase.Title), strings.ToLower(title)) {
			result = append(result, gase)
		}
	}

	return result, nil
}

// Методы для работы с журналом расчетов
func (r *Repository) GetJournal() *Journal {
	journalMutex.RLock()
	defer journalMutex.RUnlock()
	return journalStorage
}

func (r *Repository) GetJournalCount() int {
	journalMutex.RLock()
	defer journalMutex.RUnlock()
	return len(journalStorage.Calculations)
}
