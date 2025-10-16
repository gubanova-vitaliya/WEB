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
