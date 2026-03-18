package imageservice

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	"gorm.io/gorm"
)

func CreateNewsImageService(
	req entities.CreateNewsImageRequest,
	fileBytes []byte,
	db *gorm.DB,
	extReq request.ExternalRequest,
) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	//////////////////////////////////////////////////////
	// Upload to Cloudinary
	//////////////////////////////////////////////////////

	resp, err := uploadToCloudinary(extReq, fileBytes, "news/media")

	if err != nil {
		return gin.H{}, http.StatusInternalServerError, fmt.Errorf("image upload failed: %w", err)
	}

	//////////////////////////////////////////////////////
	// Handle optional NewsID
	//////////////////////////////////////////////////////

	var newsUUID *uuid.UUID

	if req.NewsID != nil {
		parsed, err := uuid.FromString(*req.NewsID)
		if err != nil {
			return gin.H{}, http.StatusBadRequest, fmt.Errorf("invalid news_id")
		}
		newsUUID = &parsed
	}

	//////////////////////////////////////////////////////
	// Create record
	//////////////////////////////////////////////////////

	image := entities.NewsImage{
		NewsID:      newsUUID,
		PublicID:    resp.PublicID,
		URL:         resp.SecureURL,
		ImageType:   req.ImageType,
		FileName:    req.FileName,
		Extension:   req.Extension,
		Size:        int64(len(fileBytes)),
		Description: req.Description,
	}

	if err := image.Create(pdb); err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"image": image}, http.StatusCreated, nil
}

func GetNewsImageByIDService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var img entities.NewsImage

	out, err := img.GetByID(pdb, id)

	if err != nil {
		return gin.H{}, http.StatusNotFound, errors.New("image not found")
	}

	return gin.H{"image": out}, http.StatusOK, nil
}

func GetAllNewsImagesService(
	db *gorm.DB,
	c *gin.Context,
) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var img entities.NewsImage

	list, pagination, err := img.GetAllImages(pdb, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"images": list}, pagination, http.StatusOK, nil
}

func DeleteNewsImageService(
	id string,
	db *gorm.DB,
	extReq request.ExternalRequest,
) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var img entities.NewsImage

	existing, err := img.GetByID(pdb, id)

	if err != nil {
		return gin.H{}, http.StatusNotFound, errors.New("image not found")
	}

	//////////////////////////////////////////////////////
	// Delete from Cloudinary
	//////////////////////////////////////////////////////

	if existing.PublicID != "" {
		_ = deleteFromCloudinary(extReq, existing.PublicID)
	}

	//////////////////////////////////////////////////////
	// Delete from DB
	//////////////////////////////////////////////////////

	err = existing.DeleteNewsImage(pdb)

	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"message": "Image deleted successfully"}, http.StatusOK, nil
}
