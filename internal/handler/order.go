package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"wbtech"
)

func (h *Handler) getOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.services.GetOrderById(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) createOrder(c *gin.Context) {
	var order wbtech.Order
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.services.Order.ProcessOrder(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}
