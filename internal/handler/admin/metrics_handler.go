package adminmetricssettings

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	adminmetricsservices "github.com/snappy-fix-golang/internal/services/admin"
	logutil "github.com/snappy-fix-golang/pkg/logger"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
	Logger    *logutil.Logger
	ExtReq    request.ExternalRequest
}

// GetPlatformMetricsHandler godoc
//
//	@Summary      Get platform uptime & SLA metrics (last N days)
//	@Description  Returns aggregated uptime, latency, error rate, and downtime incidents compared against SLA targets.
//	@Tags         Admin Settings
//	@Accept       json
//	@Produce      json
//	@Param        Authorization header string true "Bearer token"
//	@Param        window_days   query  int    false "Aggregation window in days (default 30)"
//	@Success      200 {object}  responses.Response
//	@Failure      400 {object}  responses.Response
//	@Router       /admin/settings/metrics [get]
func (base *Controller) GetPlatformMetricsHandler(c *gin.Context) {
	windowDaysStr := c.DefaultQuery("window_days", "30")
	windowDays, _ := strconv.Atoi(windowDaysStr)

	data, code, err := adminmetricsservices.GetPlatformMetricsService(base.Db.Postgresql.DB(), windowDays)
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", "failed to fetch platform metrics", err, nil)
		c.JSON(code, rd)
		return
	}

	base.Logger.Info("Platform metrics fetched successfully.")
	// No pagination/meta needed; pass nil
	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, http.StatusOK)
	c.JSON(http.StatusOK, rd)
}
