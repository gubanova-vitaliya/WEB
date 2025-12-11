package pkg

import (
	"WEB/internal/app/config"
	"WEB/internal/app/ds"
	"WEB/internal/app/handler"
	"WEB/internal/app/redis"
	"WEB/internal/app/repository"
	"WEB/internal/app/role"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func NewApp(c *config.Config, r *gin.Engine, h *handler.Handler, repo *repository.Repository, redisClient *redis.Client) (*Application, error) {
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

	a.Handler.RegisterHandler(a.Router)
	a.Handler.RegisterStatic(a.Router)
	a.Handler.RegisterAPI(a.Router)

	// Добавляем базовые маршруты без авторизации для тестирования
	a.Router.POST("/login", a.Login)
	a.Router.POST("/sign_up", a.Register)
	a.Router.GET("/ping", a.Ping)

	serverAddress := fmt.Sprintf("%s:%d", a.Config.ServiceHost, a.Config.ServicePort)
	if err := a.Router.Run(serverAddress); err != nil {
		logrus.Fatal(err)
	}
	logrus.Info("Server down")
}

// loginReq represents login request parameters
// @Description Login credentials
type loginReq struct {
	Login    string `json:"login" example:"testuser"`
	Password string `json:"password" example:"password123"`
}

// loginResp represents login response
// @Description Login response with JWT token
type loginResp struct {
	ExpiresIn   int64  `json:"expires_in" example:"86400"` // 24 hours in seconds
	AccessToken string `json:"access_token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	TokenType   string `json:"token_type" example:"Bearer"`
}

// registerReq represents registration request parameters
// @Description User registration data
type registerReq struct {
	Name  string `json:"name" example:"John Doe"`
	Pass  string `json:"pass" example:"securepassword"`
	Email string `json:"email" example:"john@example.com"`
}

// registerResp represents registration response
// @Description Registration response
type registerResp struct {
	Ok bool `json:"ok" example:"true"`
}

// pingResp represents ping response
// @Description Ping response
type pingResp struct {
	Status string `json:"status" example:"pong"`
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
	req := &loginReq{}

	err := json.NewDecoder(ctx.Request.Body).Decode(req)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем пользователя из базы данных
	user, err := a.Repository.GetUserByLogin(req.Login)
	if err != nil {
		logrus.WithError(err).Warn("User not found")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid credentials"})
		return
	}

	// Проверяем пароль
	err = a.Repository.VerifyPassword(user.Password, req.Password)
	if err != nil {
		logrus.WithError(err).Warn("Invalid password")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "Invalid credentials"})
		return
	}
	userRole := role.FromString(user.Role)
	// Генерируем JWT токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &ds.JWTClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
		UserUUID: user.UUID,
		Role:     userRole, // Используем преобразованную роль
	})

	tokenString, err := token.SignedString([]byte(a.Config.JWT.Secret))
	if err != nil {
		logrus.WithError(err).Error("Failed to sign token")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	ctx.JSON(http.StatusOK, loginResp{
		ExpiresIn:   24 * 3600, // 24 часа в секундах
		AccessToken: tokenString,
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
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Pass == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Password is required"})
		return
	}

	if req.Name == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Генерируем хеш пароля
	hashedPassword, err := a.Repository.GenerateHashString(req.Pass)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Создаем пользователя
	user := &ds.User{
		UUID:     uuid.New(),
		Role:     "buyer",
		Name:     req.Name,
		Login:    req.Name, // используем имя как логин для простоты
		Email:    req.Email,
		Password: hashedPassword,
	}

	err = a.Repository.Register(user)
	if err != nil {
		logrus.Errorf("Registration failed: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logrus.Infof("User registered successfully: %s", req.Name)

	ctx.JSON(http.StatusOK, &registerResp{
		Ok: true,
	})
}

// Ping godoc
// @Summary Ping endpoint
// @Description Check if server is running
// @Tags Tests
// @Produce json
// @Success 200 {object} pingResp
// @Router /ping [get]
func (a *Application) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, pingResp{
		Status: "pong",
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
