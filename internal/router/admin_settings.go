package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	adminmetricssettings "github.com/snappy-fix-golang/internal/handler/admin"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func AdminSettings(r *gin.Engine, ApiVersion string, validator *validator.Validate, db *db.Database, logger *logutil.Logger) *gin.Engine {
	extReq := request.ExternalRequest{Logger: logger, Test: false}
	providers := adminmetricssettings.Controller{Db: db, Validator: validator, Logger: logger, ExtReq: extReq}

	adminUrl := r.Group(fmt.Sprintf("%v/admin", ApiVersion), middleware.Authorize(db.Postgresql.DB()))
	{

		adminUrl.GET("/settings/metrics", providers.GetPlatformMetricsHandler)
	}

	return r
}
