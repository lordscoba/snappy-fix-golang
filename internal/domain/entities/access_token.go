// internal/domain/entities/access_token.go
package entities

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"golang.org/x/crypto/bcrypt"
)

type AccessToken struct {
	ID      uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey;unique" json:"id"`
	Email   string    ` json:"email"`
	OwnerID uuid.UUID `gorm:"type:uuid;index;not null" json:"owner_id"`
	// User                      User            `gorm:"constraint:OnDelete:CASCADE;"`
	IsLive                    bool   `gorm:"default:false;not null" json:"is_live"`
	LoginAccessToken          string `gorm:"column:login_access_token; type:text" json:"-"`
	LoginAccessTokenExpiresIn string `gorm:"column:login_access_token_expires_in; type:varchar(250)" json:"-"`

	// NEW: Refresh support
	RefreshTokenHash      string    `gorm:"column:refresh_token_hash; type:text" json:"-"`
	RefreshTokenExpiresAt time.Time `gorm:"column:refresh_token_expires_at" json:"-"`
	RefreshJTI            uuid.UUID `gorm:"type:uuid;index" json:"-"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// device token for FCM for managing push notifications for more than one device
	DeviceToken string `gorm:"column:device_token; type:text" json:"device_token"`
	DeviceType  string `gorm:"column:device_type; type:text" json:"device_type"`
}

// Strongly-typed creator payload (recommended)
type TokenStoreInput struct {
	OwnerID uuid.UUID

	AccessUUID        uuid.UUID
	AccessToken       string
	AccessExpiresUnix int64 // seconds

	RefreshJTI         uuid.UUID
	RefreshTokenPlain  string // will be hashed before save
	RefreshExpiresUnix int64  // seconds

	DeviceToken string
	DeviceType  string
}

func (a *AccessToken) GetByOwnerID(db repository.DatabaseManager) (int, error) {
	err, nilErr := db.SelectOneFromDb(&a, "owner_id = ? ", a.OwnerID)
	if nilErr != nil {
		return http.StatusBadRequest, nilErr
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (a *AccessToken) GetByID(db repository.DatabaseManager) (int, error) {
	err, nilErr := db.SelectOneFromDb(&a, "id = ? ", a.ID)
	if nilErr != nil {
		return http.StatusBadRequest, nilErr
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// GetByOwner: fetch *any* one token row for the given owner (owner_id + owner_type).
func (a *AccessToken) GetByOwner(db repository.DatabaseManager) (int, error) {
	if a.OwnerID == uuid.Nil {
		return http.StatusBadRequest, errors.New("missing owner id/type")
	}
	err, nilErr := db.SelectOneFromDb(&a, "owner_id = ? and owner_type = ? ", a.OwnerID)
	if nilErr != nil {
		return http.StatusBadRequest, nilErr
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

func (a *AccessToken) GetLatestByOwnerIDAndIsLive(db repository.DatabaseManager) (int, error) {
	err, nilErr := db.SelectLatestFromDb(&a, "owner_id = ? and is_live = ? ", a.OwnerID, a.IsLive)
	if nilErr != nil {
		return http.StatusBadRequest, nilErr
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// GetLatestLiveByOwner: latest token row for an owner that is live.
func (a *AccessToken) GetLatestLiveByOwner(db repository.DatabaseManager) (int, error) {
	if a.OwnerID == uuid.Nil {
		return 400, errors.New("missing owner id/type")
	}
	a.IsLive = true
	err, nilErr := db.SelectLatestFromDb(&a, "owner_id = ? AND owner_type = ? AND is_live = ?", a.OwnerID, true)
	if nilErr != nil {
		return 400, nilErr
	}
	if err != nil {
		return 500, err
	}
	return 200, nil
}

// NEW: find by RefreshJTI
func (a *AccessToken) GetByRefreshJTI(db repository.DatabaseManager, jti uuid.UUID) (int, error) {
	err, nilErr := db.SelectOneFromDb(&a,
		"refresh_jti = ? AND is_live = ?",
		jti, true,
	)
	if nilErr != nil {
		return http.StatusBadRequest, nilErr
	}
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

// Hash helper
func hashRefreshToken(raw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	return string(b), err
}

// Extend to store refresh data if provided
// tokenData keys supported:
// - "access_token", "exp", "access_exp_unix"
// - "refresh_token", "refresh_exp_unix", "refresh_jti"
func (a *AccessToken) CreateAccessToken(db repository.DatabaseManager, tokenData interface{}) error {
	if a.OwnerID == uuid.Nil {
		return fmt.Errorf("owner id not provided to create access token")
	}
	if a.ID == uuid.Nil {
		return fmt.Errorf("access id not provided to create access token")
	}

	m := tokenData.(map[string]string)

	a.IsLive = true
	a.LoginAccessToken = m["access_token"]
	if exp, ok := m["exp"]; ok {
		a.LoginAccessTokenExpiresIn = exp
	}

	// Optional refresh fields
	if rt, ok := m["refresh_token"]; ok && rt != "" {
		h, err := hashRefreshToken(rt)
		if err != nil {
			return fmt.Errorf("failed to hash refresh token: %v", err)
		}
		a.RefreshTokenHash = h
	}
	if ts, ok := m["refresh_exp_unix"]; ok && ts != "" {
		// store as time; ts is Unix string
		secs, _ := time.ParseDuration(ts + "s") // dummy; not ideal—see below
		_ = secs
		// Better: parse int
		if unix, perr := parseUnix(ts); perr == nil {
			a.RefreshTokenExpiresAt = time.Unix(unix, 0).UTC()
		}
	}
	if jtiStr, ok := m["refresh_jti"]; ok && jtiStr != "" {
		if jti, err := uuid.FromString(jtiStr); err == nil {
			a.RefreshJTI = jti
		}
	}

	if err := db.CreateOneRecord(&a); err != nil {
		return fmt.Errorf("access token create failed: %v", err)
	}
	return nil
}

func (a *AccessToken) RevokeAccessToken(db repository.DatabaseManager) error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("access token id not provided to revoke access token")
	}
	a.IsLive = false
	_, err := db.SaveAllFields(&a)
	return err
}

// CreateFromPair: preferred typed creator.
func (a *AccessToken) CreateFromPair(db repository.DatabaseManager, in TokenStoreInput) error {
	if in.OwnerID == uuid.Nil {
		return errors.New("owner id/type required")
	}
	if in.AccessUUID == uuid.Nil || in.AccessToken == "" {
		return errors.New("access uuid/token required")
	}

	a.ID = in.AccessUUID
	a.OwnerID = in.OwnerID
	a.IsLive = true

	a.LoginAccessToken = in.AccessToken
	a.LoginAccessTokenExpiresIn = strconv.FormatInt(in.AccessExpiresUnix, 10)

	if in.RefreshTokenPlain != "" {
		h, err := bcrypt.GenerateFromPassword([]byte(in.RefreshTokenPlain), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		a.RefreshTokenHash = string(h)
	}
	if in.RefreshExpiresUnix > 0 {
		a.RefreshTokenExpiresAt = time.Unix(in.RefreshExpiresUnix, 0).UTC()
	}
	a.RefreshJTI = in.RefreshJTI
	a.DeviceToken = in.DeviceToken
	a.DeviceType = in.DeviceType

	return db.CreateOneRecord(&a)
}

func parseUnix(s string) (int64, error) {
	var unix int64
	_, err := fmt.Sscan(s, &unix)
	return unix, err
}

// CheckRefreshToken compares a plaintext refresh token with the stored hash.
func (a *AccessToken) CheckRefreshToken(plain string) error {
	if a.RefreshTokenHash == "" {
		return errors.New("no refresh token hash")
	}
	return bcrypt.CompareHashAndPassword([]byte(a.RefreshTokenHash), []byte(plain))
}
