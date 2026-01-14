package cafe

import (
	"time"
)

type OrderStatus string

const (
	OrderPending    OrderStatus = "pending"
	OrderProcessing OrderStatus = "processing"
	OrderDone       OrderStatus = "done"
	OrderCancelled  OrderStatus = "cancelled"
	OrderPaid       OrderStatus = "paid"
)

type OrderCafe struct {
	ID           uint          `json:"id" gorm:"primaryKey"`
	CustomerName string        `json:"customer_name" gorm:"not null;size:255"`
	TotalPrice   float64       `json:"total_price" gorm:"not null"`
	Status       OrderStatus   `json:"status" gorm:"not null;default:'pending'"`
	Items        []OrderItemCafe `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
}

type OrderItemCafe struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	OrderID   uint    `json:"order_id" gorm:"not null"`
	ProductID uint    `json:"product_id" gorm:"not null"`
	Quantity  int     `json:"quantity" gorm:"not null;min=1"`
	Price     float64 `json:"price" gorm:"not null"` // Harga saat order (untuk hindari perubahan harga product)
	Product   ProductCafe `json:"product" gorm:"foreignKey:ProductID"`
}

type OrderCafeCreate struct {
	CustomerName string               `json:"customer_name" binding:"required"`
	Items        []OrderItemCafeCreate `json:"items" binding:"required,min=1,dive"`
}

type OrderItemCafeCreate struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type OrderCafeUpdate struct {
	Status *OrderStatus `json:"status"`
}