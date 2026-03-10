package jwtpkg

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/snappy-fix-golang/internal/config"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"gorm.io/gorm"
)

type TokenPair struct {
	AccessUuid       uuid.UUID
	AccessToken      string
	AccessExpiresAt  time.Time
	RefreshToken     string // raw (only returned to client)
	RefreshJTI       uuid.UUID
	RefreshExpiresAt time.Time
}

// ===== Helpers =====

func randToken(nBytes int) (string, error) {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// func accessTTL() time.Duration {
// 	cfg := config.GetConfig()
// 	// env: days
// 	d := cfg.Server.AccessTokenExpireDuration
// 	if d <= 0 {
// 		d = 1
// 	} // default 1 day
// 	return time.Hour * 24 * time.Duration(d)
// }

func accessTTL() time.Duration {
	cfg := config.GetConfig()

	// change to HOURS instead of DAYS
	h := cfg.Server.AccessTokenExpireDuration
	if h <= 0 {
		h = 3 // default 3 hours
	}

	return time.Hour * time.Duration(h)
}

// func refreshTTL() time.Duration {
// 	cfg := config.GetConfig()
// 	// add config.Server.RefreshTokenExpireDuration (days) in your config; default 30
// 	d := cfg.Server.RefreshTokenExpireDuration
// 	if d <= 0 {
// 		d = 30
// 	}
// 	return time.Hour * 24 * time.Duration(d)
// }

func refreshTTL() time.Duration {
	cfg := config.GetConfig()
	h := cfg.Server.RefreshTokenExpireDuration
	if h <= 0 {
		h = 3 // fallback default (3 hours)
	}
	return time.Hour * time.Duration(h)
}

// ===== User tokens =====

func CreateTokenPair(user entities.User) (*TokenPair, error) {
	cfg := config.GetConfig()

	tp := &TokenPair{
		AccessUuid: uuid.Must(uuid.NewV4()),
	}
	tp.AccessExpiresAt = time.Now().Add(accessTTL()).UTC()

	claims := jwt.MapClaims{
		"user_id":     user.ID.String(),
		"role":        string(user.Role), // "user" | "admin"
		"access_uuid": tp.AccessUuid.String(),
		"exp":         tp.AccessExpiresAt.Unix(),
		"authorised":  true,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	var err error
	tp.AccessToken, err = token.SignedString([]byte(cfg.Server.Secret))
	if err != nil {
		return nil, err
	}

	// refresh
	tp.RefreshJTI = uuid.Must(uuid.NewV4())
	tp.RefreshExpiresAt = time.Now().Add(refreshTTL()).UTC()
	if tp.RefreshToken, err = randToken(48); err != nil {
		return nil, err
	}

	return tp, nil
}

// Back-compat wrapper (returns only access details)
type TokenDetailDTO struct {
	AccessUuid  uuid.UUID `json:"access_uuid"`
	AccessToken string    `json:"access_token"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func CreateToken(user entities.User) (*TokenDetailDTO, error) {
	tp, err := CreateTokenPair(user)
	if err != nil {
		return nil, err
	}
	return &TokenDetailDTO{
		AccessUuid:  tp.AccessUuid,
		AccessToken: tp.AccessToken,
		ExpiresAt:   tp.AccessExpiresAt,
	}, nil
}

// ===== verification & helpers =====

func verifyToken(tokenString string) (*jwt.Token, error) {
	cfg := config.GetConfig()
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(cfg.Server.Secret), nil
	})
}

func IsTokenValid(bearerToken string) (*jwt.Token, error) {
	token, err := verifyToken(bearerToken)
	if err != nil {
		return token, fmt.Errorf("Unauthorized")
	}
	if !token.Valid {
		return nil, fmt.Errorf("Unauthorized")
	}
	return token, nil
}

func GetUserClaims(c *gin.Context, _ *gorm.DB, theValue string) (interface{}, error) {
	claims, exists := c.Get("userClaims")
	if !exists {
		return nil, errors.New("user claims not found")
	}
	mc, ok := claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}
	v, ok := mc[theValue]
	if !ok {
		return nil, errors.New("invalid value")
	}
	return v, nil
}

// GetUUIDClaim extracts a UUID from a jwt.MapClaims value by key (gofrs/uuid).
// Returns uuid.Nil and an error if missing or invalid.
func GetUUIDClaim(claims jwt.MapClaims, key string) (uuid.UUID, error) {
	v, ok := claims[key]
	if !ok {
		return uuid.Nil, fmt.Errorf("missing claim %q", key)
	}

	switch val := v.(type) {
	case string:
		return uuid.FromString(val) // gofrs
	case fmt.Stringer:
		return uuid.FromString(val.String())
	case []byte:
		return uuid.FromBytes(val)
	default:
		return uuid.Nil, fmt.Errorf("claim %q has type %T; want string/[]byte", key, v)
	}
}
