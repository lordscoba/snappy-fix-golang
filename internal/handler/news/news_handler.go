package news

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	imageservice "github.com/snappy-fix-golang/internal/services/image_service"
	newsservice "github.com/snappy-fix-golang/internal/services/news_service"
	"github.com/snappy-fix-golang/pkg/utils/responses"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
	ExtReq    request.ExternalRequest
}

func (base *Controller) CreateNews(c *gin.Context) {
	var req entities.CreateNewsRequest

	if err := c.ShouldBind(&req); err != nil {
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Invalid request data", err.Error(), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	// 1. Optional Category Handling
	var catID *uuid.UUID
	if req.CategoryID != "" {
		parsedID, err := uuid.FromString(req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(http.StatusBadRequest, "error", "Invalid Category ID format", err.Error(), nil))
			return
		}
		catID = &parsedID
	}

	// 2. Thumbnail Logic (File vs Library)
	var thumbBytes []byte
	if req.ThumbnailUrl == "" {
		fileHeader, err := c.FormFile("thumbnail")
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(http.StatusBadRequest, "error", "Thumbnail is required", "Provide a file or library URL", nil))
			return
		}

		f, _ := fileHeader.Open()
		defer f.Close()
		b, _ := io.ReadAll(f)

		// Optimization
		optimized, _, err := imageservice.ValidateAndOptimize(b)
		if err != nil {
			status := http.StatusInternalServerError
			if _, ok := err.(*imageservice.ValidationError); ok {
				status = http.StatusBadRequest
			}
			c.JSON(status, responses.BuildErrorResponse(status, "error", "Image optimization failed", err.Error(), nil))
			return
		}
		thumbBytes = optimized
	}

	// 3. Call Service (Notice: removed inlineBytes)
	data, code, err := newsservice.CreateNewsService(req, thumbBytes, base.Db.Postgresql.DB(), base.ExtReq, catID)
	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err.Error(), nil))
		return
	}

	c.JSON(code, responses.BuildSuccessResponse(code, "success", data, nil, code))
}

func (base *Controller) UpdateNews(c *gin.Context) {
	id := c.Param("id")
	var req entities.UpdateNewsRequest

	if err := c.ShouldBind(&req); err != nil {
		// Default error message
		errorMessage := "Invalid request data"

		// 2. Check if it's a validation error
		if vErrs, ok := err.(validator.ValidationErrors); ok {
			errorMessage = "Validation failed: "
			for i, vErr := range vErrs {
				// e.g., "Title is required", "CategoryID must be a valid UUID"
				errorMessage += fmt.Sprintf("%s is %s", vErr.Field(), vErr.Tag())
				if i < len(vErrs)-1 {
					errorMessage += ", "
				}
			}
		}

		// 3. Use your helper with the detailed message
		rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", errorMessage, err.Error(), nil)
		c.JSON(http.StatusBadRequest, rd)
		return
	}

	// 1. Optional Category Handling
	var catID *uuid.UUID
	if req.CategoryID != nil && strings.TrimSpace(*req.CategoryID) != "" {
		parsedID, err := uuid.FromString(*req.CategoryID)
		if err != nil {
			c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(http.StatusBadRequest, "error", "Invalid Category ID format", err.Error(), nil))
			return
		}
		catID = &parsedID
	}

	var thumbBytes []byte
	fileHeader, err := c.FormFile("thumbnail")
	// thumbFile, thumbHeader, err := c.Request.FormFile("thumbnail")

	// If a new file is uploaded
	if err == nil {
		file, _ := fileHeader.Open()
		defer file.Close()

		rawBytes, _ := io.ReadAll(file)

		// Optimize and Validate the image
		optimizedBytes, _, err := imageservice.ValidateAndOptimize(rawBytes)
		if err != nil {
			rd := responses.BuildErrorResponse(http.StatusBadRequest, "error", "Image validation failed", err.Error(), nil)
			c.JSON(http.StatusBadRequest, rd)
			return
		}
		thumbBytes = optimizedBytes
	}

	data, code, err := newsservice.UpdateNewsService(
		id,
		req,
		thumbBytes,
		base.Db.Postgresql.DB(),
		base.ExtReq,
		catID,
	)

	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err.Error(), nil))
		return
	}

	c.JSON(code, responses.BuildSuccessResponse(code, "success", data, nil, code))
}

//////////////////////////////////////////////////////
//// GET ALL NEWS
//////////////////////////////////////////////////////

func (base *Controller) GetAllNews(c *gin.Context) {
	data, pagination, code, err := newsservice.GetAllNewsService(base.Db.Postgresql.DB(), c)
	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}
	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// GET NEWS BY SLUG (PUBLIC BLOG)
//////////////////////////////////////////////////////

func (base *Controller) GetNewsBySlug(c *gin.Context) {

	slug := c.Param("slug")

	data, code, err := newsservice.GetNewsBySlugService(slug, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// DELETE
//////////////////////////////////////////////////////

func (base *Controller) DeleteNews(c *gin.Context) {

	id := c.Param("id")

	data, code, err := newsservice.DeleteNewsService(id, base.Db.Postgresql.DB())

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

func (base *Controller) GetFeaturedNews(c *gin.Context) {

	data, pagination, code, err := newsservice.GetFeaturedNewsService(
		base.Db.Postgresql.DB(),
		c,
	)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

func (base *Controller) GetExclusiveNews(c *gin.Context) {
	data, pagination, code, err := newsservice.GetExclusiveNewsService(base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

func (base *Controller) GetNewsByCategory(c *gin.Context) {

	category := c.Query("category")
	data, pagination, code, err := newsservice.GetNewsByCategoryService(category, base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}

func (base *Controller) SearchNews(c *gin.Context) {

	query := c.Query("search")

	data, pagination, code, err := newsservice.SearchNewsService(query, base.Db.Postgresql.DB(), c)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, pagination)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, pagination, code)
	c.JSON(http.StatusOK, rd)
}
