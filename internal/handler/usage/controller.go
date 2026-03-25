package usage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	usageservice "github.com/snappy-fix-golang/internal/services/usage_service"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
}

//////////////////////////////////////////////////////
//// GET ALL LOGS
//////////////////////////////////////////////////////

func (base *Controller) GetAllLogs(c *gin.Context) {

	data, pagination, code, err := usageservice.GetAllLogsService(base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// COUNT ALL
//////////////////////////////////////////////////////

func (base *Controller) CountAllLogs(c *gin.Context) {

	data, code, err := usageservice.CountAllLogsService(base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// COUNT ERRORS
//////////////////////////////////////////////////////

func (base *Controller) CountErrors(c *gin.Context) {

	data, code, err := usageservice.CountErrorsService(base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// COUNT BY ACTION TYPE
//////////////////////////////////////////////////////

func (base *Controller) CountByActionType(c *gin.Context) {

	var param entities.ActionTypeParam

	if err := c.ShouldBindUri(&param); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "invalid action_type", err.Error(), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, code, err := usageservice.CountByActionTypeService(param.ActionType, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// AVG PROCESSING TIME
//////////////////////////////////////////////////////

func (base *Controller) AverageProcessingTime(c *gin.Context) {

	data, code, err := usageservice.AverageProcessingTimeService(base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// ERROR RATE
//////////////////////////////////////////////////////

func (base *Controller) ErrorRate(c *gin.Context) {

	data, code, err := usageservice.ErrorRateService(base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// GROUP BY ACTION
//////////////////////////////////////////////////////

func (base *Controller) GroupByActionType(c *gin.Context) {

	data, code, err := usageservice.GroupByActionTypeService(base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}
