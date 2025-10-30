package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"WEB/internal/app/config"
	"WEB/internal/app/ds"
	"WEB/internal/app/handler"
	"WEB/internal/app/redis"
	"WEB/internal/app/repository"
	"WEB/internal/app/role"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type Application struct {
	Config      *config.Config
	Router      *gin.Engine
	Handler     *handler.Handler
	Repository  *repository.Repository
	RedisClient *redis.Client
}

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler, repo *repository.Repository) (*Application, error) {
	// Инициализируем Redis клиент
	ctx := context.Background()
	redisClient, err := redis.New(ctx, c.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %v", err)
	}

	return &Application{
		Config:      c,
		Router:      r,
		Handler:     h,
		Repository:  repo,
		RedisClient: redisClient,
	}, nil
}

func (a *Application) RunApp() {
	logrus.Info("Server start up")

	// Закрываем Redis соединение при завершении
	defer a.RedisClient.Close()

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)
	a.Handler.RegisterAPI(a.Router)

	// Добавляем эндпоинты с ролевой моделью
	a.Router.GET("/ping", a.WithAuthCheck(role.Manager, role.Admin), a.Ping)
	a.Router.POST("/login", a.Login)
	a.Router.POST("/sign_up", a.Register)
	a.Router.POST("/logout", a.WithAuthCheck(), a.Logout)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

// Структуры для запросов и ответов
type loginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResp struct {
	ExpiresIn   time.Duration `json:"expires_in"`
	AccessToken string        `json:"access_token"`
	TokenType   string        `json:"token_type"`
}

type registerReq struct {
	Name  string `json:"name"`
	Pass  string `json:"pass"`
	Email string `json:"email"`
}

type registerResp struct {
	Ok bool `json:"ok"`
}

type pingResp struct {
	Auth bool `json:"auth"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body loginReq true "Login credentials"
// @Success 200 {object} loginResp
// @Failure 400 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Router /login [post]
func (a *Application) Login(ctx *gin.Context) {
	cfg := a.Config
	req := &loginReq{}

	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	// Получаем пользователя из базы данных
	user, err := a.Repository.GetUserByLogin(req.Login)
	if err != nil {
		logrus.WithError(err).Warn("User not found")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Проверяем пароль
	err = a.Repository.VerifyPassword(user.Password, req.Password)
	if err != nil {
		logrus.WithError(err).Warn("Invalid password")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	// Генерируем JWT токен
	token := jwt.NewWithClaims(cfg.JWT.SigningMethod, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(cfg.JWT.ExpiresIn).Unix(),
			IssuedAt:  time.Now().Unix(),
			Issuer:    "gase-admin",
		},
		UserUUID: user.UUID,
		Role:     user.Role,
	})

	strToken, err := token.SignedString([]byte(cfg.JWT.Secret))
	if err != nil {
		logrus.WithError(err).Error("Failed to sign token")
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("cant create str token"))
		return
	}

	ctx.JSON(http.StatusOK, loginResp{
		ExpiresIn:   cfg.JWT.ExpiresIn,
		AccessToken: strToken,
		TokenType:   "Bearer",
	})
}

// Register godoc
// @Summary User registration
// @Description Register a new user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body registerReq true "Registration data"
// @Success 200 {object} registerResp
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sign_up [post]
func (a *Application) Register(ctx *gin.Context) {
	req := &registerReq{}

	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if req.Pass == "" {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("pass is empty"))
		return
	}

	if req.Name == "" {
		ctx.AbortWithError(http.StatusBadRequest, fmt.Errorf("name is empty"))
		return
	}

	// Генерируем хеш пароля
	hashedPassword, err := a.Repository.GenerateHashString(req.Pass)
	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, fmt.Errorf("failed to hash password"))
		return
	}

	// Создаем пользователя
	user := &ds.User{
		UUID:     uuid.New(),
		Role:     role.Buyer,
		Name:     req.Name,
		Login:    req.Name,
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = a.Repository.Register(user)
	if err != nil {
		logrus.Errorf("Registration failed: %v", err)
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	logrus.Infof("User registered successfully: %s", req.Name)

	ctx.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

// Logout godoc
// @Summary User logout
// @Description Logout user and invalidate token
// @Tags Auth
// @Param Authorization header string true "Bearer token"
// @Success 200
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /logout [post]
func (a *Application) Logout(ctx *gin.Context) {
	// Получаем заголовок
	jwtStr := ctx.GetHeader("Authorization")
	if !strings.HasPrefix(jwtStr, jwtPrefix) {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Отрезаем префикс
	jwtStr = jwtStr[len(jwtPrefix):]

	// Парсим токен чтобы убедиться в его валидности
	token, err := jwt.ParseWithClaims(jwtStr, &ds.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.Config.JWT.Secret), nil
	})
	if err != nil {
		ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}

	if !token.Valid {
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// Сохраняем в блеклист Redis
	err = a.RedisClient.WriteJWTToBlacklist(ctx.Request.Context(), jwtStr, a.Config.JWT.ExpiresIn)
	if err != nil {
		logrus.WithError(err).Error("Failed to write JWT to blacklist")
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	logrus.Info("User logged out successfully")
	ctx.Status(http.StatusOK)
}

// Ping godoc
// @Summary Ping endpoint
// @Description Check authentication status (requires Manager or Admin role)
// @Tags Tests
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} pingResp
// @Failure 403 {object} map[string]string
// @Router /ping [get]
func (a *Application) Ping(ctx *gin.Context) {
	// Middleware уже проверил авторизацию и роль
	claims, exists := ctx.Get("jwt_claims")
	if !exists {
		ctx.JSON(http.StatusOK, pingResp{Auth: false})
		return
	}

	jwtClaims := claims.(*ds.JWTClaims)
	logrus.Debugf("Ping request from user: %s with role: %s", jwtClaims.UserUUID, jwtClaims.Role)

	ctx.JSON(http.StatusOK, pingResp{
		Auth: true,
	})
}

// GetAuthClaims helper функция для получения JWT claims из контекста
func (a *Application) GetAuthClaims(ctx *gin.Context) (*ds.JWTClaims, bool) {
	claims, exists := ctx.Get("jwt_claims")
	if !exists {
		return nil, false
	}
	return claims.(*ds.JWTClaims), true
}
