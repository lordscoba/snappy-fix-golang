package usageservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	"gorm.io/gorm"
)

//////////////////////////////////////////////////////
//// GET ALL LOGS
//////////////////////////////////////////////////////

func GetAllLogsService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	list, pagination, err := log.GetAllLogs(pdb, c)
	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"logs": list}, pagination, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// COUNT ALL
//////////////////////////////////////////////////////

func CountAllLogsService(db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	count, err := log.CountAllLogs(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"total": count}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// COUNT ERRORS
//////////////////////////////////////////////////////

func CountErrorsService(db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	count, err := log.CountErrors(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"errors": count}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// COUNT BY ACTION TYPE
//////////////////////////////////////////////////////

func CountByActionTypeService(actionType string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	count, err := log.CountByActionType(pdb, actionType)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{
		"action_type": actionType,
		"count":       count,
	}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// AVERAGE PROCESSING TIME
//////////////////////////////////////////////////////

func AverageProcessingTimeService(db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	avg, err := log.AverageProcessingTime(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"average_processing_time_ms": avg}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// ERROR RATE
//////////////////////////////////////////////////////

func ErrorRateService(db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	rate, err := log.ErrorRate(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"error_rate_percent": rate}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GROUP BY ACTION TYPE
//////////////////////////////////////////////////////

func GroupByActionTypeService(db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var log entities.UsageLog

	result, err := log.GroupByActionType(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"actions": result}, http.StatusOK, nil
}
