package newsservice

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/external/thirdparty/cloudinary"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	"github.com/snappy-fix-golang/pkg/utils"
	"gorm.io/gorm"
)

func CreateNewsService(req entities.CreateNewsRequest, thumbnail []byte, inlineImages [][]byte, db *gorm.DB, extReq request.ExternalRequest, catID uuid.UUID) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	slug := utils.GenerateSlug(req.Title)

	news := entities.News{
		Title: req.Title, Slug: slug, Body: req.Body,
		CategoryID: catID, Status: req.Status,
		IsFeatured: req.IsFeatured, IsExclusive: req.IsExclusive,
		Tags: req.Tags, MetaTitle: req.MetaTitle, MetaDesc: req.MetaDesc,
	}

	//////////////////////////////////////////////////////
	// 1. Upload Thumbnail
	//////////////////////////////////////////////////////

	resp, err := uploadToCloudinary(extReq, thumbnail, "news/thumbnails")
	if err != nil {
		return gin.H{}, 500, fmt.Errorf("thumbnail upload failed: %w", err)
	}

	news.ThumbnailID = resp.PublicID
	news.ThumbnailURL = resp.SecureURL

	//////////////////////////////////////////////////////
	// 2. Upload Inline Images
	//////////////////////////////////////////////////////

	for _, img := range inlineImages {

		resp, err := uploadToCloudinary(extReq, img, "news/content")

		if err != nil {
			return gin.H{}, 500, fmt.Errorf("inline image upload failed: %w", err)
		}

		news.Images = append(news.Images, entities.NewsImage{
			PublicID:  resp.PublicID,
			URL:       resp.SecureURL,
			ImageType: "inline",
		})
	}

	if err := news.Create(pdb); err != nil {
		return gin.H{}, 500, err
	}

	// reload with category
	if err := pdb.DB().Preload("Category").First(&news, "id = ?", news.ID).Error; err != nil {
		return gin.H{}, 500, err
	}

	return gin.H{"news": news}, 201, nil
}

func UpdateNewsService(
	id string,
	req entities.UpdateNewsRequest,
	newThumbnail []byte,
	inlineImages [][]byte,
	db *gorm.DB,
	extReq request.ExternalRequest,
) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var news entities.News

	existing, err := news.GetByID(pdb, id)

	if err != nil {
		return gin.H{}, http.StatusNotFound, errors.New("news not found")
	}

	updates := make(map[string]interface{})

	//////////////////////////////////////////////////////
	// Title + Slug update
	//////////////////////////////////////////////////////

	if req.Title != nil {

		updates["title"] = *req.Title
	}

	//////////////////////////////////////////////////////
	// Body
	//////////////////////////////////////////////////////

	if req.Body != nil {
		updates["body"] = *req.Body
	}

	//////////////////////////////////////////////////////
	// Category
	//////////////////////////////////////////////////////

	if req.CategoryID != nil {
		updates["category_id"] = *req.CategoryID
	}

	//////////////////////////////////////////////////////
	// Status
	//////////////////////////////////////////////////////

	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if req.IsFeatured != nil {
		updates["is_featured"] = *req.IsFeatured
	}

	if req.IsExclusive != nil {
		updates["is_exclusive"] = *req.IsExclusive
	}

	if req.Tags != nil {
		updates["tags"] = *req.Tags
	}

	if req.MetaTitle != nil {
		updates["meta_title"] = *req.MetaTitle
	}

	if req.MetaDesc != nil {
		updates["meta_desc"] = *req.MetaDesc
	}

	//////////////////////////////////////////////////////
	// Thumbnail Update
	//////////////////////////////////////////////////////

	if len(newThumbnail) > 0 {

		if existing.ThumbnailID != "" {
			_ = deleteFromCloudinary(extReq, existing.ThumbnailID)
		}

		resp, err := uploadToCloudinary(extReq, newThumbnail, "news/thumbnails")

		if err != nil {
			return gin.H{}, 500, fmt.Errorf("thumbnail upload failed: %w", err)
		}

		updates["thumbnail_id"] = resp.PublicID
		updates["thumbnail_url"] = resp.SecureURL
	}

	//////////////////////////////////////////////////////
	// Inline Images Upload
	//////////////////////////////////////////////////////

	for _, img := range inlineImages {

		resp, err := uploadToCloudinary(extReq, img, "news/content")

		if err != nil {
			return gin.H{}, 500, fmt.Errorf("inline image upload failed: %w", err)
		}

		image := entities.NewsImage{
			NewsID:    existing.ID,
			PublicID:  resp.PublicID,
			URL:       resp.SecureURL,
			ImageType: "inline",
		}

		err = pdb.DB().Create(&image).Error

		if err != nil {
			return gin.H{}, 500, err
		}
	}

	//////////////////////////////////////////////////////
	// Update record
	//////////////////////////////////////////////////////

	updated, err := existing.UpdateByID(pdb, updates, id)

	if err != nil {
		return gin.H{}, 500, err
	}

	//////////////////////////////////////////////////////
	// Reload relations
	//////////////////////////////////////////////////////

	err = pdb.DB().
		Preload("Category").
		Preload("Images").
		First(&updated, "id = ?", id).Error

	if err != nil {
		return gin.H{}, 500, err
	}

	return gin.H{"news": updated}, http.StatusOK, nil
}

