package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"WEB/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllCalculations(ctx *gin.Context) {
	calculations, err := h.Repository.GetAllCalculations()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "calculations.html", gin.H{
		"data":          calculations,
		"journal_count": h.Repository.GetJournalCountDefault(),
	})
}

func (h *Handler) GetCalculationById(ctx *gin.Context) {
	strId := ctx.Param("id")
	id, err := strconv.Atoi(strId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	calculation, err := h.Repository.GetCalculationByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "calculation_detail.html", gin.H{ // Отдельный шаблон для деталей
		"calculation": calculation,
	})
}

func (h *Handler) CreateCalculation(ctx *gin.Context) {
	// Получаем gas_id из формы
	gasIDStr := ctx.PostForm("gas_id")
	gasID, err := strconv.Atoi(gasIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid gas ID"))
		return
	}

	// Создаем новый расчет с пустыми полями для ввода
	calculation := ds.Calculation{
		UserID:       1, // Временно используем ID пользователя 1
		GasID:        uint(gasID),
		Volume:       0.0, // Пустое поле для ввода
		Temperature1: 0.0, // Пустое поле для ввода
		Temperature2: 0.0, // Пустое поле для ввода
		Pressure1:    0.0, // Пустое поле для ввода
		Pressure2:    0.0, // Будет рассчитано после ввода данных
		Mass:         0.0, // Пустое поле для ввода
		Moles:        0.0, // Пустое поле для ввода
		DateCreate:   time.Now(),
		DateUpdate:   time.Now(),
		IsDeleted:    false,
	}

	err = h.Repository.CreateCalculation(&calculation)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/calculations")
}

func (h *Handler) ClearAllCalculations(ctx *gin.Context) {
	err := h.Repository.ClearAllCalculations()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/calculations")
}

func (h *Handler) GetAllGases(ctx *gin.Context) {
	var gases []ds.Gas
	var err error

	query := ctx.Query("query")
	if query == "" {
		// Если поисковый запрос пустой, показываем все газы
		gases, err = h.Repository.GetAllGases()
	} else {
		// Если есть поисковый запрос, ищем по нему
		gases, err = h.Repository.SearchGases(query)

		// Если найден ровно один газ, перенаправляем на его детальную страницу
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

	// Отладочный вывод
	logrus.Printf("Found %d gases (query: '%s'):", len(gases), query)
	for i, gas := range gases {
		logrus.Printf("Gas %d: %+v", i, gas)
	}

	// Передаем данные в шаблон
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
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid gas ID",
		})
		logrus.Error(err)
		return
	}

	gas, err := h.Repository.GetGasByID(id)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	if gas == nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Gas not found",
		})
		return
	}

	ctx.HTML(http.StatusOK, "gas_detail.html", gin.H{
		"gas":          gas,
		"journalCount": h.Repository.GetJournalCountDefault(),
	})
}

// DeleteCalculation обработчик для удаления расчета
func (h *Handler) DeleteCalculation(ctx *gin.Context) {
	// Поддержка как DELETE запросов, так и POST с _method=DELETE
	method := ctx.PostForm("_method")
	if method == "DELETE" || ctx.Request.Method == "DELETE" {
		strId := ctx.Param("id")
		id, err := strconv.Atoi(strId)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid calculation ID",
			})
			logrus.Error(err)
			return
		}

		err = h.Repository.DeleteCalculation(id)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}

		ctx.Redirect(http.StatusSeeOther, "/calculations")
	} else {
		ctx.JSON(http.StatusMethodNotAllowed, gin.H{
			"error": "Method not allowed",
		})
	}
}

func (h *Handler) GetHomePage(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "/gases")
}
