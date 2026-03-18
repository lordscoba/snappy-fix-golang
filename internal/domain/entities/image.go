package entities

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/adapters/repository"
)

type NewsImage struct {
	ID uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`

	// Make nullable using pointer
	NewsID *uuid.UUID `gorm:"type:uuid;index" json:"news_id,omitempty"`

	PublicID string `gorm:"type:varchar(255);not null" json:"public_id"`
	URL      string `gorm:"type:text;not null" json:"url"`

	ImageType string `gorm:"type:varchar(50)" json:"image_type"` // "thumbnail", "inline", "gallery"

	// New fields
	FileName    string  `gorm:"type:varchar(255)" json:"file_name"`
	Extension   string  `gorm:"type:varchar(20)" json:"extension"` // e.g. jpg, png, webp
	Size        int64   `gorm:"type:bigint" json:"size"`           // size in bytes
	Description *string `gorm:"type:text" json:"description,omitempty"`

	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
}

type CreateNewsImageRequest struct {
	NewsID      *string `json:"news_id" form:"news_id"` // optional
	ImageType   string  `json:"image_type" form:"image_type"`
	FileName    string  `json:"file_name" form:"file_name"`
	Extension   string  `json:"extension" form:"extension"`
	Description *string `json:"description" form:"description"`
}

//////////////////////////////////////////////////////
//// CREATE
//////////////////////////////////////////////////////

func (n *NewsImage) Create(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(n)
	if err != nil {
		return fmt.Errorf("failed to create news image record: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////
//// GET BY ID
//////////////////////////////////////////////////////

func (n *NewsImage) GetByID(db repository.DatabaseManager, id string) (NewsImage, error) {

	var out NewsImage

	err, selErr := db.SelectOneFromDb(&out, "id = ?", id)

	if selErr != nil {
		return out, fmt.Errorf("news image not found")
	}

	if err != nil {
		return out, fmt.Errorf("failed retrieving news image by id: %w", err)
	}

	return out, nil
}

//////////////////////////////////////////////////////
//// GET BY PUBLIC ID
//////////////////////////////////////////////////////

func (n *NewsImage) GetByPublicID(db repository.DatabaseManager, publicID string) (NewsImage, error) {

	var out NewsImage

	err, selErr := db.SelectOneFromDb(&out, "public_id = ?", publicID)

	if selErr != nil {
		return out, fmt.Errorf("news image with public_id '%s' not found", publicID)
	}

	if err != nil {
		return out, fmt.Errorf("failed retrieving news image by public_id: %w", err)
	}

	return out, nil
}

//////////////////////////////////////////////////////
//// GET BY NEWS ID (FOR ATTACHED IMAGES)
//////////////////////////////////////////////////////

func (n *NewsImage) GetByNewsID(
	db repository.DatabaseManager,
	newsID string,
) ([]NewsImage, error) {

	var images []NewsImage

	err := db.SelectAllFromDb(
		"created_at",  // orderBy
		"desc",        // order
		&images,       // receiver
		"news_id = ?", // query
		newsID,        // args
	)

	if err != nil {
		return nil, fmt.Errorf("failed retrieving images by news id: %w", err)
	}

	return images, nil
}

//////////////////////////////////////////////////////
//// GET ALL IMAGES (PAGINATED + SEARCH)
//////////////////////////////////////////////////////

func (n *NewsImage) GetAllImages(
	db repository.DatabaseManager,
	g *gin.Context,
) ([]NewsImage, repository.PaginationResponse, error) {

	var images []NewsImage

	pagination := postgresql.GetPagination(g)

	// 🔍 Search query param
	search := g.Query("search")

	var query interface{} = ""
	var args []interface{}

	if search != "" {
		query = `
			file_name ILIKE ? OR 
			extension ILIKE ? OR 
			description ILIKE ?
		`
		searchTerm := "%" + search + "%"
		args = []interface{}{searchTerm, searchTerm, searchTerm}
	}

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"", // filter (optional for future use)
		pagination,
		&images,
		query,
		args...,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving images list: %w", err)
	}

	return images, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func (n *NewsImage) DeleteNewsImage(db repository.DatabaseManager) error {

	err := db.DeleteRecordFromDb(n)

	if err != nil {
		return fmt.Errorf("failed to delete news image record: %w", err)
	}

	return nil
}
