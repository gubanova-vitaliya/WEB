package handler

import (
	"WEB/internal/app/ds"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllGases(ctx *gin.Context) {
	var gases []ds.Gas
	var err error

	query := ctx.Query("query")
	if query == "" {
		gases, err = h.Repository.GetAllGases()
	} else {
		gases, err = h.Repository.SearchGases(query)

		if err == nil && len(gases) == 1 {
			logrus.Printf("Found exactly one gas for query '%s', redirecting to gas detail page", query)
			ctx.Redirect(http.StatusFound, fmt.Sprintf("/gases/%d", gases[0].ID))
			return
		}
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"gases":        gases,
		"journalCount": h.Repository.GetJournalCountDefault(),
		"query":        query,
	})
}

func (h *Handler) GetGasById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid gas ID"})
		logrus.Error(err)
		return
	}

	gas, err := h.Repository.GetGasByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		logrus.Error(err)
		return
	}

	if gas == nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Gas not found"})
		return
	}

	ctx.HTML(http.StatusOK, "gas_detail.html", gin.H{
		"gas":          gas,
		"journalCount": h.Repository.GetJournalCountDefault(),
	})
}

func (h *Handler) GetJournalPage(ctx *gin.Context) {
	// Используем метод с курсором для получения данных
	calculations, err := h.Repository.GetAllCalculationsWithGases()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.HTML(http.StatusOK, "calculations.html", gin.H{
		"calculations": calculations,
		"journalCount": h.Repository.GetJournalCountDefault(),
	})
}

func (h *Handler) AddToCalculations(ctx *gin.Context) {
	gasIDStr := ctx.PostForm("gas_id")
	gasID, err := strconv.ParseUint(gasIDStr, 10, 32)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Проверяем существование газа
	gas, err := h.Repository.GetGasByID(int(gasID))
	if err != nil || gas == nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("gas not found"))
		return
	}

	// Используем метод с курсором для добавления
	err = h.Repository.AddGasToJournalWithCursor(uint(gasID), 1)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	logrus.Printf("Added gas '%s' to journal using cursor", gas.Title)

	// Возвращаем успешный статус
	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Газ добавлен в журнал",
		"count":   h.Repository.GetJournalCountDefault(),
	})
}

func (h *Handler) ClearCalculations(ctx *gin.Context) {
	// Используем метод с курсором для очистки
	err := h.Repository.ClearAllCalculations(1)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/journal")
}
