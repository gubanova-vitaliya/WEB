package main

import (
	"WEB/internal/api"
	"log"
)

func main() {
	log.Println("App start!")
	api.StartServer()
	log.Println("App end!")
}
