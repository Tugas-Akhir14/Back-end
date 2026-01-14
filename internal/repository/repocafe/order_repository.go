package repocafe

import (
	"backend/internal/models/cafe"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(order *cafe.OrderCafe) error
	FindAll() ([]cafe.OrderCafe, error)
	FindByID(id uint) (*cafe.OrderCafe, error)
	Update(order *cafe.OrderCafe) error
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db}
}

func (r *orderRepository) Create(order *cafe.OrderCafe) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindAll() ([]cafe.OrderCafe, error) {
	var orders []cafe.OrderCafe
	return orders, r.db.Preload("Items.Product").Find(&orders).Error
}

func (r *orderRepository) FindByID(id uint) (*cafe.OrderCafe, error) {
	var order cafe.OrderCafe
	err := r.db.Preload("Items.Product").First(&order, id).Error
	return &order, err
}

func (r *orderRepository) Update(order *cafe.OrderCafe) error {
	return r.db.Save(order).Error
}