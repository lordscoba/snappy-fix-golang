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

// DTOs for Clean Data Handling
type CreateNewsRequest struct {
	Title       string `json:"title" form:"title" binding:"required"`
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
	Title       *string    `json:"title" form:"title"`
	Body        *string    `json:"body" form:"body"`
	CategoryID  *uuid.UUID `json:"category_id" form:"category_id"`
	Status      *string    `json:"status" form:"status"`
	IsFeatured  *bool      `json:"is_featured" form:"is_featured"`
	IsExclusive *bool      `json:"is_exclusive" form:"is_exclusive"`
	Tags        *string    `json:"tags" form:"tags"`
	MetaTitle   *string    `json:"meta_title" form:"meta_title"`
	MetaDesc    *string    `json:"meta_desc" form:"meta_desc"`
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
//// GET BY ID / SLUG (Single Records)
//////////////////////////////////////////////////////

func (n *News) GetByID(db repository.DatabaseManager, id string) (News, error) {
	var out News

	err, selectErr := db.SelectOneFromDb(&out, "id = ?", id)
	if selectErr != nil {
		return out, err
	}

	return out, err
}

func (n *News) GetBySlug(db repository.DatabaseManager, slug string) (News, error) {
	var out News

	err, selectErr := db.SelectOneFromDb(&out, "slug = ?", slug)
	if selectErr != nil {
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
//// GET FEATURED NEWS
//////////////////////////////////////////////////////

func (n *News) GetFeaturedNews(db repository.DatabaseManager, c *gin.Context) ([]News, repository.PaginationResponse, error) {
	var news []News
	pagination := postgresql.GetPagination(c)

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at", "desc", "is_featured = true",
		pagination, &news, "1=1",
	)
	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving exclusive news list: %w", err)
	}

	db.DB().Preload("Category").Preload("Images").Order("created_at desc").Limit(pagination.Limit).Offset(pagination.Page).Where("is_featured = ?", true).Find(&news)

	return news, paginationResponse, nil
}

func (n *News) GetExclusiveNews(db repository.DatabaseManager, c *gin.Context) ([]News, repository.PaginationResponse, error) {
	var news []News
	pagination := postgresql.GetPagination(c)

	// filter for exclusive news
	where := "status = 'published' AND is_exclusive = ?"

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"",
		pagination,
		&news,
		where,
		true,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving exclusive news list: %w", err)
	}

	db.DB().Preload("Category").Preload("Images").Order("created_at desc").Limit(pagination.Limit).Offset(pagination.Page).Where("is_exclusive = ?", true).Where("status = ?", "published").Find(&news)

	return news, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// GET ALL NEWS (PAGINATED)
//////////////////////////////////////////////////////

func (n *News) GetAllNews(db repository.DatabaseManager, c *gin.Context) ([]News, repository.PaginationResponse, error) {
	var news []News
	pagination := postgresql.GetPagination(c)

	// 🔍 Search query param
	search := c.Query("search")

	// 🔽 Filter params
	categoryID := c.Query("category_id")
	status := c.Query("status")
	isFeatured := c.Query("is_featured")
	isExclusive := c.Query("is_exclusive")

	var conditions []string
	var args []interface{}

	// Search across text fields
	if search != "" {
		conditions = append(conditions, "(title ILIKE ? OR body ILIKE ? OR tags ILIKE ? OR meta_title ILIKE ? OR meta_desc ILIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Filter by category
	if categoryID != "" {
		conditions = append(conditions, "category_id = ?")
		args = append(args, categoryID)
	}

	// Filter by status (draft, published, archived, etc.)
	if status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, status)
	}

	// Filter by is_featured
	if isFeatured != "" {
		val := isFeatured == "true"
		conditions = append(conditions, "is_featured = ?")
		args = append(args, val)
	}

	// Filter by is_exclusive
	if isExclusive != "" {
		val := isExclusive == "true"
		conditions = append(conditions, "is_exclusive = ?")
		args = append(args, val)
	}

	// Build final query string
	var query interface{} = ""
	if len(conditions) > 0 {
		query = strings.Join(conditions, " AND ")
	}

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"",
		pagination,
		&news,
		query,
		args...,
	)
	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving news list: %w", err)
	}

	// Preload associations after paginated fetch
	db.DB().Preload("Category").Preload("Images").
		Order("created_at desc").Limit(pagination.Limit).Offset(pagination.Page).Where(query, args...).
		Find(&news)

	return news, paginationResponse, nil
}

// func (n *News) GetAllNews(db repository.DatabaseManager, c *gin.Context) ([]News, repository.PaginationResponse, error) {
// 	var news []News
// 	pagination := postgresql.GetPagination(c)

// 	// We pass "published_at" as the orderBy and "desc" as the order.
// 	// The helper now correctly concatenates them as "published_at desc".
// 	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
// 		"created_at",
// 		"desc",
// 		"",
// 		pagination,
// 		&news,
// 		nil,
// 	)

// 	if err != nil {
// 		return nil, paginationResponse, fmt.Errorf("failed retrieving news list: %w", err)
// 	}

// 	db.DB().Preload("Category").Preload("Images").Order("created_at desc").Limit(pagination.Limit).Offset(pagination.Page).Find(&news)

// 	return news, paginationResponse, nil
// }

//////////////////////////////////////////////////////
//// FILTER BY CATEGORY
//////////////////////////////////////////////////////

func (n *News) GetNewsByCategory(db repository.DatabaseManager, category string, c *gin.Context) ([]News, repository.PaginationResponse, error) {
	var news []News
	pagination := postgresql.GetPagination(c)

	sort := strings.TrimSpace(c.Query("sort"))
	order := strings.ToLower(strings.TrimSpace(c.Query("order")))
	if sort == "" {
		sort = "created_at"
	}
	if order != "asc" {
		order = "desc"
	}

	// Build WHERE dynamically
	where := "status = 'published'"
	args := []interface{}{}

	if category != "" {
		where += " AND category_id IN (SELECT id FROM categories WHERE slug = ?)"
		args = append(args, category)
	}

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		sort,
		order,
		"",
		pagination,
		&news,
		where,
		args...,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving news list: %w", err)
	}

	db.DB().Preload("Category").Preload("Images").Order("created_at desc").Limit(pagination.Limit).Offset(pagination.Page).Where(where, args...).Find(&news)

	return news, paginationResponse, err
}

//////////////////////////////////////////////////////
//// SEARCH NEWS
//////////////////////////////////////////////////////

func (n *News) SearchNews(db repository.DatabaseManager, search string, c *gin.Context) ([]News, repository.PaginationResponse, error) {
	var news []News
	pagination := postgresql.GetPagination(c)

	sort := strings.TrimSpace(c.Query("sort"))
	order := strings.ToLower(strings.TrimSpace(c.Query("order")))
	if sort == "" {
		sort = "created_at"
	}
	if order != "asc" {
		order = "desc"
	}

	// Build WHERE dynamically
	where := "status = 'published'"
	args := []interface{}{}

	if search != "" {
		where += " AND (LOWER(title) ILIKE ? OR LOWER(body) ILIKE ?)"
		q := "%" + strings.ToLower(search) + "%"
		args = append(args, q, q)
	}

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		sort,
		order,
		"",
		pagination,
		&news,
		where,
		args...,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving news list: %w", err)
	}

	db.DB().Preload("Category").Preload("Images").Order("created_at desc").Limit(pagination.Limit).Offset(pagination.Page).Find(&news)

	return news, paginationResponse, err
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
