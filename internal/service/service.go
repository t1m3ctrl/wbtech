package service

import (
	"context"
	"wbtech"
	"wbtech/internal/cache"
	"wbtech/internal/repository"
)

type Order interface {
	GetOrderById(ctx context.Context, id string) (wbtech.Order, error)
	ProcessOrder(ctx context.Context, order wbtech.Order) error
}

type Service struct {
	Order
}

func NewService(repos *repository.Repository, cache *cache.OrderCache) *Service {
	return &Service{
		Order: NewOrderService(repos.Order, cache),
	}
}
