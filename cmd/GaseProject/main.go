package main

import (
	_ "WEB/docs"
	"WEB/internal/app/config"
	"WEB/internal/app/dsn"
	"WEB/internal/app/handler"
	"WEB/internal/app/redis"
	"WEB/internal/app/repository"
	"WEB/internal/pkg"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title GaseProject API
// @version 1.0
// @description API для управления газами и расчетами

// @contact.name API Support
// @contact.url http://localhost:8080
// @contact.email support@gaseproject.com

// @license.name MIT

// @host localhost:8080
// @BasePath /
// @schemes http

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Token
func main() {
	router := gin.Default()

	// Добавляем Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, err := repository.New(postgresString)
	if err != nil {
		logrus.Fatalf("error initializing repository: %v", err)
	}

	// Инициализируем Redis клиент
	ctx := context.Background()
	redisClient, err := redis.New(ctx, conf.Redis)
	if err != nil {
		logrus.Warnf("error initializing Redis (continuing without Redis): %v", err)
		redisClient = nil
	} else {
		logrus.Info("Redis client initialized successfully")
		defer redisClient.Close()
	}

	hand := handler.NewHandler(rep, redisClient)

	application, err := pkg.NewApp(conf, router, hand, rep, redisClient)
	if err != nil {
		logrus.Fatalf("error creating application: %v", err)
	}

	application.RunApp()
}
