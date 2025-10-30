// internal/service/serviceauth/admin_service.go
package serviceauth

import (
	// HAPUS INI: "backend/internal/config"
	"backend/internal/models/auth"
	"backend/internal/repository/admin"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	FullName        string    `json:"full_name" binding:"required"`
	Email           string    `json:"email" binding:"required,email"`
	PhoneNumber     string    `json:"phone_number" binding:"required"`
	Password        string    `json:"password" binding:"required,min=8"`
	ConfirmPassword string    `json:"confirm_password" binding:"required,eqfield=Password"`
	Role            auth.Role `json:"role" binding:"required,oneof=admin_hotel admin_souvenir admin_buku admin_cafe guest"`
}

type LoginResponse struct {
	Token string              `json:"token"`
	User  *auth.AdminResponse `json:"user"`
}

type AdminService interface {
	Register(req *RegisterRequest) (*auth.AdminResponse, error)
	Login(email, password string) (*LoginResponse, error)
	GetProfile(id uint) (*auth.AdminResponse, error)
	ApproveUser(id uint, requesterRole auth.Role) error
	GetPendingAdmins(requesterRole auth.Role) ([]auth.AdminResponse, error)
}

type adminService struct {
	repo      admin.AdminRepository
	jwtSecret string
}

func NewAdminService(repo admin.AdminRepository, jwtSecret string) AdminService {
	return &adminService{repo, jwtSecret}
}

func (s *adminService) Register(req *RegisterRequest) (*auth.AdminResponse, error) {
	if req.Role == auth.RoleSuperAdmin {
		return nil, errors.New("superadmin hanya bisa dibuat dari seeder")
	}

	if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
		return nil, errors.New("email sudah digunakan")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	isApproved := req.Role == auth.RoleGuest

	admin := &auth.Admin{
		FullName:    req.FullName,
		Email:       req.Email,
		PhoneNumber: req.PhoneNumber,
		Password:    string(hashed),
		Role:        req.Role,
		IsApproved:  isApproved,
	}

	if err := s.repo.Create(admin); err != nil {
		return nil, err
	}

	return toResponse(admin), nil
}

func (s *adminService) Login(email, password string) (*LoginResponse, error) {
	admin, err := s.repo.FindByEmail(email)
	if err != nil || admin == nil {
		return nil, errors.New("email atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return nil, errors.New("email atau password salah")
	}

	if admin.Role != auth.RoleGuest && admin.Role != auth.RoleSuperAdmin && !admin.IsApproved {
		return nil, errors.New("akun Anda belum disetujui oleh Superadmin")
	}

	token, err := s.generateToken(admin.ID, admin.Role)
	if err != nil {
		return nil, err
	}

	return &LoginResponse{
		Token: token,
		User:  toResponse(admin),
	}, nil
}

func (s *adminService) GetProfile(id uint) (*auth.AdminResponse, error) {
	admin, err := s.repo.FindByID(id)
	if err != nil || admin == nil {
		return nil, errors.New("admin tidak ditemukan")
	}
	return toResponse(admin), nil
}

func (s *adminService) ApproveUser(id uint, requesterRole auth.Role) error {
	if requesterRole != auth.RoleSuperAdmin {
		return errors.New("hanya superadmin yang bisa approve")
	}
	return s.repo.Approve(id)
}

func (s *adminService) GetPendingAdmins(requesterRole auth.Role) ([]auth.AdminResponse, error) {
	if requesterRole != auth.RoleSuperAdmin {
		return nil, errors.New("akses ditolak")
	}
	admins, err := s.repo.GetPending()
	if err != nil {
		return nil, err
	}
	res := make([]auth.AdminResponse, len(admins))
	for i, a := range admins {
		res[i] = *toResponse(&a)
	}
	return res, nil
}

func (s *adminService) generateToken(id uint, role auth.Role) (string, error) {
	claims := jwt.MapClaims{
		"user_id": id,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func toResponse(a *auth.Admin) *auth.AdminResponse {
	return &auth.AdminResponse{
		ID:          a.ID,
		FullName:    a.FullName,
		Email:       a.Email,
		PhoneNumber: a.PhoneNumber,
		Role:        a.Role,
		IsApproved:  a.IsApproved,
	}
}