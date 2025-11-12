package handler

import (
	"WEB/internal/app/ds"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// -------- Request/Response DTOs --------

// CreateCalculationRequest represents request for creating calculation
// @Description Create calculation request
type CreateCalculationRequest struct {
	Title string `json:"title" binding:"required" example:"My Calculation"`
	Text  string `json:"text" example:"Calculation description"`
}

// CalculationResponse represents calculation response for API
// @Description Calculation response object
type CalculationResponse struct {
	ID         uint      `json:"id" example:"1"`
	Status     string    `json:"status" example:"draft"`
	Text       string    `json:"text" example:"Calculation description"`
	DateCreate time.Time `json:"date_create"`
	CreatorID  uint      `json:"creator_id" example:"1"`
}

// -------- HTML Handlers --------

// AddGasToCalculation добавляет газ в расчет - POST запрос №4
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

// GetJournal отображает журнал расчетов - GET запрос №3
func (h *Handler) GetJournal(ctx *gin.Context) {
	// TODO: Заменить на реальный ID из авторизации
	creatorID := uint(1)

	// Получаем черновик расчета с газами
	calculation, err := h.Repository.GetDraftCalculation(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "journal.html", gin.H{
		"calculation": calculation,
		"gases":       calculation.Gases,
		"cart_count":  len(calculation.Gases),
	})
}

// RemoveGasFromCalculation логически удаляет газ из расчета - POST запрос №5
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

// UpdateGasParams обновляет параметры одного поля и пересчитывает давление
func (h *Handler) UpdateGasParams(ctx *gin.Context) {
	gasCalcIDStr := ctx.Param("id")
	gasCalcID, err := strconv.ParseUint(gasCalcIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Получаем все параметры из формы
	params := map[string]float64{}

	if initialPressure := ctx.PostForm("initial_pressure"); initialPressure != "" {
		if val, err := strconv.ParseFloat(initialPressure, 64); err == nil {
			params["initial_pressure"] = val
		}
	}
	if initialTemp := ctx.PostForm("initial_temperature"); initialTemp != "" {
		if val, err := strconv.ParseFloat(initialTemp, 64); err == nil {
			params["initial_temperature"] = val
		}
	}
	if finalTemp := ctx.PostForm("final_temperature"); finalTemp != "" {
		if val, err := strconv.ParseFloat(finalTemp, 64); err == nil {
			params["final_temperature"] = val
		}
	}
	if volume := ctx.PostForm("volume"); volume != "" {
		if val, err := strconv.ParseFloat(volume, 64); err == nil {
			params["volume"] = val
		}
	}
	if gasAmount := ctx.PostForm("gas_amount"); gasAmount != "" {
		if val, err := strconv.ParseFloat(gasAmount, 64); err == nil {
			params["gas_amount"] = val
		}
	}

	// Выполняем расчет и получаем результат
	calculatedPressure, err := h.Repository.CalculateGasPressure(uint(gasCalcID), params)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект с сообщением о результате
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/journal?message=calculated&pressure=%.4f", calculatedPressure))
}

// SubmitCalculation отправляет расчет на модерацию
func (h *Handler) SubmitCalculation(ctx *gin.Context) {
	calculationIDStr := ctx.Param("id")
	calculationID, err := strconv.ParseUint(calculationIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	creatorID := h.Repository.FixedCreatorID()

	if err := h.Repository.SubmitCalculation(uint(calculationID), creatorID); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект на страницу успеха или обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal?message=submitted")
}

// CalculateGasPressure рассчитывает давление для одного газа
func (h *Handler) CalculateGasPressure(ctx *gin.Context) {
	gasCalcIDStr := ctx.Param("id")
	gasCalcID, err := strconv.ParseUint(gasCalcIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Парсим параметры из формы
	params := map[string]float64{}

	if gasAmount, err := strconv.ParseFloat(ctx.PostForm("gas_amount"), 64); err == nil && gasAmount > 0 {
		params["gas_amount"] = gasAmount
	}
	if finalTemp, err := strconv.ParseFloat(ctx.PostForm("final_temperature"), 64); err == nil && finalTemp > 0 {
		params["final_temperature"] = finalTemp
	}
	if volume, err := strconv.ParseFloat(ctx.PostForm("volume"), 64); err == nil && volume > 0 {
		params["volume"] = volume
	}
	if initialPressure, err := strconv.ParseFloat(ctx.PostForm("initial_pressure"), 64); err == nil {
		params["initial_pressure"] = initialPressure
	}
	if initialTemp, err := strconv.ParseFloat(ctx.PostForm("initial_temperature"), 64); err == nil {
		params["initial_temperature"] = initialTemp
	}

	// Выполняем расчет
	calculatedPressure, err := h.Repository.CalculateGasPressure(uint(gasCalcID), params)
	if err != nil {
		// Показываем ошибку пользователю
		creatorID := h.Repository.FixedCreatorID()
		calculation, _ := h.Repository.GetDraftCalculation(creatorID)

		ctx.HTML(http.StatusOK, "journal.html", gin.H{
			"calculation": calculation,
			"gases":       calculation.Gases,
			"cart_count":  len(calculation.Gases),
			"error":       err.Error(),
		})
		return
	}

	// Редирект с сообщением о результате
	ctx.Redirect(http.StatusFound, fmt.Sprintf("/journal?message=calculated&pressure=%.4f", calculatedPressure))
}

// CalculateAllGases рассчитывает все газы в расчете
func (h *Handler) CalculateAllGases(ctx *gin.Context) {
	creatorID := h.Repository.FixedCreatorID()

	calculation, err := h.Repository.GetDraftCalculation(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Игнорируем возвращаемое значение результатов
	_, err = h.Repository.CalculateAllGases(calculation.ID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/journal")
}

// -------- Calculation REST --------

type apiCalcListFilter struct {
	Status   string `form:"status"`
	DateFrom string `form:"date_from"`
	DateTo   string `form:"date_to"`
}

type apiCalcUpdate struct {
	Text *string `json:"text"`
}

// ApiListCalculations godoc
// @Summary List all calculations (Moderator only)
// @Description Get list of all calculations - requires moderator role
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status"
// @Param date_from query string false "Filter by date from"
// @Param date_to query string false "Filter by date to"
// @Success 200 {array} map[string]interface{}
// @Failure 403 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/calculations [get]
func (h *Handler) ApiListCalculations(ctx *gin.Context) {
	var f apiCalcListFilter
	_ = ctx.ShouldBindQuery(&f)
	list, err := h.Repository.ListCalculations(f.Status, f.DateFrom, f.DateTo)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, list)
}

// ApiGetCalculation godoc
// @Summary Get calculation details
// @Description Get calculation details by ID
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calculation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/calculations/{id} [get]
func (h *Handler) ApiGetCalculation(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	item, gases, err := h.Repository.GetCalculationDetail(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"calculation": item, "gases": gases})
}

// ApiUpdateCalculation godoc
// @Summary Update calculation
// @Description Update calculation fields
// @Tags Calculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calculation ID"
// @Param request body apiCalcUpdate true "Update data"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/calculations/{id} [put]
func (h *Handler) ApiUpdateCalculation(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req apiCalcUpdate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	if err := h.Repository.UpdateCalculationFields(uint(id), req.Text); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ApiSubmitCalculation godoc
// @Summary Submit calculation
// @Description Submit calculation for moderation
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calculation ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/calculations/{id}/submit [post]
func (h *Handler) ApiSubmitCalculation(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	creatorID := h.Repository.FixedCreatorID()
	if err := h.Repository.SubmitCalculation(uint(id), creatorID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ApiCompleteCalculation godoc
// @Summary Complete calculation (Moderator only)
// @Description Mark calculation as completed
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calculation ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/calculations/{id}/complete [put]
func (h *Handler) ApiCompleteCalculation(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	moderatorID := uint(2) // фиксированный модератор
	if err := h.Repository.CompleteCalculation(uint(id), moderatorID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ApiRejectCalculation godoc
// @Summary Reject calculation (Moderator only)
// @Description Reject calculation
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calculation ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /api/calculations/{id}/reject [put]
func (h *Handler) ApiRejectCalculation(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	moderatorID := uint(2)
	if err := h.Repository.RejectCalculation(uint(id), moderatorID); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ApiDeleteCalculation godoc
// @Summary Delete calculation (logical)
// @Description Delete calculation by ID
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Calculation ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/calculations/{id} [delete]
func (h *Handler) ApiDeleteCalculation(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.Repository.DeleteCalculation(uint(id)); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// ApiMMDelete godoc
// @Summary Remove gas from draft
// @Description Remove gas from draft calculation
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Param id path int true "Gas ID"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/mm/gas/{id} [delete]
func (h *Handler) ApiMMDelete(ctx *gin.Context) {
	gasIDU64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID := h.Repository.FixedCreatorID()
	if err := h.Repository.RemoveGasFromDraft(creatorID, uint(gasIDU64)); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

type apiMMUpdateReq struct {
	Sound    *bool `json:"sound"`
	Quantity *int  `json:"quantity"`
	Position *int  `json:"position"`
}

// ApiMMUpdate godoc
// @Summary Update gas calculation fields
// @Description Update sound, quantity or position for gas in calculation
// @Tags Calculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Gas ID"
// @Param request body apiMMUpdateReq true "Update data"
// @Success 204
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/mm/gas/{id} [put]
func (h *Handler) ApiMMUpdate(ctx *gin.Context) {
	gasIDU64, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	var req apiMMUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	creatorID := h.Repository.FixedCreatorID()
	if err := h.Repository.UpdateMM(creatorID, uint(gasIDU64), req.Sound, req.Quantity, req.Position); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// UpdateAllGasParams обновляет параметры всех газов разом
func (h *Handler) UpdateAllGasParams(ctx *gin.Context) {
	creatorID := h.Repository.FixedCreatorID()
	calculation, err := h.Repository.GetDraftCalculation(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Обновляем параметры для каждого газа
	for _, gasCalc := range calculation.Gases {
		params := map[string]interface{}{}

		gasIDStr := strconv.FormatUint(uint64(gasCalc.ID), 10)

		if initialPressure := ctx.PostForm("initial_pressure_" + gasIDStr); initialPressure != "" {
			if val, err := strconv.ParseFloat(initialPressure, 64); err == nil {
				params["initial_pressure"] = val
			}
		}
		if initialTemp := ctx.PostForm("initial_temperature_" + gasIDStr); initialTemp != "" {
			if val, err := strconv.ParseFloat(initialTemp, 64); err == nil {
				params["initial_temperature"] = val
			}
		}
		if finalTemp := ctx.PostForm("final_temperature_" + gasIDStr); finalTemp != "" {
			if val, err := strconv.ParseFloat(finalTemp, 64); err == nil {
				params["final_temperature"] = val
			}
		}
		if volume := ctx.PostForm("volume_" + gasIDStr); volume != "" {
			if val, err := strconv.ParseFloat(volume, 64); err == nil {
				params["volume"] = val
			}
		}
		if gasAmount := ctx.PostForm("gas_amount_" + gasIDStr); gasAmount != "" {
			if val, err := strconv.ParseFloat(gasAmount, 64); err == nil {
				params["gas_amount"] = val
			}
		}

		// Обновляем параметры если есть изменения
		if len(params) > 0 {
			if err := h.Repository.UpdateGasCalculationParams(gasCalc.ID, params); err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
		}
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal?message=saved")
}

// SaveAllGasParams сохраняет все параметры без расчета
func (h *Handler) SaveAllGasParams(ctx *gin.Context) {
	creatorID := h.Repository.FixedCreatorID()
	calculation, err := h.Repository.GetDraftCalculation(creatorID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Сохраняем параметры для каждого газа
	for _, gasCalc := range calculation.Gases {
		params := map[string]interface{}{}

		gasIDStr := strconv.FormatUint(uint64(gasCalc.ID), 10)

		if initialPressure := ctx.PostForm("initial_pressure_" + gasIDStr); initialPressure != "" {
			if val, err := strconv.ParseFloat(initialPressure, 64); err == nil {
				params["initial_pressure"] = val
			}
		}
		if initialTemp := ctx.PostForm("initial_temperature_" + gasIDStr); initialTemp != "" {
			if val, err := strconv.ParseFloat(initialTemp, 64); err == nil {
				params["initial_temperature"] = val
			}
		}
		if finalTemp := ctx.PostForm("final_temperature_" + gasIDStr); finalTemp != "" {
			if val, err := strconv.ParseFloat(finalTemp, 64); err == nil {
				params["final_temperature"] = val
			}
		}
		if volume := ctx.PostForm("volume_" + gasIDStr); volume != "" {
			if val, err := strconv.ParseFloat(volume, 64); err == nil {
				params["volume"] = val
			}
		}
		if gasAmount := ctx.PostForm("gas_amount_" + gasIDStr); gasAmount != "" {
			if val, err := strconv.ParseFloat(gasAmount, 64); err == nil {
				params["gas_amount"] = val
			}
		}

		// Сохраняем параметры если есть изменения
		if len(params) > 0 {
			if err := h.Repository.UpdateGasCalculationParams(gasCalc.ID, params); err != nil {
				h.errorHandler(ctx, http.StatusInternalServerError, err)
				return
			}
		}
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal?message=saved")
}

// ApiGetMyCalculations godoc
// @Summary Get user's calculations
// @Description Get list of calculations for authenticated user
// @Tags Calculations
// @Produce json
// @Security BearerAuth
// @Success 200 {array} CalculationResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/my-calculations [get]
func (h *Handler) ApiGetMyCalculations(ctx *gin.Context) {
	// ВРЕМЕННО: Возвращаем пустой массив для тестирования
	ctx.JSON(http.StatusOK, []CalculationResponse{})
}

// ApiCreateCalculation godoc
// @Summary Create calculation
// @Description Create a new calculation
// @Tags Calculations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateCalculationRequest true "Calculation data"
// @Success 201 {object} CalculationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/calculations [post]
func (h *Handler) ApiCreateCalculation(ctx *gin.Context) {
	userUUID, exists := ctx.Get("user_uuid")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	var req CreateCalculationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.GetUserByUUID(userUUID.(string))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	calculation := &ds.Calculation{
		Status:     "draft",
		Text:       sql.NullString{String: req.Text, Valid: req.Text != ""},
		DateCreate: time.Now(),
		CreatorID:  user.ID,
	}

	err = h.Repository.CreateCalculation(calculation)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Возвращаем DTO
	ctx.JSON(http.StatusCreated, CalculationResponse{
		ID:         calculation.ID,
		Status:     calculation.Status,
		Text:       calculation.Text.String,
		DateCreate: calculation.DateCreate,
		CreatorID:  calculation.CreatorID,
	})
}
