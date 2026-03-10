package entities

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/adapters/repository"
)

type News struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Title      string    `gorm:"type:text;not null" json:"title"`
	Slug       string    `gorm:"size:500;uniqueIndex;not null" json:"slug"`
	Body       string    `gorm:"type:text;not null" json:"body"`
	CategoryID uuid.UUID `gorm:"type:uuid;not null;index" json:"category_id"`
	Category   Category  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"category"`

	// Main Thumbnail
	ThumbnailID  string `json:"thumbnail_id"` // Cloudinary PublicID
	ThumbnailURL string `json:"thumbnail_url"`

	// Multiple Images (Gallery/Inline)
	Images []NewsImage `gorm:"foreignKey:NewsID;constraint:OnDelete:CASCADE" json:"images"`

	Status      string    `gorm:"size:50;default:'draft';index" json:"status"`
	IsFeatured  bool      `gorm:"default:false;index" json:"is_featured"`
	IsExclusive bool      `gorm:"default:false;index" json:"is_exclusive"`
	Tags        string    `gorm:"type:text" json:"tags"`
	MetaTitle   string    `gorm:"type:text" json:"meta_title"`
	MetaDesc    string    `gorm:"type:text" json:"meta_desc"`
	PublishedAt time.Time `gorm:"index" json:"published_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type NewsImage struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	NewsID    uuid.UUID `gorm:"type:uuid;index" json:"news_id"`
	PublicID  string    `json:"public_id"`
	URL       string    `json:"url"`
	ImageType string    `json:"image_type"` // "thumbnail", "inline", "gallery"
}

// DTOs for Clean Data Handling
type CreateNewsRequest struct {
	Title       string `json:"title" form:"title" binding:"required"`
	Slug        string `json:"slug" form:"slug" binding:"required"`
	Body        string `json:"body" form:"body" binding:"required"`
	CategoryID  string `json:"category_id" form:"category_id" binding:"required,uuid"`
	Status      string `json:"status" form:"status"`
	IsFeatured  bool   `json:"is_featured" form:"is_featured"`
	IsExclusive bool   `json:"is_exclusive" form:"is_exclusive"`
	Tags        string `json:"tags" form:"tags"`
	MetaTitle   string `json:"meta_title" form:"meta_title"`
	MetaDesc    string `json:"meta_desc" form:"meta_desc"`
}

type UpdateNewsRequest struct {
	Title       *string    `json:"title"`
	Slug        *string    `json:"slug"`
	Body        *string    `json:"body"`
	CategoryID  *uuid.UUID `json:"category_id"`
	Status      *string    `json:"status"`
	IsFeatured  *bool      `json:"is_featured"`
	IsExclusive *bool      `json:"is_exclusive"`
	Tags        *string    `json:"tags"`
	MetaTitle   *string    `json:"meta_title"`
	MetaDesc    *string    `json:"meta_desc"`
}

//////////////////////////////////////////////////////
//// CREATE
//////////////////////////////////////////////////////

