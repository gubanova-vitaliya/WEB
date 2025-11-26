package repository

import (
	"WEB/internal/app/ds"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// CalculateGasPressure рассчитывает давление для конкретного газа в расчете
func (r *Repository) CalculateGasPressure(gasCalculationID uint, params map[string]float64) (float64, error) {
	var gasCalc ds.GasCalculation
	if err := r.db.Preload("Gas").First(&gasCalc, gasCalculationID).Error; err != nil {
		return 0, err
	}

	// Получаем значения из параметров или из сохраненных данных
	gasAmount := getParamValue(params, "gas_amount", gasCalc.GasAmount)
	finalTemp := getParamValue(params, "final_temperature", gasCalc.FinalTemperature)
	volume := getParamValue(params, "volume", gasCalc.Volume)

	// Проверяем, что все обязательные параметры есть
	if gasAmount == 0 || finalTemp == 0 || volume == 0 {
		return 0, errors.New("missing required parameters: gas_amount, final_temperature, volume")
	}

	// Универсальная газовая постоянная (Дж/(моль·К))
	const R = 8.314462618

	// Расчет давления по уравнению Менделеева-Клапейрона: P = nRT/V
	// P в Паскалях, n в молях, T в Кельвинах, V в м³
	pressure := (gasAmount * R * finalTemp) / volume

	// Конвертируем в атмосферы (1 атм = 101325 Па)
	pressureAtm := pressure / 101325.0

	// Подготавливаем обновления
	updates := map[string]interface{}{
		"gas_amount":        gasAmount,
		"final_temperature": finalTemp,
		"volume":            volume,
		"final_pressure":    pressureAtm, // Сохраняем в атмосферах
	}

	// Добавляем опциональные параметры
	if initialPressure, ok := params["initial_pressure"]; ok {
		updates["initial_pressure"] = initialPressure
	}
	if initialTemp, ok := params["initial_temperature"]; ok {
		updates["initial_temperature"] = initialTemp
	}

	// Сохраняем в базу
	if err := r.db.Model(&ds.GasCalculation{}).Where("id = ?", gasCalculationID).Updates(updates).Error; err != nil {
		return 0, err
	}

	return pressureAtm, nil
}

// Вспомогательная функция для получения значения параметра
func getParamValue(params map[string]float64, key string, dbValue sql.NullFloat64) float64 {
	if val, ok := params[key]; ok {
		return val
	}
	if dbValue.Valid {
		return dbValue.Float64
	}
	return 0
}

// UpdateGasCalculationParams обновляет параметры расчета для газа
func (r *Repository) UpdateGasCalculationParams(gasCalculationID uint, params map[string]interface{}) error {
	return r.db.Model(&ds.GasCalculation{}).Where("id = ?", gasCalculationID).Updates(params).Error
}

// GetCalculationWithGases возвращает расчет с газами и их параметрами
func (r *Repository) GetCalculationWithGases(calculationID uint) (*ds.Calculation, error) {
	var calculation ds.Calculation
	err := r.db.
		Preload("Gases").
		Preload("Gases.Gas").
		Preload("Creator").
		First(&calculation, calculationID).Error
	if err != nil {
		return nil, err
	}
	return &calculation, nil
}

// GetDraftCalculation возвращает черновик расчета пользователя
func (r *Repository) GetDraftCalculation(creatorID uint) (*ds.Calculation, error) {
	var calculation ds.Calculation
	err := r.db.
		Preload("Gases").
		Preload("Gases.Gas").
		Where("creator_id = ? AND status = ?", creatorID, "draft").
		First(&calculation).Error

	if err != nil {
		// Создаем новый черновик, если не найден
		return r.ensureDraftCalculation(creatorID)
	}

	return &calculation, nil
}

// CalculateAllGases рассчитывает все газы в расчете
func (r *Repository) CalculateAllGases(calculationID uint) (map[uint]float64, error) {
	var gasCalcs []ds.GasCalculation
	if err := r.db.Where("calculation_id = ?", calculationID).Find(&gasCalcs).Error; err != nil {
		return nil, err
	}

	results := make(map[uint]float64)
	const R = 8.314462618

	for _, gasCalc := range gasCalcs {
		// Проверяем, что все необходимые параметры заполнены
		if gasCalc.GasAmount.Valid && gasCalc.FinalTemperature.Valid && gasCalc.Volume.Valid &&
			gasCalc.GasAmount.Float64 > 0 && gasCalc.FinalTemperature.Float64 > 0 && gasCalc.Volume.Float64 > 0 {

			// Расчет давления в Паскалях
			pressurePa := (gasCalc.GasAmount.Float64 * R * gasCalc.FinalTemperature.Float64) / gasCalc.Volume.Float64

			// Конвертируем в атмосферы
			pressureAtm := pressurePa / 101325.0
			results[gasCalc.ID] = pressureAtm

			// Сохраняем результат
			r.db.Model(&ds.GasCalculation{}).Where("id = ?", gasCalc.ID).Update("final_pressure", pressureAtm)
		}
	}

	return results, nil
}

// ListCalculations - обновляем для работы с вычисляемым полем
func (r *Repository) ListCalculations(status string, dateFrom string, dateTo string) ([]map[string]interface{}, error) {
	// Создаем подзапрос для подсчета газов с рассчитанным давлением
	subQuery := r.db.Model(&ds.GasCalculation{}).
		Select("calculation_id, COUNT(*) as calculated_count").
		Where("final_pressure > 0").
		Group("calculation_id")

	q := r.db.Model(&ds.Calculation{}).
		Select("calculations.*, COALESCE(sq.calculated_count, 0) as calculated_count").
		Joins("LEFT JOIN (?) AS sq ON calculations.id = sq.calculation_id", subQuery).
		Where("calculations.status <> ?", "deleted").
		Where("calculations.status <> ?", "draft").
		Preload("Creator").
		Preload("Moderator")

	if status != "" {
		q = q.Where("calculations.status = ?", status)
	}
	if dateFrom != "" {
		q = q.Where("calculations.date_form >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("calculations.date_form <= ?", dateTo)
	}

	var results []struct {
		ds.Calculation
		CalculatedCount int `gorm:"column:calculated_count"`
	}

	if err := q.Order("calculations.date_create desc").Find(&results).Error; err != nil {
		return nil, err
	}

	out := make([]map[string]interface{}, 0, len(results))
	for _, result := range results {
		out = append(out, map[string]interface{}{
			"id":               result.ID,
			"status":           result.Status,
			"text":             result.Text,
			"date_create":      result.DateCreate,
			"date_form":        result.DateForm,
			"date_complete":    result.DateComplete,
			"creator_login":    result.Creator.Login,
			"moderator_login":  result.Moderator.Login,
			"calculated_count": result.CalculatedCount,
		})
	}
	return out, nil
}

// GetCalculationDetail возвращает детали расчета с газами
func (r *Repository) GetCalculationDetail(id uint) (*ds.Calculation, []map[string]interface{}, error) {
	var c ds.Calculation
	if err := r.db.Preload("Creator").Preload("Moderator").First(&c, id).Error; err != nil {
		return nil, nil, err
	}
	var mm []ds.GasCalculation
	if err := r.db.Preload("Gas").Where("calculation_id = ?", id).Find(&mm).Error; err != nil {
		return &c, nil, err
	}
	list := make([]map[string]interface{}, 0, len(mm))
	for _, m := range mm {
		list = append(list, map[string]interface{}{
			"gas_id":         m.GasID,
			"title":          m.Gas.Title,
			"formula":        m.Gas.Formula,
			"molar_mass":     m.Gas.MolarMass,
			"image_url":      m.Gas.ImageURL,
			"description":    m.Gas.Description,
			"sound":          m.Sound,
			"quantity":       m.Quantity,
			"position":       m.Position,
			"final_pressure": m.FinalPressure,
		})
	}
	return &c, list, nil
}

// UpdateCalculationFields обновляет поля расчета
func (r *Repository) UpdateCalculationFields(id uint, text *string) error {
	updates := map[string]interface{}{}
	if text != nil {
		updates["text"] = *text
	}
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(updates).Error
}

// SubmitCalculation с валидацией обязательных полей
func (r *Repository) SubmitCalculation(id uint, creatorID uint) error {
	var c ds.Calculation
	if err := r.db.Preload("Gases").First(&c, id).Error; err != nil {
		return err
	}
	if c.CreatorID != creatorID {
		return errors.New("only creator can submit")
	}
	if c.Status != "draft" {
		return errors.New("only draft can be submitted")
	}

	// Проверка обязательных полей
	if len(c.Gases) == 0 {
		return errors.New("calculation must contain at least one gas")
	}

	// Проверка что у всех газов заполнены обязательные параметры
	for _, gas := range c.Gases {
		if !(gas.GasAmount.Valid && gas.FinalTemperature.Valid && gas.Volume.Valid &&
			gas.GasAmount.Float64 > 0 && gas.FinalTemperature.Float64 > 0 && gas.Volume.Float64 > 0) {
			return errors.New("all gases must have gas_amount, final_temperature and volume filled")
		}
	}

	now := time.Now()
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":    "formed",
		"date_form": now, // Устанавливаем дату формирования
	}).Error
}

