package handler

import (
	"github.com/gin-gonic/gin"
	"wbtech/pkg/service"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	api := router.Group("/api")
	{
		order := api.Group("/order")
		{
			order.GET("/:id", h.getOrder)
			order.POST("/", h.getOrder)
		}
	}

	return router
}
