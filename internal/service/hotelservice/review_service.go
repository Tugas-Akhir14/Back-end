package hotelservice

import (
	"backend/internal/models/hotel"
	"backend/internal/repository/repohotel"
	"fmt"
)

type ReviewService interface {
	Create(input CreateInput, ip string) error
	GetApproved() ([]hotel.GuestReview, error)
	GetPending() ([]hotel.GuestReview, error)
	Approve(id uint) error
	Delete(id uint) error
}

type service struct {
	repo repohotel.ReviewRepository
}

type CreateInput struct {
	Rating    int
	Comment   string
	GuestName string
}

func NewReviewService(repo repohotel.ReviewRepository) ReviewService {
	return &service{repo}
}

func (s *service) Create(input CreateInput, ip string) error {
	// Gunakan method repository, bukan akses db langsung
	limited, err := s.repo.CheckRateLimit(ip, 3)
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
	}
	return s.repo.Create(rev)
}

func (s *service) GetApproved() ([]hotel.GuestReview, error) {
	return s.repo.GetApproved()
}

func (s *service) GetPending() ([]hotel.GuestReview, error) {
	return s.repo.GetPending()
}

func (s *service) Approve(id uint) error {
	return s.repo.Approve(id)
}

func (s *service) Delete(id uint) error {
	return s.repo.Delete(id)
}