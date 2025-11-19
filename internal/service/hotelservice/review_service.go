package hotelservice

import (
    "errors"
    "time"
    "backend/internal/models/hotel"
    "backend/internal/repository/repohotel"
)

type CreateReviewInput struct {
    Rating    int
    Comment   string
    GuestName string
}

type UpdateReviewInput struct {
    Rating  *int
    Comment *string
}

type ReviewService interface {
    Create(input CreateReviewInput, ip string, userID uint) error
    GetAll() ([]hotel.GuestReview, error)
    GetMyReviews(userID uint) ([]hotel.GuestReview, error)
    Update(id uint, input UpdateReviewInput, userID uint) error
    Delete(id uint, userID uint, isAdmin bool) error 
}

type reviewService struct {
    repo repohotel.ReviewRepository
}

func NewReviewService(repo repohotel.ReviewRepository) ReviewService {
    return &reviewService{repo}
}

func (s *reviewService) Create(input CreateReviewInput, ip string, userID uint) error {
	hasReviewed, err := s.repo.HasUserReviewed(userID)
	if err != nil {
		return err
	}
	if hasReviewed {
		return errors.New("Anda hanya diperbolehkan mengirim satu ulasan saja")
	}

	// Validasi rating
	if input.Rating < 1 || input.Rating > 5 {
		return errors.New("rating harus antara 1 sampai 5")
	}

	review := &hotel.GuestReview{
		Rating:    input.Rating,
		Comment:   input.Comment,
		GuestName: input.GuestName,
		IPAddress: ip,
		AdminID:   userID,
	}

	return s.repo.Create(review)
}

func (s *reviewService) GetAll() ([]hotel.GuestReview, error) {
    return s.repo.GetAll()
}

func (s *reviewService) GetMyReviews(userID uint) ([]hotel.GuestReview, error) {
    return s.repo.GetByUserID(userID)
}

func (s *reviewService) Update(id uint, input UpdateReviewInput, userID uint) error {
    review, err := s.repo.GetByID(id)
    if err != nil {
        return errors.New("ulasan tidak ditemukan")
    }
    if review.AdminID != userID {
        return errors.New("bukan pemilik ulasan")
    }
    // Maksimal edit 24 jam setelah dibuat
    if time.Since(review.CreatedAt) > 24*time.Hour {
        return errors.New("ulasan hanya bisa diedit dalam 24 jam")
    }

    if input.Rating != nil {
        review.Rating = *input.Rating
    }
    if input.Comment != nil {
        review.Comment = *input.Comment
    }

    return s.repo.Update(review)
}

func (s *reviewService) Delete(id uint, userID uint, isAdmin bool) error {
    review, err := s.repo.GetByID(id)
    if err != nil {
        return errors.New("ulasan tidak ditemukan")
    }

    if !isAdmin && review.AdminID != userID {
        return errors.New("bukan pemilik ulasan")
    }

    return s.repo.Delete(id)
}