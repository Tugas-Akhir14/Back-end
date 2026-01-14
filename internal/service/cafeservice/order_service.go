package cafeservice

import (
	"backend/internal/models/cafe"
	"backend/internal/repository/repocafe"
	"errors"
	"gorm.io/gorm"
)

type OrderService interface {
	Create(input cafe.OrderCafeCreate) (*cafe.OrderCafe, error)
	GetAll() ([]cafe.OrderCafe, error)
	GetByID(id uint) (*cafe.OrderCafe, error)
	Update(id uint, input cafe.OrderCafeUpdate) (*cafe.OrderCafe, error)
	ConfirmPayment(id uint) (*cafe.OrderCafe, error)
	CancelOrder(id uint) (*cafe.OrderCafe, error)
}

type orderService struct {
	repo        repocafe.OrderRepository
	productRepo repocafe.ProductRepository
	db          *gorm.DB // Untuk transaction
}

func NewOrderService(repo repocafe.OrderRepository, productRepo repocafe.ProductRepository, db *gorm.DB) OrderService {
	return &orderService{repo, productRepo, db}
}

func (s *orderService) Create(input cafe.OrderCafeCreate) (*cafe.OrderCafe, error) {
	order := &cafe.OrderCafe{
		CustomerName: input.CustomerName,
		Status:       cafe.OrderPending,
	}

	var total float64
	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, itemInput := range input.Items {
			product, err := s.productRepo.FindByID(itemInput.ProductID)
			if err != nil {
				return errors.New("product not found")
			}
			if product.Stok < itemInput.Quantity {
				return errors.New("insufficient stock for product: " + product.Nama)
			}

			item := cafe.OrderItemCafe{
				ProductID: itemInput.ProductID,
				Quantity:  itemInput.Quantity,
				Price:     product.Harga,
			}
			order.Items = append(order.Items, item)
			total += product.Harga * float64(itemInput.Quantity)

			// Kurangi stok
			product.Stok -= itemInput.Quantity
			if err := tx.Save(product).Error; err != nil {
				return err
			}
		}
		order.TotalPrice = total
		return tx.Create(order).Error
	})
	if err != nil {
		return nil, err
	}

	return s.repo.FindByID(order.ID)
}

func (s *orderService) GetAll() ([]cafe.OrderCafe, error) {
	return s.repo.FindAll()
}

func (s *orderService) GetByID(id uint) (*cafe.OrderCafe, error) {
	return s.repo.FindByID(id)
}

func (s *orderService) Update(id uint, input cafe.OrderCafeUpdate) (*cafe.OrderCafe, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if input.Status != nil {
		order.Status = *input.Status
	}

	if err := s.repo.Update(order); err != nil {
		return nil, err
	}
	return s.repo.FindByID(id)
}

func (s *orderService) ConfirmPayment(id uint) (*cafe.OrderCafe, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if order.Status == cafe.OrderPaid {
		return nil, errors.New("order already paid")
	}

	if order.Status == cafe.OrderCancelled {
		return nil, errors.New("cannot confirm payment for a cancelled order")
	}

	order.Status = cafe.OrderPaid

	if err := s.repo.Update(order); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func (s *orderService) CancelOrder(id uint) (*cafe.OrderCafe, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if order.Status == cafe.OrderCancelled {
		return nil, errors.New("order already cancelled")
	}

	if order.Status == cafe.OrderPaid {
		return nil, errors.New("cannot cancel a paid order")
	}

	// Optional: tambahkan validasi tambahan jika ingin blokir cancel saat sudah "done"
	// if order.Status == cafe.OrderDone { ... }

	order.Status = cafe.OrderCancelled

	if err := s.repo.Update(order); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

