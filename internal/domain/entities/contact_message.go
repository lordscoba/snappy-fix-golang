package entities

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/enums"
)

type ContactMessage struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	Category enums.ContactCategory `gorm:"type:varchar(50);check:category IN ('website_development','online_tools_feedback','general_enquiry','partnership')"`
	// e.g: "website_development", "feedback", "general_enquiry", "partnership"

	Name  string `gorm:"type:varchar(150);not null" json:"name"`
	Email string `gorm:"type:varchar(150);index;not null" json:"email"`

	Subject string `gorm:"type:varchar(255);not null" json:"subject"`
	Message string `gorm:"type:text;not null" json:"message"`

	// Metadata (very useful in production)
	IP        string `gorm:"type:varchar(45);index" json:"ip"`
	UserAgent string `gorm:"type:text" json:"user_agent"`

	// Status handling (for admin workflow)
	IsRead bool `gorm:"default:false;index" json:"is_read"`

	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

type CreateContactMessageRequest struct {
	Category enums.ContactCategory `json:"category" validate:"required"`
	Name     string                `json:"name" validate:"required,min=2"`
	Email    string                `json:"email" validate:"required,email"`
	Subject  string                `json:"subject" validate:"required"`
	Message  string                `json:"message" validate:"required"`
}

func (c *ContactMessage) Create(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(c)
	if err != nil {
		return fmt.Errorf("failed to create contact message: %w", err)
	}

	return nil
}

func (c *ContactMessage) GetByID(
	db repository.DatabaseManager,
	id string,
) (ContactMessage, error) {

	var out ContactMessage

	err, selectErr := db.SelectOneFromDb(&out, "id = ?", id)
	if selectErr != nil {
		return out, err
	}

	return out, err
}

func (c *ContactMessage) MarkAsRead(
	db repository.DatabaseManager,
	id string,
) error {

	updates := map[string]interface{}{
		"is_read": true,
	}

	result, err := db.UpdateFields(c, updates, id)
	if err != nil {
		return fmt.Errorf("failed to mark message as read: %w", err)
	}

	if result.RowsAffected == 0 {
		return errors.New("no message found to update")
	}

	return nil
}

func (c *ContactMessage) Delete(
	db repository.DatabaseManager,
) error {

	err := db.DeleteRecordFromDb(c)
	if err != nil {
		return fmt.Errorf("failed to delete contact message: %w", err)
	}

	return nil
}

func (c *ContactMessage) CountAll(
	db repository.DatabaseManager,
) (int64, error) {

	count, err := db.CountRecords(&ContactMessage{})
	if err != nil {
		return 0, fmt.Errorf("failed counting contact messages: %w", err)
	}

	return count, nil
}

// ?search="hello"&category="website_development"&is_read="true"
func (c *ContactMessage) GetAll(db repository.DatabaseManager, ctx *gin.Context) ([]ContactMessage, repository.PaginationResponse, error) {

	var messages []ContactMessage
	pagination := postgresql.GetPagination(ctx)

	// ─── Query params ───────────────────────────────
	isRead := ctx.Query("is_read")
	category := ctx.Query("category")
	search := strings.TrimSpace(ctx.Query("search"))

	var conditions []string
	var args []interface{}

	// ─── Filter: is_read ────────────────────────────
	if isRead != "" {
		val := isRead == "true"
		conditions = append(conditions, "is_read = ?")
		args = append(args, val)
	}

	// ─── Filter: category ───────────────────────────
	if category != "" {
		conditions = append(conditions, "category = ?")
		args = append(args, category)
	}

	// ─── Search (multi-column) ──────────────────────
	if search != "" {
		q := "%" + strings.ToLower(search) + "%"
		conditions = append(conditions,
			"(LOWER(name) ILIKE ? OR LOWER(email) ILIKE ? OR LOWER(subject) ILIKE ? OR LOWER(message) ILIKE ?)",
		)
		args = append(args, q, q, q, q)
	}

	// ─── Build final WHERE ──────────────────────────
	var query interface{} = "1=1"
	if len(conditions) > 0 {
		query = strings.Join(conditions, " AND ")
	}

	// ─── Execute ────────────────────────────────────
	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"",
		pagination,
		&messages,
		query,
		args...,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving messages: %w", err)
	}

	return messages, paginationResponse, nil
}
