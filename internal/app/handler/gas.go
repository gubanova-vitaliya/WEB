package handler

import (
	"net/http"
	"strconv"

	"WEB/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetAllCalculations(ctx *gin.Context) {
	var calculations []ds.Calculation
	var err error

	search := ctx.Query("search")
	if search == "" {
		calculations, err = h.Repository.GetAllCalculations()
	} else {
		calculations, err = h.Repository.SearchCalculationsByName(search)
	}

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		logrus.Error(err)
		return
	}

	ctx.HTML(http.StatusOK, "calculations.html", gin.H{ // Используем отдельный шаблон
		"data":          calculations,
		"journal_count": h.Repository.GetJournalCountDefault(), // Используем новую функцию
		"search":        search,
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
	var calculation ds.Calculation

	if err := ctx.ShouldBind(&calculation); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Логика расчета давления по формуле идеального газа
	// P₂ = P₁ * T₂ / T₁ (температуры в Кельвинах)
	// Добавьте здесь расчет, если нужно

	err := h.Repository.CreateCalculation(&calculation)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/calculations")
}

func (h *Handler) GetAllGases(ctx *gin.Context) {
	gases, err := h.Repository.GetAllGases()
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	// Отладочный вывод
	logrus.Printf("Found %d gases:", len(gases))
	for i, gas := range gases {
		logrus.Printf("Gas %d: %+v", i, gas)
	}

	// Передаем данные в шаблон
	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"gases":        gases,
		"journalCount": h.Repository.GetJournalCountDefault(), // Используем новую функцию
		"query":        ctx.Query("query"),
	})
}

// Добавьте обработчик для главной страницы
func (h *Handler) GetHomePage(ctx *gin.Context) {
	ctx.Redirect(http.StatusFound, "/gases")
}
