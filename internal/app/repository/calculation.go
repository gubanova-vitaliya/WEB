package repository

import (
	"WEB/internal/app/ds"
	"errors"
	"time"
)

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
	calc, err := r.ensureDraftCalculation(creatorID)
	if err != nil {
		return err
	}
	// check gas exists
	var gas ds.Gas
	if err := r.db.First(&gas, gasID).Error; err != nil {
		return err
	}
	// create m-m unique pair
	mm := ds.GasCalculation{CalculationID: calc.ID, GasID: gasID, Sound: true}
	if err := r.db.Create(&mm).Error; err != nil {
		return err
	}
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

// UpdateMM updates fields in m-m (here only Sound)
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

// -------- Higher level calculation operations --------

func (r *Repository) ListCalculations(status string, dateFrom string, dateTo string) ([]map[string]interface{}, error) {
	q := r.db.Model(&ds.Calculation{}).Where("status <> ?", "deleted").Where("status <> ?", "draft").Preload("Creator").Preload("Moderator")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if dateFrom != "" {
		q = q.Where("date_form >= ?", dateFrom)
	}
	if dateTo != "" {
		q = q.Where("date_form <= ?", dateTo)
	}
	var items []ds.Calculation
	if err := q.Order("date_create desc").Find(&items).Error; err != nil {
		return nil, err
	}
	out := make([]map[string]interface{}, 0, len(items))
	for _, c := range items {
		out = append(out, map[string]interface{}{
			"id":              c.ID,
			"status":          c.Status,
			"text":            c.Text,
			"date_create":     c.DateCreate,
			"date_update":     c.DateUpdate,
			"date_finish":     c.DateFinish,
			"creator_login":   c.Creator.Login,
			"moderator_login": c.Moderator.Login,
		})
	}
	return out, nil
}

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
			"gas_id":      m.GasID,
			"title":       m.Gas.Title,
			"formula":     m.Gas.Formula,
			"molar_mass":  m.Gas.MolarMass,
			"image_url":   m.Gas.ImageURL,
			"description": m.Gas.Description,
			"sound":       m.Sound,
			"quantity":    m.Quantity,
			"position":    m.Position,
		})
	}
	return &c, list, nil
}

func (r *Repository) UpdateCalculationFields(id uint, text *string) error {
	updates := map[string]interface{}{"date_update": time.Now()}
	if text != nil {
		updates["text"] = *text
	}
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) SubmitCalculation(id uint, creatorID uint) error {
	var c ds.Calculation
	if err := r.db.First(&c, id).Error; err != nil {
		return err
	}
	if c.CreatorID != creatorID {
		return errors.New("only creator can submit")
	}
	if c.Status != "draft" {
		return errors.New("only draft can be submitted")
	}
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":      "formed",
		"date_update": time.Now(),
	}).Error
}

func (r *Repository) CompleteCalculation(id uint, moderatorID uint) error {
	var c ds.Calculation
	if err := r.db.First(&c, id).Error; err != nil {
		return err
	}
	if c.Status != "formed" {
		return errors.New("only formed can be completed")
	}
	if c.GasAmount.Valid && c.FinalTemperature.Valid && c.Volume.Valid {
		const R = 8.314462618
		fp := (c.GasAmount.Float64 * R * c.FinalTemperature.Float64) / c.Volume.Float64
		_ = r.db.Model(&ds.Calculation{}).Where("id = ?", id).Update("final_pressure", fp).Error
	}
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"moderator_id": moderatorID,
		"date_finish":  time.Now(),
		"date_update":  time.Now(),
	}).Error
}

func (r *Repository) RejectCalculation(id uint, moderatorID uint) error {
	var c ds.Calculation
	if err := r.db.First(&c, id).Error; err != nil {
		return err
	}
	if c.Status != "formed" {
		return errors.New("only formed can be rejected")
	}
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "rejected",
		"moderator_id": moderatorID,
		"date_finish":  time.Now(),
		"date_update":  time.Now(),
	}).Error
}

func (r *Repository) DeleteCalculation(id uint) error {
	return r.db.Model(&ds.Calculation{}).Where("id = ?", id).Update("status", "deleted").Error
}
