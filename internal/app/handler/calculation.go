// handler/calculation.go
package handler

import (
	"net/http"
	"strconv"

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
