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

type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ParentID    *uuid.UUID `gorm:"type:uuid" json:"parent_id,omitempty"`
	Name        string     `gorm:"size:255;not null;index" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	Slug        string     `gorm:"size:255;uniqueIndex;not null" json:"slug"`
	Level       int        `gorm:"default:0" json:"level"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	// One category has many news items
	News []News `gorm:"foreignKey:CategoryID" json:"news,omitempty"`
}

// DTO for Creating
type CreateCategoryRequest struct {
	ParentID    *uuid.UUID `json:"parent_id"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Level       int        `json:"level"`
}

// DTO for Updating (Pointers allow us to see what was actually sent)
type UpdateCategoryRequest struct {
	ParentID    **uuid.UUID `json:"parent_id"` // Double pointer handles setting a field to NULL explicitly
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Level       *int        `json:"level"`
}

//////////////////////////////////////////////////////
//// CREATE
//////////////////////////////////////////////////////

func (c *Category) Create(db repository.DatabaseManager) error {

	err := db.CreateOneRecord(c)
	if err != nil {
		return fmt.Errorf("failed to create category record: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////
//// UPDATE
//////////////////////////////////////////////////////

func (c *Category) UpdateByID(
	db repository.DatabaseManager,
	updates map[string]interface{},
	id string,
) (*Category, error) {

	result, err := db.UpdateFields(c, updates, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update category record: %w", err)
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("no category record found to update")
	}

	return c, nil
}

//////////////////////////////////////////////////////
//// GET BY ID
//////////////////////////////////////////////////////

func (c *Category) GetByID(db repository.DatabaseManager, id string) (Category, error) {

	var out Category

	err, selErr := db.SelectOneFromDb(&out, "id = ?", id)

	if selErr != nil {
		return out, fmt.Errorf("category not found")
	}

	if err != nil {
		return out, fmt.Errorf("failed retrieving category by id: %w", err)
	}

	return out, nil
}

//////////////////////////////////////////////////////
//// GET BY SLUG
//////////////////////////////////////////////////////

func (c *Category) GetBySlug(db repository.DatabaseManager, slug string) (Category, error) {

	var out Category

	err, selErr := db.SelectOneFromDb(&out, "slug = ?", slug)

	if selErr != nil {
		return out, fmt.Errorf("category with slug '%s' not found", slug)
	}

	if err != nil {
		return out, fmt.Errorf("failed retrieving category by slug: %w", err)
	}

	return out, nil
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func (c *Category) DeleteCategory(db repository.DatabaseManager) error {

	err := db.DeleteRecordFromDb(c)

	if err != nil {
		return fmt.Errorf("failed to delete category record: %w", err)
	}

	return nil
}

//////////////////////////////////////////////////////
//// GET ALL CATEGORIES (PAGINATED)
//////////////////////////////////////////////////////

func (c *Category) GetAllCategories(
	db repository.DatabaseManager,
	g *gin.Context,
) ([]Category, repository.PaginationResponse, error) {

	var categories []Category

	pagination := postgresql.GetPagination(g)

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"",
		pagination,
		&categories,
		nil,
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving categories list: %w", err)
	}

	return categories, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// GET TOP LEVEL CATEGORIES
//////////////////////////////////////////////////////

func (c *Category) GetTopLevelCategories(
	db repository.DatabaseManager,
	g *gin.Context,
) ([]Category, repository.PaginationResponse, error) {

	var categories []Category

	pagination := postgresql.GetPagination(g)

	paginationResponse, err := db.SelectAllFromDbOrderByPaginated(
		"created_at",
		"desc",
		"",
		pagination,
		&categories,
		"parent_id IS NULL",
	)

	if err != nil {
		return nil, paginationResponse, fmt.Errorf("failed retrieving top level categories: %w", err)
	}

	return categories, paginationResponse, nil
}

//////////////////////////////////////////////////////
//// COUNT ALL CATEGORIES
//////////////////////////////////////////////////////

func (c *Category) CountAllCategories(db repository.DatabaseManager) (int64, error) {

	count, err := db.CountRecords(&Category{})

	if err != nil {
		return 0, fmt.Errorf("failed counting category records: %w", err)
	}

	return count, nil
}

//////////////////////////////////////////////////////
//// COUNT CHILD CATEGORIES
//////////////////////////////////////////////////////

func (c *Category) CountChildCategories(
	db repository.DatabaseManager,
	parentID string,
) (int64, error) {

	count, err := db.CountSpecificRecords(&Category{}, "parent_id = ?", parentID)

	if err != nil {
		return 0, fmt.Errorf("failed counting child categories: %w", err)
	}

	return count, nil
}
