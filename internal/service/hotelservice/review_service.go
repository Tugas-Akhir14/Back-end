package hotelservice

import (
	"backend/internal/models/auth"
	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
	"backend/internal/repository/admin" // GUNAKAN YANG SUDAH ADA
	"fmt"
)

type ReviewService interface {
	Create(input CreateInput, ip string, adminID uint) error
	GetApproved() ([]hotel.GuestReview, error)
	GetPending() ([]hotel.GuestReview, error)
	Approve(id uint) error
	Delete(id uint) error
}

type service struct {
	reviewRepo repohotel.ReviewRepository
	adminRepo  admin.AdminRepository // GUNAKAN YANG SUDAH ADA
}

type CreateInput struct {
	Rating    int
	Comment   string
	GuestName string
}

func NewReviewService(reviewRepo repohotel.ReviewRepository, adminRepo admin.AdminRepository) ReviewService {
	return &service{
		reviewRepo: reviewRepo,
		adminRepo:  adminRepo,
	}
}

func (s *service) Create(input CreateInput, ip string, adminID uint) error {
	// Validasi role via adminRepo
	adminData, err := s.adminRepo.FindByID(adminID)
	if err != nil {
		return fmt.Errorf("gagal memeriksa user: %v", err)
	}
	if adminData == nil {
		return fmt.Errorf("user tidak ditemukan")
	}
	if adminData.Role != auth.RoleGuest {
		return fmt.Errorf("hanya role guest yang dapat mengirim ulasan")
	}

	limited, err := s.reviewRepo.CheckRateLimit(ip, 3)
	if err != nil {
		return err
	}
	if limited {
		return fmt.Errorf("rate limit exceeded")
	}

	rev := &hotel.GuestReview{
		Rating:     input.Rating,
		Comment:    input.Comment,
		GuestName:  input.GuestName,
		IPAddress:  ip,
		IsApproved: false,
		AdminID:    adminID,
	}
	return s.reviewRepo.Create(rev)
}

func (s *service) GetApproved() ([]hotel.GuestReview, error) {
	return s.reviewRepo.GetApproved()
}

func (s *service) GetPending() ([]hotel.GuestReview, error) {
	return s.reviewRepo.GetPending()
}

func (s *service) Approve(id uint) error {
	return s.reviewRepo.Approve(id)
}

func (s *service) Delete(id uint) error {
	return s.reviewRepo.Delete(id)
}