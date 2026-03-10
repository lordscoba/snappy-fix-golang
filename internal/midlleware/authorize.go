package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/domain/enums"
	"github.com/snappy-fix-golang/internal/inst"
	jwtpkg "github.com/snappy-fix-golang/pkg/jwt"
	"github.com/snappy-fix-golang/pkg/utils/responses"
	"gorm.io/gorm"
)

func Authorize(db *gorm.DB, allowedRoles ...enums.UserType) gin.HandlerFunc {
	d := inst.InitDB(db)
	return func(c *gin.Context) {
		var tokenStr string
		if parts := strings.Split(c.GetHeader("Authorization"), " "); len(parts) == 2 {
			tokenStr = parts[1]
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token could not be found!", "Unauthorized", nil))
			return
		}

		token, err := jwtpkg.IsTokenValid(tokenStr)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token is invalid!", "Unauthorized", nil))
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token is invalid!", "Unauthorized", nil))
			return
		}

		userID, _ := claims["user_id"].(string)
		accessID, _ := claims["access_uuid"].(string)
		if userID == "" || accessID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token is invalid!", "Unauthorized", nil))
			return
		}

		// session check
		accUUID, err := uuid.FromString(accessID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token is invalid!", "Unauthorized", nil))
			return
		}
		var session entities.AccessToken
		session = entities.AccessToken{ID: accUUID}
		if code, err := session.GetByID(d); err != nil {
			c.AbortWithStatusJSON(code, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token is invalid!", "Unauthorized", nil))
			return
		}

		roleStr, _ := claims["role"].(string) // "USER" | "ADMIN"
		if len(allowedRoles) > 0 {
			okRole := false
			for _, r := range allowedRoles {
				if roleStr == string(r) {
					okRole = true
					break
				}
			}
			if !okRole {
				c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(
					http.StatusUnauthorized, "error", "role not authorized!", "Unauthorized", nil))
				return
			}
		}

		authorised, _ := claims["authorised"].(bool)
		if !authorised {
			c.AbortWithStatusJSON(http.StatusUnauthorized, responses.BuildErrorResponse(http.StatusUnauthorized, "error", "status not authorized!", "Unauthorized", nil))
			return
		}

		c.Set("userClaims", claims)
		c.Next()
	}
}

func GetIdFromToken(c *gin.Context) (string, interface{}) {
	var tokenStr string
	bearerToken := c.GetHeader("Authorization")
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		tokenStr = strArr[1]
	}

	if tokenStr == "" {
		r := responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token could not be found!", "Unauthorized", nil)
		return "", r
	}

	token, err := jwtpkg.IsTokenValid(tokenStr)
	if err != nil {
		r := responses.BuildErrorResponse(http.StatusUnauthorized, "error", "Token is invalid!", "Unauthorized", nil)
		return "", r
	}

	// access user claims

	claims := token.Claims.(jwt.MapClaims)
	id, ok := claims["user_id"].(string)
	if !ok {
		return "", responses.BuildErrorResponse(http.StatusForbidden, "error", "Forbidden", "Unauthorized", nil)
	}
	return id, ""
}
