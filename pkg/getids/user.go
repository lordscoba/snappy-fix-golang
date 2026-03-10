package getids

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
)

func UserIDFromCtx(c *gin.Context) (uuid.UUID, error) {
	// Step 1: extract claims from context

	claimsAny, exists := c.Get("userClaims")
	if !exists {
		return uuid.Nil, errors.New("no claims found in context")
	}

	claims := claimsAny.(jwt.MapClaims)

	// Step 2: try user_id
	var user_raw string
	if v, ok := claims["user_id"].(string); ok {
		user_raw = v
	}

	if strings.TrimSpace(user_raw) == "" {
		return uuid.Nil, errors.New("no user id in claims")
	}

	// Step 3: parse UUID
	userId, err := uuid.FromString(strings.TrimSpace(user_raw))
	if err != nil || userId == uuid.Nil {
		return uuid.Nil, errors.New("invalid user id format in token")
	}

	return userId, nil
}
