package repository

type Order interface {
}

type Repository struct {
	Order
}

func NewRepository() *Repository {
	return &Repository{}
}
