// internal/service/serviceauth/admin_service.go
package serviceauth

import (
	// HAPUS INI: "backend/internal/config"
	"backend/internal/models/auth"
	"backend/internal/repository/admin"
	"backend/utils"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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

type UpdateProfileRequest struct {
	FullName    string `json:"full_name" binding:"omitempty,min=2"`
	Email       string `json:"email" binding:"omitempty,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty,min=10"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required,min=8"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}

type AdminService interface {
	Register(req *RegisterRequest) (*auth.AdminResponse, error)
	Login(email, password string) (*LoginResponse, error)
	GetProfile(id uint) (*auth.AdminResponse, error)
	ApproveUser(id uint, requesterRole auth.Role) error
	GetPendingAdmins(requesterRole auth.Role) ([]auth.AdminResponse, error)
	UpdateProfile(id uint, req *UpdateProfileRequest) (*auth.AdminResponse, error)
	ChangePassword(id uint, req *ChangePasswordRequest) error
	VerifyGuestOTP(email, otp string) error
}

type adminService struct {
	repo      admin.AdminRepository
	db		  *gorm.DB
	jwtSecret string
}

// DIUBAH: tambah parameter db
func NewAdminService(repo admin.AdminRepository, db *gorm.DB, jwtSecret string) AdminService {
	return &adminService{repo: repo, db: db, jwtSecret: jwtSecret}
}

// Helper generate OTP 6 digit
func generateOTP() string {
	const digits = "0123456789"
	otp := make([]byte, 6)
	for i := range otp {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		otp[i] = digits[n.Int64()]
	}
	return string(otp)
}

func (s *adminService) Register(req *RegisterRequest) (*auth.AdminResponse, error) {
    if req.Role == auth.RoleSuperAdmin {
        return nil, errors.New("superadmin hanya bisa dibuat dari seeder")
    }

    // Cek email sudah ada
    if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
        if req.Role == auth.RoleGuest && existing.Role == auth.RoleGuest {
            // logic kirim ulang OTP untuk guest...
            // (tetap sama)
        }
        return nil, errors.New("email sudah digunakan")
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // ────────────────────────────────────────────────
    // PERBAIKAN UTAMA: SEMUA ADMIN HARUS MENUNGGU APPROVE
    // Hanya superadmin yang langsung approved (dari seeder)
    // ────────────────────────────────────────────────
    var isApproved bool
    switch req.Role {
    case auth.RoleGuest:
        isApproved = false // tetap pakai OTP
    default:
        isApproved = false // admin_hotel, admin_souvenir, dll → WAJIB approve
    }

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

    // Kirim email sesuai role
    if req.Role == auth.RoleGuest {
        // logic OTP guest (tetap sama)
        otp := generateOTP()
        expiresAt := time.Now().Add(5 * time.Minute)
        if err := s.repo.SaveOTP(req.Email, otp, expiresAt); err == nil {
            go utils.SendGuestOTPEmail(req.Email, req.FullName, otp)
        }
    } else {
        // Kirim email ke admin baru → "menunggu persetujuan"
        go utils.SendApprovalPendingEmail(admin.Email, admin.FullName)
        
        // Opsional: kirim juga notifikasi ke superadmin (via email atau sistem notifikasi)
    }

    return toResponse(admin), nil
}

func (s *adminService) VerifyGuestOTP(email, otp string) error {
	ok, err := s.repo.VerifyOTP(email, otp)
	if err != nil || !ok {
		return errors.New("OTP salah atau sudah kadaluarsa")
	}

	var admin auth.Admin
	if err := s.db.Where("email = ?", email).First(&admin).Error; err != nil {
		return errors.New("akun tidak ditemukan")
	}

	if admin.Role != auth.RoleGuest {
		return errors.New("hanya guest yang perlu verifikasi OTP")
	}
	if admin.IsApproved {
		return errors.New("akun sudah aktif")
	}

	// Approve akun
	if err := s.db.Model(&admin).Update("is_approved", true).Error; err != nil {
		return err
	}

	s.repo.DeleteOTP(email)
	go utils.SendApprovalSuccessEmail(email, admin.FullName)
	return nil
}
func (s *adminService) Login(email, password string) (*LoginResponse, error) {
    admin, err := s.repo.FindByEmail(email)
    if err != nil || admin == nil {
        return nil, errors.New("email atau password salah")
    }

    if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
        return nil, errors.New("email atau password salah")
    }

    // Pengecekan approval wajib untuk SEMUA kecuali superadmin
    if !admin.IsApproved && admin.Role != auth.RoleSuperAdmin {
        if admin.Role == auth.RoleGuest {
            return nil, errors.New("akun guest belum diverifikasi. Silakan masukkan kode OTP yang dikirim ke email Anda.")
        }
        return nil, errors.New("akun Anda belum disetujui oleh Superadmin. Harap menunggu persetujuan.")
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

func (s *adminService) UpdateProfile(id uint, req *UpdateProfileRequest) (*auth.AdminResponse, error) {
	admin, err := s.repo.FindByID(id)
	if err != nil || admin == nil {
		return nil, errors.New("admin tidak ditemukan")
	}

	// Cek email unik kalau diubah
	if req.Email != "" && req.Email != admin.Email {
		if existing, _ := s.repo.FindByEmail(req.Email); existing != nil {
			return nil, errors.New("email sudah digunakan oleh akun lain")
		}
		admin.Email = req.Email
	}

	if req.FullName != "" {
		admin.FullName = req.FullName
	}
	if req.PhoneNumber != "" {
		admin.PhoneNumber = req.PhoneNumber
	}

	if err := s.repo.Update(admin); err != nil {
		return nil, err
	}

	return toResponse(admin), nil
}

func (s *adminService) ChangePassword(id uint, req *ChangePasswordRequest) error {
	admin, err := s.repo.FindByID(id)
	if err != nil || admin == nil {
		return errors.New("admin tidak ditemukan")
	}

	// Verifikasi password lama
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(req.OldPassword)); err != nil {
		return errors.New("password lama salah")
	}

	// Hash password baru
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	admin.Password = string(hashed)
	return s.repo.Update(admin)
}