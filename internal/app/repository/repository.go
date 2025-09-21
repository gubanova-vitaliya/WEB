package repository

import (
	"fmt"
	"strings"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Gas struct {
	ID          int
	Title       string  //Название газа
	Formula     string  //Химическая формула
	MolarMass   float64 //Молярная масса, г/моль
	ImageURL    string  //Ссылка на картинку из Minio
	Description string  //Описание газа
}

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
		return Gas{}, err // тут у нас уже есть кастомная ошибка из нашего метода, поэтому мы можем просто вернуть ее
	}

	for _, gas := range gases {
		if gas.ID == id {
			return gas, nil // если нашли, то просто возвращаем найденный заказ (услугу) без ошибок
		}
	}
	return Gas{}, fmt.Errorf("заказ не найден")
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
