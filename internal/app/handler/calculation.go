package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// AddGasToCalculation добавляет газ в расчет
func (h *Handler) AddGasToCalculation(ctx *gin.Context) {
	gasIDStr := ctx.PostForm("gas_id")
	gasID, err := strconv.Atoi(gasIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Получаем данные газа
	gas, err := h.Repository.GetGasByID(gasID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	// Добавляем газ в расчет (в память)
	err = h.Repository.AddGasToCalculation(creatorID, gas)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект обратно на страницу газов
	ctx.Redirect(http.StatusFound, "/gas")
}

// GetJournal отображает журнал расчетов с добавленными газами
func (h *Handler) GetJournal(ctx *gin.Context) {
	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	gases, err := h.Repository.GetGasesInCalculation(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "journal.html", gin.H{
		"gases":      gases,
		"cart_count": h.Repository.GetCartCount(),
	})
}

// UpdateCalculationParams обновляет параметры расчета
func (h *Handler) UpdateCalculationParams(ctx *gin.Context) {
	gasCalculationIDStr := ctx.Param("id")
	gasCalculationID, err := strconv.ParseUint(gasCalculationIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Получаем данные из формы
	initialPressure := ctx.PostForm("initial_pressure")
	initialTemperature := ctx.PostForm("initial_temperature")
	finalTemperature := ctx.PostForm("final_temperature")
	volume := ctx.PostForm("volume")
	gasAmount := ctx.PostForm("gas_amount")

	data := make(map[string]interface{})

	if initialPressure != "" {
		if val, err := strconv.ParseFloat(initialPressure, 64); err == nil {
			data["initial_pressure"] = val
		}
	}
	if initialTemperature != "" {
		if val, err := strconv.ParseFloat(initialTemperature, 64); err == nil {
			data["initial_temperature"] = val
		}
	}
	if finalTemperature != "" {
		if val, err := strconv.ParseFloat(finalTemperature, 64); err == nil {
			data["final_temperature"] = val
		}
	}
	if volume != "" {
		if val, err := strconv.ParseFloat(volume, 64); err == nil {
			data["volume"] = val
		}
	}
	if gasAmount != "" {
		if val, err := strconv.ParseFloat(gasAmount, 64); err == nil {
			data["gas_amount"] = val
		}
	}

	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	err = h.Repository.UpdateCalculationParams(creatorID, uint(gasCalculationID), data)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal")
}

// RemoveGasFromCalculation удаляет газ из расчета
func (h *Handler) RemoveGasFromCalculation(ctx *gin.Context) {
	gasCalculationIDStr := ctx.Param("id")
	gasCalculationID, err := strconv.ParseUint(gasCalculationIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	err = h.Repository.RemoveGasFromCalculation(creatorID, uint(gasCalculationID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal")
}

// CalculatePressure рассчитывает конечное давление
func (h *Handler) CalculatePressure(ctx *gin.Context) {
	gasCalculationIDStr := ctx.Param("id")
	gasCalculationID, err := strconv.ParseUint(gasCalculationIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	// Получаем данные расчета
	gases, err := h.Repository.GetGasesInCalculation(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Ищем нужный газ
	var calculationData map[string]interface{}
	for _, gas := range gases {
		if gasCalcID, ok := gas["gas_calculation_id"].(uint); ok && gasCalcID == uint(gasCalculationID) {
			calculationData = gas
			break
		}
	}

	if calculationData == nil {
		ctx.Redirect(http.StatusFound, "/journal")
		return
	}

	// Извлекаем данные для расчета
	initialPressure, _ := calculationData["initial_pressure"].(float64)
	initialTemperature, _ := calculationData["initial_temperature"].(float64)
	finalTemperature, _ := calculationData["final_temperature"].(float64)

	// Проверяем, что все необходимые данные есть
	if initialPressure == 0 || initialTemperature == 0 || finalTemperature == 0 {
		ctx.Redirect(http.StatusFound, "/journal")
		return
	}

	// Расчет давления по закону Шарля: P2 = P1 * T2 / T1
	finalPressure := initialPressure * finalTemperature / initialTemperature

	// Сохраняем результат в память
	err = h.Repository.UpdateCalculationParams(creatorID, uint(gasCalculationID), map[string]interface{}{
		"final_pressure": finalPressure,
	})
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal")
}

// ClearAllCalculations очищает все расчеты
func (h *Handler) ClearAllCalculations(ctx *gin.Context) {
	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	err := h.Repository.ClearAllCalculations(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal")
}
