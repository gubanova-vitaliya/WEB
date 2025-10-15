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

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  h.Repository.GetCartCount(),
	})
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

	var data map[string]interface{}
	if err := ctx.BindJSON(&data); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	err = h.Repository.UpdateCalculationParams(creatorID, uint(gasCalculationID), data)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
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

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  h.Repository.GetCartCount(),
	})
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
		if gasID, ok := gas["gas_calculation_id"].(uint); ok && gasID == uint(gasCalculationID) {
			calculationData = gas
			break
		}
	}

	if calculationData == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"status": "error",
		})
		return
	}

	// Извлекаем данные для расчета
	initialPressure, _ := calculationData["initial_pressure"].(float64)
	initialTemperature, _ := calculationData["initial_temperature"].(float64)
	finalTemperature, _ := calculationData["final_temperature"].(float64)

	// Проверяем, что все необходимые данные есть
	if initialPressure == 0 || initialTemperature == 0 || finalTemperature == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status": "error",
		})
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

	ctx.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"final_pressure": finalPressure,
	})
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

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  0,
	})
}
