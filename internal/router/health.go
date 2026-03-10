package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"

	"github.com/snappy-fix-golang/internal/adapters/db"
	httphealth "github.com/snappy-fix-golang/internal/handler/http_health"

	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func HttpHealth(r *gin.Engine, ApiVersion string, validator *validator.Validate, db *db.Database, logger *logutil.Logger) *gin.Engine {
	extReq := request.ExternalRequest{Logger: logger, Test: false}
	health := httphealth.Controller{Db: db, Validator: validator, Logger: logger, ExtReq: extReq}

	healthUrl := r.Group(fmt.Sprintf("%v", ApiVersion))
	{
		healthUrl.POST("/health", health.Post)
		healthUrl.GET("/health", health.Get)
	}

	return r
}
