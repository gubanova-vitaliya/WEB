package handler

import (
	"WEB/internal/app/ds"
	"WEB/internal/app/repository"
	"WEB/internal/app/role"
	"errors"
	"net/http"

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
	router.POST("/calculation/update-all", h.UpdateAllGasParams)
	router.POST("/calculation/save-all", h.SaveAllGasParams)
	router.POST("/calculation/:id/submit", h.SubmitCalculation)
}

// RegisterStatic То же самое, что и с маршрутами, регистрируем статику
func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/static", "./resources")
	router.Static("/images", "./resources/images")
}

// RegisterAPI регистрирует REST API с префиксом /api
func (h *Handler) RegisterAPI(router *gin.Engine) {
	api := router.Group("/api")

	// Публичные эндпоинты (доступны без авторизации)
	api.GET("/gases", h.ApiGetGases)
	api.GET("/gases/:id", h.ApiGetGas)
	api.POST("/auth/register", h.ApiRegister)
	api.POST("/auth/login", h.ApiLogin)
	api.POST("/auth/login-session", h.ApiLoginWithSession) // Новый метод с сессией

	// Защищенные эндпоинты (требуют авторизации)
	protected := api.Group("")
	protected.Use(h.AuthMiddleware())
	{
		// Пользовательские методы
		protected.GET("/users/me", h.ApiGetProfile)
		protected.PUT("/users/me", h.ApiUpdateMe)
		protected.POST("/auth/logout", h.ApiLogout)

		// Заявки пользователя
		protected.GET("/my-calculations", h.ApiGetMyCalculations)
		protected.POST("/calculations", h.ApiCreateCalculation)
		protected.PUT("/calculations/:id", h.ApiUpdateCalculation)
		protected.POST("/calculations/:id/submit", h.ApiSubmitCalculation)
	}

	// Модераторские эндпоинты
	moderator := api.Group("")
	moderator.Use(h.AuthMiddleware())
	moderator.Use(h.RoleMiddleware(role.Manager, role.Admin))
	{
		moderator.GET("/calculations", h.ApiListCalculations)
		moderator.PUT("/calculations/:id/complete", h.ApiCompleteCalculation)
		moderator.PUT("/calculations/:id/reject", h.ApiRejectCalculation)
		moderator.GET("/users", h.ApiGetAllUsers)
	}

	// Админские эндпоинты
	admin := api.Group("")
	admin.Use(h.AuthMiddleware())
	admin.Use(h.RoleMiddleware(role.Admin))
	{
		admin.DELETE("/users/:uuid", h.ApiDeleteUser)
		admin.PUT("/users/:uuid/role", h.ApiUpdateUserRole)
	}
}

// -------- Users Handlers --------

type apiRegisterReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
	Email    string `json:"email"`
	Name     string `json:"name" binding:"required"`
}

type apiLoginReq struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type apiUpdateMeReq struct {
	Login *string `json:"login"`
}

type userResp struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Login string `json:"login"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func (h *Handler) ApiRegister(ctx *gin.Context) {
	var req apiRegisterReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Генерируем хеш пароля
	hashedPassword, err := h.Repository.GenerateHashString(req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	user := &ds.User{
		Name:     req.Name,
		Login:    req.Login,
		Email:    req.Email,
		Role:     role.Buyer,
		Password: hashedPassword,
	}

	err = h.Repository.Register(user)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": userResp{
			UUID:  user.UUID.String(),
			Name:  user.Name,
			Login: user.Login,
			Email: user.Email,
			Role:  user.Role.String(),
		},
	})
}

func (h *Handler) ApiLogin(ctx *gin.Context) {
	var req apiLoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	user, err := h.Repository.AuthenticateUser(req.Login, req.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"user": userResp{
			UUID:  user.UUID.String(),
			Name:  user.Name,
			Login: user.Login,
			Email: user.Email,
			Role:  user.Role.String(),
		},
		"message": "Login successful (JWT generation to be implemented)",
	})
}

func (h *Handler) ApiGetProfile(ctx *gin.Context) {
	userUUID, exists := ctx.Get("user_uuid")
	if !exists {
		h.errorHandler(ctx, http.StatusUnauthorized, errors.New("user not authenticated"))
		return
	}

	user, err := h.Repository.GetUserByUUID(userUUID.(string))
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	ctx.JSON(http.StatusOK, userResp{
		UUID:  user.UUID.String(),
		Name:  user.Name,
		Login: user.Login,
		Email: user.Email,
		Role:  user.Role.String(),
	})
}

func (h *Handler) ApiRegisterOld(ctx *gin.Context) {
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

func (h *Handler) ApiLoginOld(ctx *gin.Context) {
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

// AuthMiddleware middleware для проверки JWT авторизации
func (h *Handler) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Реализация JWT проверки должна быть здесь
		// Временно разрешаем все запросы
		ctx.Next()
	}
}
