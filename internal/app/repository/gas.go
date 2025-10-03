package repository

import (
	"WEB/internal/app/ds"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) GetAllCalculations() ([]ds.Calculation, error) {
	var calculations []ds.Calculation
	err := r.db.Preload("Gas").Preload("User").Find(&calculations).Error
	if err != nil {
		return nil, err
	}
	return calculations, nil
}

func (r *Repository) GetCalculationByID(id int) (*ds.Calculation, error) {
	var calculation ds.Calculation
	err := r.db.Preload("Gas").Preload("User").
		Where("id = ? AND is_deleted = ?", id, false).
		First(&calculation).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &calculation, nil
}

func (r *Repository) SearchCalculationsByName(name string) ([]ds.Calculation, error) {
	var calculations []ds.Calculation

	// Если у Calculation нет title/description, ищите по другим полям
	// Например, по связанному газу:
	err := r.db.Preload("Gas", "title ILIKE ?", "%"+name+"%").
		Preload("User").
		Where("is_deleted = ?", false).
		Find(&calculations).Error

	if err != nil {
		return nil, err
	}
	return calculations, nil
}

func (r *Repository) CreateCalculation(calculation *ds.Calculation) error {
	return r.db.Create(calculation).Error
}

func (r *Repository) DeleteCalculation(id int) error {
	return r.db.Delete(&ds.Calculation{}, id).Error
}

func (r *Repository) ClearAllCalculations() error {
	// Сначала очищаем связанную таблицу calculation_gas
	err := r.db.Where("1 = 1").Delete(&ds.CalculationGas{}).Error
	if err != nil {
		return err
	}

	// Затем очищаем таблицу calculations
	return r.db.Where("1 = 1").Delete(&ds.Calculation{}).Error
}

func (r *Repository) GetAllGases() ([]ds.Gas, error) {
	var gases []ds.Gas
	err := r.db.Where("is_deleted = ?", false).Find(&gases).Error
	if err != nil {
		return nil, err
	}
	return gases, nil
}

func (r *Repository) GetGasByID(id int) (*ds.Gas, error) {
	var gas ds.Gas
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&gas).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &gas, nil
}

func (r *Repository) SearchGases(query string) ([]ds.Gas, error) {
	var gases []ds.Gas
	err := r.db.Where("(title ILIKE ? OR formula ILIKE ? OR description ILIKE ?) AND is_deleted = ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", false).Find(&gases).Error
	if err != nil {
		return nil, err
	}
	return gases, nil
}

// GetJournalCount для получения количества расчетов в журнале
func (r *Repository) GetJournalCount(userID uint) int64 {
	var count int64
	err := r.db.Model(&ds.Calculation{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error

	if err != nil {
		logrus.Error("Error counting records in calculations:", err)
		return 0
	}
	return count
}

// Временная функция для обратной совместимости
func (r *Repository) GetJournalCountDefault() int64 {
	return r.GetJournalCount(1) // временное решение
}
