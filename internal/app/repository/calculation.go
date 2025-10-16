// repository/calculation.go
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
	IsActive           bool // Флаг для логического удаления
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

// AddGasToCalculation добавляет газ в расчет (в память)
func (r *Repository) AddGasToCalculation(creatorID uint, gas *ds.Gas) error {
	storage := getStorage()
	storage.mu.Lock()
	defer storage.mu.Unlock()

	// Создаем новую запись
	gasCalc := &GasCalculationData{
		GasCalculationID: uint(len(storage.calculations[creatorID]) + 1),
		GasID:            uint(gas.ID),
		GasTitle:         gas.Title,
		GasFormula:       gas.Formula,
		GasMolarMass:     gas.MolarMass,
		GasImageURL:      gas.ImageURL,
		GasDescription:   gas.Description,
		IsActive:         true, // По умолчанию активен
	}

	storage.calculations[creatorID] = append(storage.calculations[creatorID], gasCalc)
	return nil
}

// GetGasesInCalculation возвращает только активные газы в расчете
func (r *Repository) GetGasesInCalculation(creatorID uint) ([]map[string]interface{}, error) {
	storage := getStorage()
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	allGases := storage.calculations[creatorID]
	var activeGases []*GasCalculationData

	// Фильтруем только активные
	for _, gas := range allGases {
		if gas.IsActive {
			activeGases = append(activeGases, gas)
		}
	}

	results := make([]map[string]interface{}, len(activeGases))
	for i, gas := range activeGases {
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
			"is_active":           gas.IsActive,
		}
	}

	return results, nil
}

// RemoveGasFromCalculation ЛОГИЧЕСКИ удаляет газ из расчета
func (r *Repository) RemoveGasFromCalculation(creatorID uint, gasCalculationID uint) error {
	storage := getStorage()
	storage.mu.Lock()
	defer storage.mu.Unlock()

	gases := storage.calculations[creatorID]
	for _, gas := range gases {
		if gas.GasCalculationID == gasCalculationID {
			// Логическое удаление - просто помечаем как неактивный
			gas.IsActive = false
			break
		}
	}

	return nil
}

// GetCartCount для получения количества АКТИВНЫХ газов в расчете
func (r *Repository) GetCartCount() int64 {
	storage := getStorage()
	storage.mu.RLock()
	defer storage.mu.RUnlock()

	creatorID := uint(1)
	gases := storage.calculations[creatorID]

	count := 0
	for _, gas := range gases {
		if gas.IsActive {
			count++
		}
	}

	return int64(count)
}
