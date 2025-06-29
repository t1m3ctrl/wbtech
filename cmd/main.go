package main

import (
	"log"
	"wbtech"
	"wbtech/pkg/handler"
	"wbtech/pkg/repository"
	"wbtech/pkg/service"
)

func main() {
	repos := repository.NewRepository()
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	srv := new(wbtech.Server)
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occurred while running http server: %s", err.Error())
	}
}