// CompleteCalculation завершает расчет с вычислением давления
func (r *Repository) CompleteCalculation(id uint, moderatorID uint) error {
	var c ds.Calculation
	if err := r.db.First(&c, id).Error; err != nil {
		return err
	}
	if c.Status != "formed" {
		return errors.New("only formed can be completed")
	}

	// Расчет конечного давления для всех газов
	_, err := r.CalculateAllGases(id)
	if err != nil {
		return err
	}

	now := time.Now()
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "completed",
		"moderator_id":  moderatorID,
		"date_complete": now, // Устанавливаем дату завершения
	}).Error
}

// RejectCalculation отклоняет расчет
func (r *Repository) RejectCalculation(id uint, moderatorID uint) error {
	var c ds.Calculation
	if err := r.db.First(&c, id).Error; err != nil {
		return err
	}
	if c.Status != "formed" {
		return errors.New("only formed can be rejected")
	}

	now := time.Now()
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":        "rejected",
		"moderator_id":  moderatorID,
		"date_complete": now,
	}).Error
}

// DeleteCalculation логически удаляет расчет
func (r *Repository) DeleteCalculation(id uint) error {
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Update("status", "deleted").Error
}

// AddGasToCalculation добавляет газ в расчет (черновик)
func (r *Repository) AddGasToCalculation(creatorID uint, gas *ds.Gas) error {
	return r.addGasToDraftDB(uint(gas.ID), creatorID)
}

