package handler

import (
	"WEB/internal/app/repository"
	"os"
	"path/filepath"

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

// errorHandler для более удобного вывода ошибок
func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":      "error",
		"description": err.Error(),
	})
}

// RegisterHandler Функция, в которой мы отдельно регистрируем маршруты
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

	// 6. Новые маршруты для работы с расчетами
	router.POST("/calculation/:id/calculate", h.CalculateGasPressure)
	router.POST("/calculation/:id/update", h.UpdateGasParams)
	router.POST("/calculation/calculate-all", h.CalculateAllGases)
	router.POST("/calculation/update-all", h.UpdateAllGasParams) // Новый
	router.POST("/calculation/save-all", h.SaveAllGasParams)     // Новый
	router.POST("/calculation/:id/submit", h.SubmitCalculation)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	// Определяем корректные пути до шаблонов и статики независимо от рабочей директории
	templatesGlob := "templates/*.html"
	staticDir := "./resources"
	imagesDir := "./resources/images"

	if _, err := os.Stat("templates"); err != nil {
		// если запускают из cmd/GaseProject, поднимемся на два уровня
		altTemplates := filepath.FromSlash("../../templates/*.html")
		altStatic := filepath.FromSlash("../../resources")
		altImages := filepath.FromSlash("../../resources/images")
		templatesGlob = altTemplates
		staticDir = altStatic

		// Проверяем существование images директории
		if _, err := os.Stat(altImages); err == nil {
			imagesDir = altImages
		}
	}

	router.LoadHTMLGlob(templatesGlob)
	router.Static("/static", staticDir)

	// Отдельно обслуживаем images если директория существует
	if _, err := os.Stat(imagesDir); err == nil {
		router.Static("/images", imagesDir)
	}
}

// RegisterAPI регистрирует REST API с префиксом /api
func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")

	// Домен услуги (газ)
	api.GET("/gases", h.ApiGetGases)
	api.GET("/gases/:id", h.ApiGetGas)
	api.POST("/gases", h.ApiCreateGas)
	api.PUT("/gases/:id", h.ApiUpdateGas)
	api.DELETE("/gases/:id", h.ApiDeleteGas)
	api.POST("/gases/:id/image", h.ApiUploadGasImage)
	api.POST("/gases/:id/add-to-draft", h.ApiAddGasToDraft)

	// Домен заявки
	api.GET("/cart", h.ApiGetCart)
	api.GET("/calculations", h.ApiListCalculations)
	api.GET("/calculations/:id", h.ApiGetCalculation)
	api.PUT("/calculations/:id", h.ApiUpdateCalculation)
	api.PUT("/calculations/:id/submit", h.ApiSubmitCalculation)
	api.PUT("/calculations/:id/complete", h.ApiCompleteCalculation)
	api.PUT("/calculations/:id/reject", h.ApiRejectCalculation)
	api.DELETE("/calculations/:id", h.ApiDeleteCalculation)

	// Домен м-м
	api.DELETE("/mm/gas/:id", h.ApiMMDelete)
	api.PUT("/mm/gas/:id", h.ApiMMUpdate)

	// Домен пользователь
	api.POST("/auth/register", h.ApiRegister)
	api.POST("/auth/login", h.ApiLogin)
	api.POST("/auth/logout", h.ApiLogout)
	api.GET("/users/me", h.ApiMe)
	api.PUT("/users/me", h.ApiUpdateMe)
}

// -------- Users ----------
type apiRegisterReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type apiLoginReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) ApiRegister(ctx *gin.Context) {
	var req apiRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, 400, err)
		return
	}
	if err := h.Repository.UserRegister(req.Login, req.Password); err != nil {
		h.errorHandler(ctx, 400, err)
		return
	}
	ctx.Status(201)
}

func (h *Handler) ApiLogin(ctx *gin.Context) {
	var req apiLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, 400, err)
		return
	}
	if err := h.Repository.UserLogin(req.Login, req.Password); err != nil {
		h.errorHandler(ctx, 401, err)
		return
	}
	ctx.Status(204)
}

func (h *Handler) ApiLogout(ctx *gin.Context) {
	h.Repository.UserLogout()
	ctx.Status(204)
}

func (h *Handler) ApiMe(ctx *gin.Context) {
	u, err := h.Repository.UserMe()
	if err != nil {
		h.errorHandler(ctx, 401, err)
		return
	}
	ctx.JSON(200, u)
}

type apiUpdateMeReq struct {
	Login *string `json:"login"`
}

func (h *Handler) ApiUpdateMe(ctx *gin.Context) {
	var req apiUpdateMeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, 400, err)
		return
	}
	if err := h.Repository.UserUpdateMe(req.Login); err != nil {
		h.errorHandler(ctx, 400, err)
		return
	}
	ctx.Status(204)
}
