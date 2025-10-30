package handler

import (
	"WEB/internal/app/ds"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

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
	params := map[string]interface{}{}

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

	// Обновляем параметры в БД
	if err := h.Repository.UpdateGasCalculationParams(uint(gasCalcID), params); err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Автоматически пересчитываем давление, если все необходимые параметры есть
	if finalTemp, hasFinalTemp := params["final_temperature"]; hasFinalTemp {
		if volume, hasVolume := params["volume"]; hasVolume {
			if gasAmount, hasGasAmount := params["gas_amount"]; hasGasAmount {
				// Выполняем расчет
				calcParams := map[string]float64{
					"final_temperature": finalTemp.(float64),
					"volume":            volume.(float64),
					"gas_amount":        gasAmount.(float64),
				}

				// Добавляем опциональные параметры если они есть
				if initialPressure, hasInitialPressure := params["initial_pressure"]; hasInitialPressure {
					calcParams["initial_pressure"] = initialPressure.(float64)
				}
				if initialTemp, hasInitialTemp := params["initial_temperature"]; hasInitialTemp {
					calcParams["initial_temperature"] = initialTemp.(float64)
				}

				// Вызываем расчет, но игнорируем возвращаемое значение давления
				_, _ = h.Repository.CalculateGasPressure(uint(gasCalcID), calcParams)
			}
		}
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal")
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
	if initialPressure, err := strconv.ParseFloat(ctx.PostForm("initial_pressure"), 64); err == nil {
		params["initial_pressure"] = initialPressure
	}
	if initialTemp, err := strconv.ParseFloat(ctx.PostForm("initial_temperature"), 64); err == nil {
		params["initial_temperature"] = initialTemp
	}
	if finalTemp, err := strconv.ParseFloat(ctx.PostForm("final_temperature"), 64); err == nil {
		params["final_temperature"] = finalTemp
	}
	if volume, err := strconv.ParseFloat(ctx.PostForm("volume"), 64); err == nil {
		params["volume"] = volume
	}
	if gasAmount, err := strconv.ParseFloat(ctx.PostForm("gas_amount"), 64); err == nil {
		params["gas_amount"] = gasAmount
	}

	// Выполняем расчет (игнорируем возвращаемое значение давления)
	_, err = h.Repository.CalculateGasPressure(uint(gasCalcID), params)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Редирект обратно в журнал
	ctx.Redirect(http.StatusFound, "/journal")
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

// ApiListCalculations GET /api/calculations?status=&date_from=&date_to=
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

// ApiGetCalculation GET /api/calculations/:id
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

// ApiUpdateCalculation PUT /api/calculations/:id (only editable fields)
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

// ApiSubmitCalculation PUT /api/calculations/:id/submit
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

// ApiCompleteCalculation PUT /api/calculations/:id/complete
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

// ApiRejectCalculation PUT /api/calculations/:id/reject
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

// ApiDeleteCalculation DELETE /api/calculations/:id (logical)
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

// ApiMMDelete DELETE /api/mm/gas/:id (id = gas_id), удаление из черновика без PK м-м
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

// ApiMMUpdate PUT /api/mm/gas/:id (id = gas_id), изменение полей м-м
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

// ApiGetMyCalculations GET /api/my-calculations - заявки текущего пользователя
func (h *Handler) ApiGetMyCalculations(ctx *gin.Context) {
	userUUID, exists := ctx.Get("user_uuid")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	// Получаем пользователя
	user, err := h.Repository.GetUserByUUID(userUUID.(string))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	// Получаем заявки пользователя
	calculations, err := h.Repository.GetCalculationsByUser(user.UUID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, calculations)
}

// ApiCreateCalculation POST /api/calculations - создание заявки
func (h *Handler) ApiCreateCalculation(ctx *gin.Context) {
	userUUID, exists := ctx.Get("user_uuid")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
		Text  string `json:"text"`
	}

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
		CreatorID:  user.ID, // Предполагаем, что у User есть ID
	}

	err = h.Repository.CreateCalculation(calculation)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, calculation)
}
