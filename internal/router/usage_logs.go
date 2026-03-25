package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/handler/usage"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
)

func UsageLog(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	controller := usage.Controller{
		Db:        db,
		Validator: validator,
	}

	logs := r.Group(fmt.Sprintf("%v/usage-logs", apiVersion), middleware.Authorize(db.Postgresql.DB()))
	{
		// Logs listing + filtering
		logs.GET("", controller.GetAllLogs)

		// Stats / analytics
		logs.GET("/count/all", controller.CountAllLogs)
		logs.GET("/count/errors", controller.CountErrors)
		logs.GET("/count/action/:action_type", controller.CountByActionType)

		logs.GET("/stats/average-processing-time", controller.AverageProcessingTime)
		logs.GET("/stats/error-rate", controller.ErrorRate)
		logs.GET("/stats/group-by-action", controller.GroupByActionType)
	}

	return r
}
