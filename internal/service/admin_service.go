package service

import (
	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type AdminService interface {
    Register(admin *models.Admin, confirmPassword string) error
    Login(email, password string) (string, error)
    GetProfile(id uint) (*models.Admin, error)
}

type adminService struct {
    repo      repository.AdminRepository
    jwtSecret string
}

func NewAdminService(repo repository.AdminRepository, jwtSecret string) AdminService {
    return &adminService{repo, jwtSecret}
}

// Register creates a new admin and hashes the password
func (s *adminService) Register(admin *models.Admin, confirmPassword string) error {
    // Validasi password dan confirm password
    if admin.Password != confirmPassword {
        return errors.New("password and confirm password do not match")
    }

    // Cek apakah email sudah terdaftar
    existingAdmin, err := s.repo.FindByEmail(admin.Email)
    if err == nil && existingAdmin != nil {
        return errors.New("email already registered")
    } else if err != nil && err.Error() != "record not found" {
        return errors.New("failed to check for existing admin email")
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
    if err != nil {
        return errors.New("failed to hash password")
    }
    admin.Password = string(hashedPassword)

    // Simpan admin ke database
    if err := s.repo.Create(admin); err != nil {
        return errors.New("failed to register admin")
    }

    return nil
}

// Login checks email and password, returns JWT token if valid
func (s *adminService) Login(email, password string) (string, error) {
	admin, err := s.repo.FindByEmail(email)
	if err != nil {
		return "", errors.New("invalid email or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", errors.New("invalid email or password")
	}

	now := time.Now().UTC()
	claims := jwt.MapClaims{
		"user_id": admin.ID,
		"email":   email,
		"iat":     now.Unix(),                     // detik
		"exp":     now.Add(24 * time.Hour).Unix(), // detik
		// jangan pakai milli/nano
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}
	return tokenString, nil
}

// GetProfile retrieves admin profile by ID
func (s *adminService) GetProfile(id uint) (*models.Admin, error) {
    admin, err := s.repo.FindByID(id)
    if err != nil {
        return nil, errors.New("admin not found")
    }
    return admin, nil
}