// GetGasesInCalculation возвращает газы в расчете
func (r *Repository) GetGasesInCalculation(creatorID uint) ([]map[string]interface{}, error) {
	calc, err := r.ensureDraftCalculation(creatorID)
	if err != nil {
		return nil, err
	}

	var mm []ds.GasCalculation
	if err := r.db.Preload("Gas").Where("calculation_id = ?", calc.ID).Find(&mm).Error; err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, len(mm))
	for i, m := range mm {
		results[i] = map[string]interface{}{
			"gas_calculation_id": m.ID,
			"gas_id":             m.GasID,
			"gas_title":          m.Gas.Title,
			"gas_formula":        m.Gas.Formula,
			"gas_molar_mass":     m.Gas.MolarMass,
			"gas_image_url":      m.Gas.ImageURL,
			"gas_description":    m.Gas.Description,
		}
	}

	return results, nil
}

// RemoveGasFromCalculation удаляет газ из расчета
func (r *Repository) RemoveGasFromCalculation(creatorID uint, gasCalculationID uint) error {
	return r.db.Where("id = ?", gasCalculationID).Delete(&ds.GasCalculation{}).Error
}

// GetCartCount возвращает количество газов в корзине
func (r *Repository) GetCartCount() int64 {
	creatorID := r.FixedCreatorID()
	_, count, err := r.draftCartInfo(creatorID)
	if err != nil {
		return 0
	}
	return count
}

