package repository

import (
	"WEB/internal/app/ds"
	"sync"
)

// InMemoryStorage для временного хранения расчетов
type InMemoryStorage struct {
	mu           sync.RWMutex
	calculations map[uint][]*GasCalculationData // creatorID -> список газов
}

type GasCalculationData struct {
	GasCalculationID   uint
	GasID              uint
	GasTitle           string
	GasFormula         string
	GasMolarMass       float64
	GasImageURL        string
	GasDescription     string
	InitialPressure    float64
	InitialTemperature float64
	FinalTemperature   float64
	Volume             float64
	GasAmount          float64
	FinalPressure      float64
}

var (
	storage *InMemoryStorage
	once    sync.Once
)

func getStorage() *InMemoryStorage {
	once.Do(func() {
		storage = &InMemoryStorage{
			calculations: make(map[uint][]*GasCalculationData),
		}
	})
	return storage
}

// AddGasToCalculation добавляет газ в расчет (в память) - БЕЗ ПРОВЕРКИ ДУБЛИКАТОВ
func (r *Repository) AddGasToCalculation(creatorID uint, gas *ds.Gas) error {
	storage := getStorage()
	storage.mu.Lock()
	defer storage.mu.Unlock()

	// Создаем новую запись (теперь можно добавлять один и тот же газ много раз)
	gasCalc := &GasCalculationData{
		GasCalculationID: uint(len(storage.calculations[creatorID]) + 1),
		GasID:            uint(gas.ID),
		GasTitle:         gas.Title,
		GasFormula:       gas.Formula,
		GasMolarMass:     gas.MolarMass,
		GasImageURL:      gas.ImageURL,
		GasDescription:   gas.Description,
	}

	storage.calculations[creatorID] = append(storage.calculations[creatorID], gasCalc)
	return nil
}

// GetGasesInCalculation возвращает все газы в расчете
func (r *Repository) GetGasesInCalculation(creatorID uint) ([]map[string]interface{}, error) {
	storage := getStorage()
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	gases := storage.calculations[creatorID]
	results := make([]map[string]interface{}, len(gases))

	for i, gas := range gases {
		results[i] = map[string]interface{}{
			"gas_calculation_id":  gas.GasCalculationID,
			"gas_id":              gas.GasID,
			"gas_title":           gas.GasTitle,
			"gas_formula":         gas.GasFormula,
			"gas_molar_mass":      gas.GasMolarMass,
			"gas_image_url":       gas.GasImageURL,
			"gas_description":     gas.GasDescription,
			"initial_pressure":    gas.InitialPressure,
			"initial_temperature": gas.InitialTemperature,
			"final_temperature":   gas.FinalTemperature,
			"volume":              gas.Volume,
			"gas_amount":          gas.GasAmount,
			"final_pressure":      gas.FinalPressure,
		}
	}

	return results, nil
}

// RemoveGasFromCalculation удаляет газ из расчета
func (r *Repository) RemoveGasFromCalculation(creatorID uint, gasCalculationID uint) error {
	storage := getStorage()
	storage.mu.Lock()
	defer storage.mu.Unlock()

	gases := storage.calculations[creatorID]
	for i, gas := range gases {
		if gas.GasCalculationID == gasCalculationID {
			// Удаляем газ из slice
			storage.calculations[creatorID] = append(gases[:i], gases[i+1:]...)
			return nil
		}
	}

	return nil
}

// UpdateCalculationParams обновляет параметры расчета
func (r *Repository) UpdateCalculationParams(creatorID uint, gasCalculationID uint, params map[string]interface{}) error {
	storage := getStorage()
	storage.mu.Lock()
	defer storage.mu.Unlock()

	gases := storage.calculations[creatorID]
	for _, gas := range gases {
		if gas.GasCalculationID == gasCalculationID {
			// Обновляем параметры
			if pressure, ok := params["initial_pressure"].(float64); ok {
				gas.InitialPressure = pressure
			}
			if temp, ok := params["initial_temperature"].(float64); ok {
				gas.InitialTemperature = temp
			}
			if temp, ok := params["final_temperature"].(float64); ok {
				gas.FinalTemperature = temp
			}
			if volume, ok := params["volume"].(float64); ok {
				gas.Volume = volume
			}
			if amount, ok := params["gas_amount"].(float64); ok {
				gas.GasAmount = amount
			}
			if pressure, ok := params["final_pressure"].(float64); ok {
				gas.FinalPressure = pressure
			}
			break
		}
	}

	return nil
}

// GetCartCount для получения количества газов в расчете
func (r *Repository) GetCartCount() int64 {
	storage := getStorage()
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	creatorID := uint(1)
	return int64(len(storage.calculations[creatorID]))
}

// ClearAllCalculations очищает все расчеты (для кнопки "Очистить все")
func (r *Repository) ClearAllCalculations(creatorID uint) error {
	storage := getStorage()
	storage.mu.Lock()
	defer storage.mu.Unlock()

	storage.calculations[creatorID] = []*GasCalculationData{}
	return nil
}