func (n *News) Create(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(n)
	if err != nil {
		return fmt.Errorf("failed to create news record: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////
//// UPDATE
//////////////////////////////////////////////////////

func (n *News) UpdateByID(
	db repository.DatabaseManager,
	updates map[string]interface{},
	id string,
) (*News, error) {

	result, err := db.UpdateFields(n, updates, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update news record: %w", err)
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("no news record found to update")
	}

	return n, nil
}

//////////////////////////////////////////////////////
//// GET BY ID
//////////////////////////////////////////////////////

// func (n *News) GetByID(db repository.DatabaseManager, id string) (News, error) {

// 	var out News

// 	err, selErr := db.SelectOneFromDb(&out, "id = ?", id)

// 	if selErr != nil {
// 		return out, fmt.Errorf("news record not found")
// 	}

// 	if err != nil {
// 		return out, fmt.Errorf("failed retrieving news by id: %w", err)
// 	}

// 	return out, nil
// }

func (n *News) GetByID(db repository.DatabaseManager, id string) (News, error) {

	var out News

	pg := db.(*postgresql.Postgresql)

	query := pg.PreloadEntities(nil, &News{}, "Category", "Images")

	err := query.Where("id = ?", id).First(&out).Error

	if err != nil {
		return out, err
	}

	return out, nil
}

//////////////////////////////////////////////////////
//// GET BY SLUG
//////////////////////////////////////////////////////

func (n *News) GetBySlug(db repository.DatabaseManager, slug string) (News, error) {

	var out News

	pg := db.(*postgresql.Postgresql)

	query := pg.PreloadEntities(nil, &News{}, "Category", "Images")

	err := query.Where("slug = ?", slug).First(&out).Error

	if err != nil {
		return out, err
	}

	return out, nil
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func (n *News) DeleteNews(db repository.DatabaseManager) error {

	err := db.DeleteRecordFromDb(n)

	if err != nil {
		return fmt.Errorf("failed to delete news record: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////
//// GET ALL NEWS (PAGINATED)
//////////////////////////////////////////////////////

func (n *News) GetAllNews(
	db repository.DatabaseManager,
	c *gin.Context,
) ([]News, repository.PaginationResponse, error) {

	var news []News

	pagination := postgresql.GetPagination(c)

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"published_at",
		"desc",
		"",
		pagination,
		&news,
		nil,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving news list: %w", err)
	}

	pg := db.(*postgresql.Postgresql)

	pg.PreloadEntities(nil, &News{}, "Category", "Images").Find(&news)

	return news, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// COUNT ALL NEWS
//////////////////////////////////////////////////////

func (n *News) CountAllNews(db repository.DatabaseManager) (int64, error) {

	count, err := db.CountRecords(&News{})

	if err != nil {
		return 0, fmt.Errorf("failed counting news records: %w", err)
	}

	return count, nil
}

//////////////////////////////////////////////////////
//// COUNT NEWS BY STATUS
//////////////////////////////////////////////////////

func (n *News) CountNewsByStatus(
	db repository.DatabaseManager,
	status string,
) (int64, error) {

	count, err := db.CountSpecificRecords(&News{}, "status = ?", status)

	if err != nil {
		return 0, fmt.Errorf("failed counting news with status '%s': %w", status, err)
	}

	return count, nil
}

//////////////////////////////////////////////////////
//// GET FEATURED NEWS
//////////////////////////////////////////////////////

func (n *News) GetFeaturedNews(
	db repository.DatabaseManager,
	c *gin.Context,
) ([]News, repository.PaginationResponse, error) {

	var news []News

	pagination := postgresql.GetPagination(c)

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"published_at",
		"desc",
		"is_featured = true AND status = 'published'",
		pagination,
		&news,
		nil,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving featured news: %w", err)
	}

	return news, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// GET EXCLUSIVE NEWS
//////////////////////////////////////////////////////

func (n *News) GetExclusiveNews(
	db repository.DatabaseManager,
	c *gin.Context,
) ([]News, repository.PaginationResponse, error) {

	var news []News

	pagination := postgresql.GetPagination(c)

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"published_at",
		"desc",
		"is_exclusive = true AND status = 'published'",
		pagination,
		&news,
		nil,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving exclusive news: %w", err)
	}

	return news, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// FILTER BY CATEGORY
//////////////////////////////////////////////////////

func (n *News) GetNewsByCategory(
	db repository.DatabaseManager,
	category string,
	c *gin.Context,
) ([]News, repository.PaginationResponse, error) {

	var news []News

	pagination := postgresql.GetPagination(c)

	condition := "category_id IN (SELECT id FROM categories WHERE slug = ?) AND status = 'published'"

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"published_at",
		"desc",
		condition,
		pagination,
		&news,
		[]interface{}{category},
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving news by category: %w", err)
	}

	return news, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// SEARCH NEWS
//////////////////////////////////////////////////////

func (n *News) SearchNews(
	db repository.DatabaseManager,
	query string,
	c *gin.Context,
) ([]News, repository.PaginationResponse, error) {

	var news []News

	pagination := postgresql.GetPagination(c)

	condition := "status = 'published' AND (title ILIKE ? OR body ILIKE ?)"

	search := "%" + query + "%"

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"published_at",
		"desc",
		condition,
		pagination,
		&news,
		[]interface{}{search, search},
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed searching news: %w", err)
	}

	return news, paginationResponse, nil
}