// ---------- DB-backed draft and m-m operations ----------

// ensureDraftCalculation returns existing draft or creates a new one for creator
func (r *Repository) ensureDraftCalculation(creatorID uint) (*ds.Calculation, error) {
	var calc ds.Calculation
	err := r.db.Where("creator_id = ? AND status = ?", creatorID, "draft").First(&calc).Error
	if err == nil {
		return &calc, nil
	}
	now := time.Now()
	calc = ds.Calculation{
		Status:     "draft",
		DateCreate: now,
		CreatorID:  creatorID,
	}
	if err := r.db.Create(&calc).Error; err != nil {
		return nil, err
	}
	return &calc, nil
}

// addGasToDraftDB creates m-m link if not exists
func (r *Repository) addGasToDraftDB(gasID uint, creatorID uint) error {
	logrus.Infof("addGasToDraftDB: gasID=%d, creatorID=%d", gasID, creatorID)

	calc, err := r.ensureDraftCalculation(creatorID)
	if err != nil {
		logrus.Errorf("ensureDraftCalculation error: %v", err)
		return err
	}
	logrus.Infof("Draft calculation ID: %d", calc.ID)

	// check gas exists
	var gas ds.Gas
	if err := r.db.First(&gas, gasID).Error; err != nil {
		logrus.Errorf("Gas not found: %v", err)
		return err
	}
	logrus.Infof("Gas found: %s", gas.Title)

	// Try to create the record, if it fails due to duplicate key, ignore the error
	mm := ds.GasCalculation{
		CalculationID: calc.ID,
		GasID:         gasID,
		Sound:         true,
		Quantity:      1,
		Position:      0,
	}

	err = r.db.Create(&mm).Error
	if err != nil {
		// Check if it's a duplicate key error
		if strings.Contains(err.Error(), "idx_calculation_gas") || strings.Contains(err.Error(), "duplicate key") {
			logrus.Infof("Gas already exists in calculation, ignoring duplicate key error")
			return nil // Ignore duplicate key error - gas is already in calculation
		}
		logrus.Errorf("Error creating gas calculation: %v", err)
		return err
	}

	logrus.Infof("Gas calculation created successfully")
	return nil
}

// draftCartInfo returns draft id and count of gases
func (r *Repository) draftCartInfo(creatorID uint) (uint, int64, error) {
	calc, err := r.ensureDraftCalculation(creatorID)
	if err != nil {
		return 0, 0, err
	}
	var count int64
	if err := r.db.Model(&ds.GasCalculation{}).Where("calculation_id = ?", calc.ID).Count(&count).Error; err != nil {
		return 0, 0, err
	}
	return calc.ID, count, nil
}

// RemoveGasFromDraft removes by gas id (without PK of m-m)
func (r *Repository) RemoveGasFromDraft(creatorID uint, gasID uint) error {
	calc, err := r.ensureDraftCalculation(creatorID)
	if err != nil {
		return err
	}
	return r.db.Where("calculation_id = ? AND gas_id = ?", calc.ID, gasID).Delete(&ds.GasCalculation{}).Error
}

// UpdateMM updates fields in m-m
func (r *Repository) UpdateMM(creatorID uint, gasID uint, sound *bool, quantity *int, position *int) error {
	calc, err := r.ensureDraftCalculation(creatorID)
	if err != nil {
		return err
	}
	updates := map[string]interface{}{}
	if sound != nil {
		updates["sound"] = *sound
	}
	if quantity != nil {
		updates["quantity"] = *quantity
	}
	if position != nil {
		updates["position"] = *position
	}
	if len(updates) == 0 {
		return errors.New("no updatable fields")
	}
	return r.db.Model(&ds.GasCalculation{}).Where("calculation_id = ? AND gas_id = ?", calc.ID, gasID).Updates(updates).Error
}
