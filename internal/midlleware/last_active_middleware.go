package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/pkg/getids"
	"gorm.io/gorm"
)

// Prefer proxy headers if you're behind Cloudflare/ALB/NGINX.
func clientIP(c *gin.Context) *string {
	// Cloudflare
	if h := c.GetHeader("CF-Connecting-IP"); h != "" {
		v := strings.TrimSpace(h)
		return &v
	}

	// NGINX / Proxy
	if h := c.GetHeader("X-Real-IP"); h != "" {
		v := strings.TrimSpace(h)
		return &v
	}

	if h := c.GetHeader("X-Forwarded-For"); h != "" {
		parts := strings.Split(h, ",")
		// take first non-private entry
		for _, p := range parts {
			v := strings.TrimSpace(p)
			if !strings.HasPrefix(v, "10.") &&
				!strings.HasPrefix(v, "192.168.") &&
				!strings.HasPrefix(v, "172.") {
				return &v
			}
		}
		v := strings.TrimSpace(parts[0])
		return &v
	}

	v := c.ClientIP()
	return &v
}

// Trim UA to something reasonable to avoid giant headers filling the DB.
func userAgent(c *gin.Context) string {
	ua := strings.TrimSpace(c.Request.UserAgent())
	if len(ua) > 512 {
		return ua[:512]
	}
	return ua
}

// Call this right after a successful login to stamp initial activity.
// NOTE: only for ProviderUser in this example; mirror for User if you have user login.
func OnSuccessfulProviderUserLogin(gdb *gorm.DB, c *gin.Context, puID string) error {
	ip := clientIP(c)
	ua := userAgent(c)
	// If you also have last_login_at column, include it here.
	return gdb.Exec(`
		UPDATE provider_users
		   SET last_active_at = NOW(),
		       last_ip        = ?,
		       last_user_agent= ?
		 WHERE id = ?
	`, ip, ua, puID).Error
}

func UserActiveTracker(gdb *gorm.DB, minInterval time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		userID, err := getids.UserIDFromCtx(c)
		if err != nil || userID.IsNil() {
			return // not a patient/user route or unauthenticated
		}

		ip := clientIP(c)
		ua := userAgent(c)
		seconds := int(minInterval.Seconds())
		gdb.Exec(`
			UPDATE users
			   SET last_active_at = NOW(),
			       last_ip        = ?,
			       last_user_agent= ?
			 WHERE id = ?
			   AND (
			        last_active_at IS NULL
			     OR last_active_at < NOW() - make_interval(secs := ?)
			   )
		`, ip, ua, userID, seconds)
	}
}
