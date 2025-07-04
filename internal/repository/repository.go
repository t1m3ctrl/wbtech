package repository

import (
	"context"
	"github.com/jmoiron/sqlx"
	"wbtech"
)

type Order interface {
	GetOrderById(ctx context.Context, id string) (wbtech.Order, error)
	CreateOrder(ctx context.Context, order wbtech.Order) error
}

type Repository struct {
	Order
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Order: NewOrderPostgres(db),
	}
}
