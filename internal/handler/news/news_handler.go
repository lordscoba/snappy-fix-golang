package news

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
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
	// Use ShouldBind to catch both JSON and Form values
	// 1. Attempt to bind

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

	catID, err := uuid.FromString(req.CategoryID)
	if err != nil {

		rd := responses.BuildErrorResponse(
			http.StatusBadRequest,
			"error",
			"Invalid Category ID format",
			err.Error(),
			nil,
		)

		c.JSON(http.StatusBadRequest, rd)
		return
	}

	// // 1. Get Thumbnail
	// thumbFile, _, _ := c.Request.FormFile("thumbnail")
	// var thumbBytes []byte
	// if thumbFile != nil {
	// 	thumbBytes, _ = io.ReadAll(thumbFile)
	// }

	// // 2. Get Multiple Inline Images
	// form, err := c.MultipartForm()
	// if err != nil {
	// 	form = &multipart.Form{}
	// }
	// files := form.File["images"]
	// var inlineBytes [][]byte
	// for _, file := range files {
	// 	f, _ := file.Open()
	// 	b, _ := io.ReadAll(f)
	// 	inlineBytes = append(inlineBytes, b)
	// }

	// 1. Get Thumbnail (Required)
	thumbFile, thumbHeader, err := c.Request.FormFile("thumbnail")
	if err != nil || thumbHeader == nil || thumbHeader.Size == 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Thumbnail is required",
		})
		return
	}

	thumbBytes, err := io.ReadAll(thumbFile)
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Failed to read thumbnail",
		})
		return
	}

	// 2. Get Multiple Inline Images (Required)
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "Images are required",
		})
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(400, gin.H{
			"status":  "error",
			"message": "At least one image is required",
		})
		return
	}

	var inlineBytes [][]byte

	for _, file := range files {
		f, err := file.Open()
		if err != nil {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Failed to open image",
			})
			return
		}

		b, err := io.ReadAll(f)
		if err != nil || len(b) == 0 {
			c.JSON(400, gin.H{
				"status":  "error",
				"message": "Invalid image file",
			})
			return
		}

		inlineBytes = append(inlineBytes, b)
	}

	data, code, err := newsservice.CreateNewsService(req, thumbBytes, inlineBytes, base.Db.Postgresql.DB(), base.ExtReq, catID)
	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err, nil))
		return
	}
	c.JSON(code, responses.BuildSuccessResponse(code, "success", data, nil, code))
}

func (base *Controller) UpdateNews(c *gin.Context) {
	id := c.Param("id")
	var req entities.UpdateNewsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, responses.BuildErrorResponse(400, "error", "invalid body", err, nil))
		return
	}

	// Handle thumbnail update via form-data if provided
	thumbFile, _, _ := c.Request.FormFile("thumbnail")
	var thumbBytes []byte
	if thumbFile != nil {
		thumbBytes, _ = io.ReadAll(thumbFile)
	}

	data, code, err := newsservice.UpdateNewsService(id, req, thumbBytes, base.Db.Postgresql.DB(), base.ExtReq)
	if err != nil {
		c.JSON(code, responses.BuildErrorResponse(code, "error", err.Error(), err, nil))
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
