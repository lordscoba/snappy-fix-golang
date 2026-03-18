package images

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	imageservice "github.com/snappy-fix-golang/internal/services/image_service"
	"github.com/snappy-fix-golang/pkg/utils/responses"
	"golang.org/x/sync/errgroup"
)

type Controller struct {
	Db        *db.Database
	Validator *validator.Validate
	ExtReq    request.ExternalRequest
}

func (base *Controller) CreateImages(c *gin.Context) {
	var baseReq entities.CreateNewsImageRequest

	if err := c.ShouldBind(&baseReq); err != nil {
		// ... (Keep your existing validator logic here)
		return
	}

	form, err := c.MultipartForm()
	if err != nil || form == nil {
		c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(400, "error", "Invalid form data", nil, nil))
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(400, "error", "No images uploaded", nil, nil))
		return
	}

	var mu sync.Mutex
	var uploaded []interface{}
	g := new(errgroup.Group)

	for _, fileHeader := range files {
		fileHeader := fileHeader

		g.Go(func() error {
			f, err := fileHeader.Open()
			if err != nil {
				return fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
			}
			defer f.Close()

			b, err := io.ReadAll(f)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", fileHeader.Filename, err)
			}

			// 1. Process Image
			optimizedBytes, mimeType, err := imageservice.ValidateAndOptimize(b)
			if err != nil {
				// We return the error as-is; it might be a ValidationError
				return err
			}

			// 2. Prepare file-specific request
			fileReq := baseReq
			if fileReq.FileName == "" {
				fileReq.FileName = fileHeader.Filename
			}
			fileReq.ImageType = mimeType
			fileReq.Extension = "webp"

			// 3. Service Call
			data, _, err := imageservice.CreateNewsImageService(
				fileReq,
				optimizedBytes,
				base.Db.Postgresql.DB(),
				base.ExtReq,
			)
			if err != nil {
				return err
			}

			mu.Lock()
			uploaded = append(uploaded, data["image"])
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		var vErr *imageservice.ValidationError

		// Check if it's a validation/security error (400)
		if errors.As(err, &vErr) {
			c.JSON(http.StatusBadRequest, responses.BuildErrorResponse(
				http.StatusBadRequest,
				"error",
				vErr.Message, // "invalid format: application/x-sh" etc
				err.Error(),
				nil,
			))
			return
		}

		// Otherwise, it's a real server error (500)
		c.JSON(http.StatusInternalServerError, responses.BuildErrorResponse(
			http.StatusInternalServerError,
			"error",
			"Internal server error during processing",
			err.Error(),
			nil,
		))
		return
	}

	c.JSON(http.StatusCreated, responses.BuildSuccessResponse(
		http.StatusCreated,
		"success",
		gin.H{"images": uploaded},
		nil,
		http.StatusCreated,
	))
}

//////////////////////////////////////////////////////
//// GET IMAGE BY ID
//////////////////////////////////////////////////////

func (base *Controller) GetImageByID(c *gin.Context) {

	id := c.Param("id")

	data, code, err := imageservice.GetNewsImageByIDService(
		id,
		base.Db.Postgresql.DB(),
	)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}

//////////////////////////////////////////////////////
//// GET ALL IMAGES
//////////////////////////////////////////////////////

func (base *Controller) GetAllImages(c *gin.Context) {

	data, pagination, code, err := imageservice.GetAllNewsImagesService(
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

//////////////////////////////////////////////////////
//// DELETE IMAGE
//////////////////////////////////////////////////////

func (base *Controller) DeleteImage(c *gin.Context) {

	id := c.Param("id")

	data, code, err := imageservice.DeleteNewsImageService(
		id,
		base.Db.Postgresql.DB(),
		base.ExtReq,
	)

	if err != nil {
		rd := responses.BuildErrorResponse(code, "error", err.Error(), err, nil)
		c.JSON(code, rd)
		return
	}

	rd := responses.BuildSuccessResponse(http.StatusOK, "success", data, nil, code)
	c.JSON(http.StatusOK, rd)
}