// Helpers
func uploadToCloudinary(extReq request.ExternalRequest, b []byte, folder string) (cloudinary.UploadResponse, error) {
	res, err := extReq.SendExternalRequestCloudinary("cloudinary_upload_image", cloudinary.UploadInput{
		Bytes: b, Folder: folder,
	})
	if r, ok := res.(cloudinary.UploadResponse); ok {
		return r, nil
	}
	return cloudinary.UploadResponse{}, err
}

func deleteFromCloudinary(extReq request.ExternalRequest, publicID string) error {
	_, err := extReq.SendExternalRequestCloudinary("cloudinary_delete_image", cloudinary.DeleteInput{PublicID: publicID})
	return err
}

//////////////////////////////////////////////////////
//// GET NEWS BY ID
//////////////////////////////////////////////////////

func GetNewsByIDService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)
	var news entities.News
	out, err := news.GetByID(pdb, id)
	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gin.H{}, http.StatusNotFound, errors.New("news not found")
		}

		return gin.H{}, http.StatusBadRequest, err
	}

	return gin.H{"news": out}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET NEWS BY SLUG
//////////////////////////////////////////////////////

func GetNewsBySlugService(slug string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var news entities.News

	out, err := news.GetBySlug(pdb, slug)

	if err != nil {
		return gin.H{}, http.StatusNotFound, err
	}

	return gin.H{"news": out}, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// GET ALL NEWS
//////////////////////////////////////////////////////

func GetAllNewsService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var news entities.News

	list, pagination, err := news.GetAllNews(pdb, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"news": list}, pagination, http.StatusOK, nil
}

func GetFeaturedNewsService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var n entities.News

	list, pagination, err := n.GetFeaturedNews(pdb, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"news": list}, pagination, http.StatusOK, nil
}

func GetExclusiveNewsService(db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var n entities.News

	list, pagination, err := n.GetExclusiveNews(pdb, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"news": list}, pagination, http.StatusOK, nil
}

func GetNewsByCategoryService(category string, db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var n entities.News

	list, pagination, err := n.GetNewsByCategory(pdb, category, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"news": list}, pagination, http.StatusOK, nil
}

func SearchNewsService(query string, db *gorm.DB, c *gin.Context) (gin.H, repository.PaginationResponse, int, error) {

	pdb := inst.InitDB(db)

	var n entities.News

	list, pagination, err := n.SearchNews(pdb, query, c)

	if err != nil {
		return gin.H{}, pagination, http.StatusBadRequest, err
	}

	return gin.H{"news": list}, pagination, http.StatusOK, nil
}

//////////////////////////////////////////////////////
//// DELETE NEWS
//////////////////////////////////////////////////////

func DeleteNewsService(id string, db *gorm.DB) (gin.H, int, error) {

	pdb := inst.InitDB(db)

	var n entities.News

	existing, err := n.GetByID(pdb, id)

	if err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return gin.H{}, http.StatusNotFound, errors.New("news not found")
		}

		return gin.H{}, http.StatusBadRequest, err
	}

	err = existing.DeleteNews(pdb)
	if err != nil {
		return gin.H{}, http.StatusInternalServerError, err
	}

	return gin.H{"message": "News deleted successfully"}, http.StatusOK, nil
}
