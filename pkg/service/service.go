package service

import "wbtech/pkg/repository"

type Order interface {
}

type Service struct {
	Order
}

func NewService(repos *repository.Repository) *Service {
	return &Service{}
}
