package entities

import (
	"errors"
	"time"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"gorm.io/gorm"
)

type PasswordReset struct {
	ID        uuid.UUID      `gorm:"type:uuid;default:uuid_generate_v4();primaryKey;unique;not null" json:"id"`
	Email     string         `gorm:"index" json:"email"`
	Token     string         `gorm:"uniqueIndex" json:"token"`
	ExpiresAt time.Time      `gorm:"column:expires_at" json:"expires_at"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type RegisterRequest struct {
	FirstName   string  `json:"first_name"   binding:"required"`
	LastName    string  `json:"last_name"    binding:"required"`
	Email       string  `json:"email"        binding:"required,email"`
	Password    *string `json:"password"     binding:"omitempty,min=8"` // nullable/optional
	DateOfBirth string  `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	PhoneNumber string  `json:"phone_number" binding:"required"`
}

type UpdateUserRequestModel struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	UserName    string `json:"username" validate:"required"`
	PhoneNumber string `json:"phone_number"`
}

type LoginRequest struct {
	Email       string `json:"email" binding:"omitempty,email"`
	PhoneNumber string `json:"phone_number" binding:"omitempty"`
	Password    string `json:"password" binding:"required"`
	DeviceToken string `json:"device_token"`
	DeviceType  string `json:"device_type"`
}

type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" validate:""`
	NewPassword     string `json:"new_password" validate:"required,min=7"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=7,eqfield=NewPassword"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
	Token           string `json:"token" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=7"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=7"`
}

func (u *User) CreateUser(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(&u)

	if err != nil {
		return err
	}

	return nil
}

func (p *PasswordReset) CreatePasswordReset(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(&p)

	if err != nil {
		return err
	}

	return nil
}

func (pr *PasswordReset) GetPasswordResetByToken(db repository.DatabaseManager, token string) (PasswordReset, error) {
	var reset PasswordReset
	if err := db.DB().Where("token = ? AND expires_at > ?", token, time.Now()).First(&reset).Error; err != nil {
		return reset, err
	}
	return reset, nil
}

func (pr *PasswordReset) GetPasswordResetByEmail(db repository.DatabaseManager, email string) (*PasswordReset, error) {
	var reset PasswordReset
	if err := db.DB().Where("email = ?", email).First(&reset).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &reset, nil
}

func (pr *PasswordReset) DeletePasswordReset(db repository.DatabaseManager) error {

	err := db.DeleteRecordFromDb(pr)

	if err != nil {
		return err
	}

	return nil
}

type ResendVerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
	// optionally accept email for extra validation noise (not required)
	Email string `json:"email"`
}
