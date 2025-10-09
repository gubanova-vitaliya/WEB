package repository

import (
	"WEB/internal/app/ds"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func (r *Repository) GetAllGases() ([]ds.Gas, error) {
	var gases []ds.Gas
	err := r.db.Where("is_deleted = ?", false).Find(&gases).Error
	return gases, err
}

func (r *Repository) GetGasByID(id int) (*ds.Gas, error) {
	var gas ds.Gas
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&gas).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &gas, err
}

func (r *Repository) SearchGases(query string) ([]ds.Gas, error) {
	var gases []ds.Gas
	err := r.db.Where("(title ILIKE ? OR formula ILIKE ? OR description ILIKE ?) AND is_deleted = ?",
		"%"+query+"%", "%"+query+"%", "%"+query+"%", false).Find(&gases).Error
	return gases, err
}

// CreateCalculationWithGas - создание расчета с газом через курсор
func (r *Repository) CreateCalculationWithGas(gasID uint, userID uint) (*ds.Calculation, error) {
	// Используем транзакцию для атомарности операций
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Создаем расчет
	calculation := &ds.Calculation{
		UserID:       userID,
		GasID:        gasID,
		Volume:       0,
		Temperature1: 0,
		Temperature2: 0,
		Pressure1:    0,
		Pressure2:    0,
		Mass:         0,
		Moles:        0,
		DateCreate:   time.Now(),
		DateUpdate:   time.Now(),
		IsDeleted:    false,
	}

	// Используем курсор для вставки
	err := tx.Create(calculation).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Создаем связь в CalculationGas через курсор
	calculationGas := &ds.CalculationGas{
		CalculationID: calculation.ID,
		GasID:         gasID,
		Comment:       "Добавлен в журнал",
	}

	err = tx.Create(calculationGas).Error
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	// Коммитим транзакцию
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return calculation, nil
}

// GetAllCalculationsWithGases - получение всех расчетов с газами через курсор
func (r *Repository) GetAllCalculationsWithGases() ([]ds.Calculation, error) {
	var calculations []ds.Calculation

	// Используем курсор для выборки данных
	err := r.db.Preload("Gas").
		Where("is_deleted = ?", false).
		Order("date_create DESC").
		Find(&calculations).Error

	return calculations, err
}

// ClearAllCalculations - очистка всех расчетов через курсор
func (r *Repository) ClearAllCalculations(userID uint) error {
	// Используем курсор для обновления записей
	return r.db.Model(&ds.Calculation{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Update("is_deleted", true).Error
}

// GetJournalCount - получение количества записей через курсор
func (r *Repository) GetJournalCount(userID uint) int64 {
	var count int64

	// Используем курсор для подсчета
	err := r.db.Model(&ds.Calculation{}).
		Where("user_id = ? AND is_deleted = ?", userID, false).
		Count(&count).Error

	if err != nil {
		logrus.Error("Error counting records in calculations:", err)
		return 0
	}
	return count
}

// GetJournalCountDefault - получение количества записей по умолчанию
func (r *Repository) GetJournalCountDefault() int64 {
	return r.GetJournalCount(1)
}

// GetCalculationsWithCursor - получение расчетов с использованием курсора для больших данных
func (r *Repository) GetCalculationsWithCursor(batchSize int) ([]ds.Calculation, error) {
	var calculations []ds.Calculation

	// Используем курсор для пагинации (если данных много)
	err := r.db.Preload("Gas").
		Where("is_deleted = ?", false).
		Order("date_create DESC").
		Limit(batchSize).
		Find(&calculations).Error

	return calculations, err
}

// AddGasToJournalWithCursor - добавление газа в журнал с использованием курсора
func (r *Repository) AddGasToJournalWithCursor(gasID uint, userID uint) error {
	// Начинаем транзакцию
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Используем курсор для проверки существования газа
	var gas ds.Gas
	if err := tx.Where("id = ? AND is_deleted = ?", gasID, false).First(&gas).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Создаем расчет через курсор
	calculation := &ds.Calculation{
		UserID:       userID,
		GasID:        gasID,
		Volume:       0.0,
		Temperature1: 0.0,
		Temperature2: 0.0,
		Pressure1:    0.0,
		Pressure2:    0.0,
		Mass:         0.0,
		Moles:        0.0,
		DateCreate:   time.Now(),
		DateUpdate:   time.Now(),
		IsDeleted:    false,
	}

	if err := tx.Create(calculation).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Создаем связь через курсор
	calculationGas := &ds.CalculationGas{
		CalculationID: calculation.ID,
		GasID:         gasID,
		Comment:       "Добавлен через курсор",
	}

	if err := tx.Create(calculationGas).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Коммитим транзакцию
	return tx.Commit().Error
}
