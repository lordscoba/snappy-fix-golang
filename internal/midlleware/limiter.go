package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/pkg/utils/responses"
	"golang.org/x/time/rate"
)

// clientLimiter holds the rate.Limiter and its last access time.
type clientLimiter struct {
	limiter      *rate.Limiter
	lastAccessed time.Time
}

var visitors = make(map[string]*clientLimiter)
var mu sync.Mutex

// inactivityTimeout defines how long a limiter can be inactive before being removed.
const inactivityTimeout = 5 * time.Minute

// getLimiter returns a rate limiter for a specific userID or IP.
// It also updates the lastAccessed timestamp for the limiter.
func getLimiter(key string) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	client, exists := visitors[key]
	if !exists {
		// Create a new limiter: 2 request per second, burst of 4.
		// Adjust these values (rate, burst) as needed for your application.
		limiter := rate.NewLimiter(10, 4)
		client = &clientLimiter{
			limiter:      limiter,
			lastAccessed: time.Now(),
		}
		visitors[key] = client
	} else {
		// Update the last accessed time for existing clients.
		client.lastAccessed = time.Now()
	}

	return client.limiter
}

// cleanupOldVisitors periodically removes inactive rate limiters from the visitors map.
// This function should be run as a goroutine.
func cleanupOldVisitors() {
	for {
		// Clean up every minute. Adjust the interval as needed.
		time.Sleep(1 * time.Minute)

		mu.Lock()
		for k, client := range visitors {
			// If the limiter hasn't been accessed for inactivityTimeout, delete it.
			if time.Since(client.lastAccessed) > inactivityTimeout {
				delete(visitors, k)
			}
		}
		mu.Unlock()
	}
}

// RateLimiter is a Gin middleware that applies rate limiting.
func RateLimiter() gin.HandlerFunc {
	// Start the cleanup goroutine once when the middleware is initialized.
	// This ensures it runs in the background.
	go cleanupOldVisitors()

	return func(c *gin.Context) {
		// Determine the key for rate limiting (user ID or IP address).
		var key string
		if userID, exists := c.Get("user_id"); exists {
			// Assuming "user_id" is set by an authentication middleware.
			key = userID.(string)
		} else {
			// Fallback to IP address if user ID is not available.
			key = c.ClientIP()
		}

		limiter := getLimiter(key)

		// Check if the request is allowed by the rate limiter.
		if limiter.Allow() {
			c.Next() // Request is allowed, proceed to the next handler.
		} else {
			// Rate limit exceeded, abort the request with a 429 Too Many Requests status.
			rd := responses.BuildErrorResponse(http.StatusTooManyRequests, "error", "Rate limit exceeded", nil, nil)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, rd)
		}
	}
}

// package middleware

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/hngprojects/hng_boilerplate_golang_web/utility"
// 	"golang.org/x/time/rate"
// )

// func RateLimiter() gin.HandlerFunc {
// 	limiter := rate.NewLimiter(1, 4)
// 	return func(c *gin.Context) {

// 		if limiter.Allow() {
// 			c.Next()
// 		} else {
// 			rd := utility.BuildErrorResponse(http.StatusTooManyRequests, "error", "Limit exceed", nil, nil)
// 			c.JSON(http.StatusTooManyRequests, rd)
// 		}

// 	}
// }
