package handler

import (
	"WEB/internal/app/repository"

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

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты, чтобы не писать все в одном месте
// handler/handler.go
func (h *Handler) RegisterHandler(router *gin.Engine) {
	// 1. GET-запрос на просмотр всех карточек на главной странице
	router.GET("/gas", h.GetAllGases)

	// 2. GET-запрос на просмотр одной карточки
	router.GET("/gas/:id", h.GetGasById)

	// 3. GET-запрос на просмотр текущего расчета в журнале
	router.GET("/journal", h.GetJournal)

	// 4. POST-запрос на добавление расчета в журнал
	router.POST("/calculation/add", h.AddGasToCalculation)

	// 5. POST-запрос на логическое удаление расчета из журнала
	router.POST("/calculation/:id/remove", h.RemoveGasFromCalculation)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "./resources")
}

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}
