package api

import (
	"WEB/internal/app/handler"
	"WEB/internal/app/repository"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Error("ошибка инициализации репозитория")
	}

	handler := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	// Добавьте этот обработчик для корневого пути
	r.GET("/", handler.GetGases)        // Главная страница по корневому URL
	r.GET("/gases", handler.GetGases)   // Главная страница
	r.GET("/gases/:id", handler.GetGas) // Страница подробного описания
	r.GET("/cart", handler.GetCart)     // Журнал расчетов

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
	log.Println("Server down")
}
