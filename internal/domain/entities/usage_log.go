package entities

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/enums"
)

type UsageLog struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id"`

	// ─── Core tracking ─────────────────────────────
	ActionType enums.ActionType `gorm:"type:varchar(50);index;not null" json:"action_type"`
	Endpoint   *string          `gorm:"type:varchar(255)" json:"endpoint,omitempty"`
	Method     *string          `gorm:"type:varchar(10)" json:"method,omitempty"`

	// ─── Request info ──────────────────────────────
	IPAddress *string `gorm:"type:varchar(45);index" json:"ip_address,omitempty"`
	UserAgent *string `gorm:"type:varchar(500)" json:"user_agent,omitempty"`
	RequestID *string `gorm:"type:varchar(100);index" json:"request_id,omitempty"`
	Country   *string `gorm:"type:varchar(40)" json:"country,omitempty"`

	// ─── File metadata ─────────────────────────────
	FileSize       *int    `json:"file_size,omitempty"`
	OriginalFormat *string `gorm:"type:varchar(20)" json:"original_format,omitempty"`
	TargetFormat   *string `gorm:"type:varchar(20)" json:"target_format,omitempty"`
	ImageWidth     *int    `json:"image_width,omitempty"`
	ImageHeight    *int    `json:"image_height,omitempty"`

	// ─── Result tracking ───────────────────────────
	Success      bool    `gorm:"default:true;index" json:"success"`
	StatusCode   *int    `json:"status_code,omitempty"`
	ErrorType    *string `gorm:"type:varchar(255)" json:"error_type,omitempty"`
	ErrorMessage *string `gorm:"type:text" json:"error_message,omitempty"`

	// ─── Performance tracking ──────────────────────
	ProcessingTimeMs *int `json:"processing_time_ms,omitempty"`

	// ─── Timestamp ─────────────────────────────────
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
}

type ActionTypeParam struct {
	ActionType string `uri:"action_type" validate:"required"`
}

func (u *UsageLog) Create(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(u)
	if err != nil {
		return fmt.Errorf("failed to create usage log: %w", err)
	}

	return nil
}

func (u *UsageLog) GetAllLogs(
	db repository.DatabaseManager,
	ctx *gin.Context,
) ([]UsageLog, repository.PaginationResponse, error) {

	var logs []UsageLog
	pagination := postgresql.GetPagination(ctx)

	// ─── Query params ─────────────────────────────
	search := strings.TrimSpace(ctx.Query("search"))
	actionType := ctx.Query("action_type")
	success := ctx.Query("success")
	method := ctx.Query("method")
	fromDate := ctx.Query("from_date")
	toDate := ctx.Query("to_date")

	var conditions []string
	var args []interface{}

	// ─── Search (text fields) ─────────────────────
	if search != "" {
		q := "%" + strings.ToLower(search) + "%"
		conditions = append(conditions,
			"(LOWER(endpoint) ILIKE ? OR LOWER(error_message) ILIKE ? OR LOWER(user_agent) ILIKE ?)",
		)
		args = append(args, q, q, q)
	}

	// ─── Filter: action type ──────────────────────
	if actionType != "" {
		conditions = append(conditions, "action_type = ?")
		args = append(args, actionType)
	}

	// ─── Filter: success ──────────────────────────
	if success != "" {
		val := success == "true"
		conditions = append(conditions, "success = ?")
		args = append(args, val)
	}

	// ─── Filter: method ───────────────────────────
	if method != "" {
		conditions = append(conditions, "method = ?")
		args = append(args, method)
	}

	// ─── Filter: date range ───────────────────────
	if fromDate != "" {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, fromDate)
	}
	if toDate != "" {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, toDate)
	}

	var query interface{} = "1=1"
	if len(conditions) > 0 {
		query = strings.Join(conditions, " AND ")
	}

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"",
		pagination,
		&logs,
		query,
		args...,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving logs: %w", err)
	}

	return logs, paginationResponse, nil
}

func (u *UsageLog) CountAllLogs(db repository.DatabaseManager) (int64, error) {

	count, err := db.CountRecords(&UsageLog{})
	if err != nil {
		return 0, fmt.Errorf("failed counting logs: %w", err)
	}

	return count, nil
}

func (u *UsageLog) CountByActionType(
	db repository.DatabaseManager,
	actionType string,
) (int64, error) {

	count, err := db.CountSpecificRecords(
		&UsageLog{},
		"action_type = ?",
		actionType,
	)

	if err != nil {
		return 0, fmt.Errorf("failed counting logs for action '%s': %w", actionType, err)
	}

	return count, nil
}

func (u *UsageLog) CountErrors(db repository.DatabaseManager) (int64, error) {

	count, err := db.CountSpecificRecords(
		&UsageLog{},
		"success = ?",
		false,
	)

	if err != nil {
		return 0, fmt.Errorf("failed counting error logs: %w", err)
	}

	return count, nil
}

func (u *UsageLog) AverageProcessingTime(db repository.DatabaseManager) (float64, error) {

	var avg float64

	err := db.DB().
		Model(&UsageLog{}).
		Select("AVG(processing_time_ms)").
		Scan(&avg).Error

	if err != nil {
		return 0, fmt.Errorf("failed calculating average processing time: %w", err)
	}

	return avg, nil
}

type ActionCount struct {
	ActionType string
	Count      int64
}

func (u *UsageLog) GroupByActionType(db repository.DatabaseManager) ([]ActionCount, error) {

	var result []ActionCount

	err := db.DB().
		Model(&UsageLog{}).
		Select("action_type, COUNT(*) as count").
		Group("action_type").
		Scan(&result).Error

	if err != nil {
		return nil, fmt.Errorf("failed grouping logs: %w", err)
	}

	return result, nil
}

func (u *UsageLog) ErrorRate(db repository.DatabaseManager) (float64, error) {

	var total int64
	var errors int64

	db.DB().Model(&UsageLog{}).Count(&total)
	db.DB().Model(&UsageLog{}).Where("success = ?", false).Count(&errors)

	if total == 0 {
		return 0, nil
	}

	return (float64(errors) / float64(total)) * 100, nil
}
