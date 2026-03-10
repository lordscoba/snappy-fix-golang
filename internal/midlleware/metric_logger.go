package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
)

func MetricsLogger(db *gorm.DB) gin.HandlerFunc {
	pdb := inst.InitDB(db)

	return func(c *gin.Context) {
		start := time.Now()

		// Use named route path if available, else raw URL path
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}
		method := c.Request.Method

		// Process request
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		metric := &entities.ApiRequestMetric{
			Path:       path,
			Method:     method,
			StatusCode: status,
			LatencyMs:  float64(duration.Microseconds()) / 1000.0, // ms
			IsError:    status >= 500,
		}

		// Fire-and-forget; if it fails, we don't block the request
		_ = metric.Create(pdb)
	}
}
