package entities

import (
	"errors"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"gorm.io/gorm"
)

type EmailVerification struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey;unique;not null" json:"id"`
	Email     string    `gorm:"index;not null" json:"email"`
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	Consumed  bool      `gorm:"default:false;not null" json:"consumed"`

	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (v *EmailVerification) Create(db repository.DatabaseManager) error {
	v.Email = strings.ToLower(strings.TrimSpace(v.Email))
	if v.Email == "" || v.Token == "" {
		return errors.New("invalid verification payload")
	}
	return db.CreateOneRecord(&v)
}

func (v *EmailVerification) GetByToken(db repository.DatabaseManager, token string) (EmailVerification, error) {
	var out EmailVerification
	if err := db.DB().
		Where("token = ? AND consumed = false AND expires_at > now()", token).
		First(&out).Error; err != nil {
		return out, err
	}
	return out, nil
}

func (v *EmailVerification) MarkConsumed(db repository.DatabaseManager) error {
	v.Consumed = true
	_, err := db.SaveAllFields(&v)
	return err
}
