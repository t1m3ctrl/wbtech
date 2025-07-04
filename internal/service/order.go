package service

import (
	"context"
	"log/slog"
	"wbtech"
	"wbtech/internal/cache"
	"wbtech/internal/repository"
)

type OrderService struct {
	repo  repository.Order
	cache *cache.OrderCache
}

func NewOrderService(repo repository.Order, cache *cache.OrderCache) *OrderService {
	return &OrderService{
		repo:  repo,
		cache: cache,
	}
}

func (s *OrderService) GetOrderById(ctx context.Context, id string) (wbtech.Order, error) {
	if order, ok := s.cache.Get(id); ok {
		slog.Info("Cache HIT for order", "orderUID", id)
		return order, nil
	}

	slog.Info("Cache MISS for order", "orderUID", id)

	order, err := s.repo.GetOrderById(ctx, id)
	if err != nil {
		return wbtech.Order{}, err
	}

	if err := s.cache.Set(order); err != nil {
		slog.Error("Failed to cache order", "error", err)
	}

	return order, nil
}

func (s *OrderService) ProcessOrder(ctx context.Context, order wbtech.Order) error {
	if err := s.repo.CreateOrder(ctx, order); err != nil {
		return err
	}

	return nil
}
