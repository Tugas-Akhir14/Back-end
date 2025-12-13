// internal/service/serviceauth/admin_service.go
package serviceauth

import (
	// HAPUS INI: "backend/internal/config"
	"backend/internal/models/auth"
	"backend/internal/repository/admin"
	"backend/utils"
	"crypto/rand"
	"errors"
	"log"
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
        // Kalau guest dan email sudah ada → hanya kirim OTP baru
        if req.Role == auth.RoleGuest && existing.Role == auth.RoleGuest {
            otp := generateOTP()
            expiresAt := time.Now().Add(5 * time.Minute)

            // PASTIKAN INI ADA: SaveOTP
            if err := s.repo.SaveOTP(req.Email, otp, expiresAt); err != nil {
                log.Printf("Gagal simpan OTP: %v", err)
            } else {
                go utils.SendGuestOTPEmail(req.Email, existing.FullName, otp)
            }
            return toResponse(existing), nil
        }
        return nil, errors.New("email sudah digunakan")
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    isApproved := req.Role != auth.RoleGuest // guest butuh OTP, yang lain butuh approve

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

    // Kirim OTP untuk guest
    if req.Role == auth.RoleGuest {
        otp := generateOTP()
        expiresAt := time.Now().Add(5 * time.Minute)

        // INI YANG WAJIB ADA!
        if err := s.repo.SaveOTP(req.Email, otp, expiresAt); err != nil {
            log.Printf("GAGAL SIMPAN OTP: %v", err)
        } else {
            log.Printf("OTP berhasil disimpan dan dikirim ke %s", req.Email)
            go utils.SendGuestOTPEmail(req.Email, req.FullName, otp)
        }
    } else {
        go utils.SendApprovalPendingEmail(admin.Email, admin.FullName)
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

	// PERBAIKAN DI SINI: Guest SEKARANG WAJIB IsApproved = true
	if !admin.IsApproved {
		// Hanya superadmin yang boleh login tanpa approval
		if admin.Role == auth.RoleSuperAdmin {
			// superadmin boleh langsung login
		} else {
			return nil, errors.New("akun Anda belum diverifikasi. Silakan cek email untuk kode OTP.")
		}
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