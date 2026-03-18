package categoryservice

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	"github.com/snappy-fix-golang/pkg/utils"
	"gorm.io/gorm"
)

func CreateCategoryService(req entities.CreateCategoryRequest, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	slug := utils.GenerateSlug(req.Name)
	category := entities.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Description: req.Description,
		Slug:        slug,
		Level:       req.Level,
	}

	if err := category.Create(pdb); err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"category": category}, http.StatusCreated, nil
}

func UpdateCategoryService(id string, req entities.UpdateCategoryRequest, db *gorm.DB) (gin.H, int, error) {
	pdb := inst.InitDB(db)
	var cat entities.Category

	existing, err := cat.GetByID(pdb, id)
	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	// Dynamic updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}

	if req.Level != nil {
		updates["level"] = *req.Level
	}

	// Special handling for ParentID to allow setting it to NULL
	if req.ParentID != nil {
		updates["parent_id"] = *req.ParentID
	}

	updated, err := existing.UpdateByID(pdb, updates, id)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"category": updated}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET CATEGORY BY ID
//////////////////////////////////////////////////////

func GetCategoryByIDService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var cat entities.Category

	out, err := cat.GetByID(pdb, id)

	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	return gin.H{"category": out}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET CATEGORY BY SLUG
//////////////////////////////////////////////////////

func GetCategoryBySlugService(slug string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var cat entities.Category

	out, err := cat.GetBySlug(pdb, slug)

	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	return gin.H{"category": out}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET ALL CATEGORIES
//////////////////////////////////////////////////////

func GetAllCategoriesService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var cat entities.Category

	list, pagination, err := cat.GetAllCategories(pdb, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"categories": list}, pagination, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET TOP LEVEL CATEGORIES
//////////////////////////////////////////////////////

func GetTopLevelCategoriesService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var cat entities.Category

	list, pagination, err := cat.GetTopLevelCategories(pdb, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"categories": list}, pagination, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// DELETE CATEGORY
//////////////////////////////////////////////////////

func DeleteCategoryService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var cat entities.Category

	existing, err := cat.GetByID(pdb, id)

	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	if err := existing.DeleteCategory(pdb); err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"message": "Category deleted successfully"}, http.StatusOK, nil
}
