package category

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	categoryservice "github.com/snappy-fix-golang/internal/services/category_service"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
}

func (base *Controller) CreateCategory(c *gin.Context) {
	var req entities.CreateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "validation failed", err.Error(), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, code, err := categoryservice.CreateCategoryService(req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(code, "success", data, nil, code)
	c.JSON(code, rd)
}

func (base *Controller) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var req entities.UpdateCategoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "invalid update data", err.Error(), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	data, code, err := categoryservice.UpdateCategoryService(id, req, base.Db.Postgresql.DB())
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(code, "success", data, nil, code)
	c.JSON(code, rd)
}

//////////////////////////////////////////////////////
//// GET ALL
//////////////////////////////////////////////////////

func (base *Controller) GetAllCategories(c *gin.Context) {

	data, pagination, code, err := categoryservice.GetAllCategoriesService(base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// GET TOP LEVEL
//////////////////////////////////////////////////////

func (base *Controller) GetTopLevelCategories(c *gin.Context) {

	data, pagination, code, err := categoryservice.GetTopLevelCategoriesService(base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func (base *Controller) DeleteCategory(c *gin.Context) {

	id := c.Param("id")

	data, code, err := categoryservice.DeleteCategoryService(id, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// GET CATEGORY BY SLUG (BLOG)
//////////////////////////////////////////////////////

func (base *Controller) GetCategoryBySlug(c *gin.Context) {

	slug := c.Param("slug")

	data, code, err := categoryservice.GetCategoryBySlugService(slug, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}
