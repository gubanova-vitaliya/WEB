package handler

import (
	"WEB/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetGases(ctx *gin.Context) {
	var gases []repository.Gas
	var err error

	searchQuery := ctx.Query("query")
	if searchQuery == "" {
		gases, err = h.Repository.GetGases()
		if err != nil {
			logrus.Error(err)
		}
	} else {
		gases, err = h.Repository.GetGasesByTitle(searchQuery)
		if err != nil {
			logrus.Error(err)
		}
	}

	ctx.HTML(http.StatusOK, "index.html", gin.H{
		"time":  time.Now().Format("15:04:05"),
		"gases": gases,
		"query": searchQuery,
	})
}

// Страница конкретного газа
func (h *Handler) GetGas(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusBadRequest, "Неверный ID")
		return
	}

	gas, err := h.Repository.GetGas(id)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, "Газ не найден")
		return
	}

	ctx.HTML(http.StatusOK, "gases.html", gin.H{
		"gas": gas,
	})
}

// Страница журнала расчетов
func (h *Handler) GetCart(ctx *gin.Context) {
	// Получаем все газы из репозитория
	gases, err := h.Repository.GetGases()
	if err != nil {
		logrus.Error(err)
		// В случае ошибки покажем пустую страницу с сообщением
		ctx.HTML(http.StatusOK, "cart.html", gin.H{
			"gases": []repository.Gas{},
		})
		return
	}

	ctx.HTML(http.StatusOK, "cart.html", gin.H{
		"gases": gases,
	})
}

// Временная структура для расчетов (добавьте в начало файла)
type Calculation struct {
	ID              int
	GasName         string
	InitialPressure float64
	FinalPressure   float64
	InitialTemp     float64
	FinalTemp       float64
	Date            string
}
