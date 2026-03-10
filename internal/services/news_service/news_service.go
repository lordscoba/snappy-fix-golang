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
	"gorm.io/gorm"
)

func CreateNewsService(req entities.CreateNewsRequest, thumbnail []byte, inlineImages [][]byte, db *gorm.DB, extReq request.ExternalRequest, catID uuid.UUID) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	news := entities.News{
		Title: req.Title, Slug: req.Slug, Body: req.Body,
		CategoryID: catID, Status: req.Status,
		IsFeatured: req.IsFeatured, IsExclusive: req.IsExclusive,
		Tags: req.Tags, MetaTitle: req.MetaTitle, MetaDesc: req.MetaDesc,
	}

	// // 1. Upload Thumbnail
	// if len(thumbnail) > 0 {
	// 	resp, _ := uploadToCloudinary(extReq, thumbnail, "news/thumbnails")
	// 	news.ThumbnailID = resp.PublicID
	// 	news.ThumbnailURL = resp.SecureURL
	// }

	// // 2. Upload Inline Images
	// for _, img := range inlineImages {
	// 	resp, _ := uploadToCloudinary(extReq, img, "news/content")
	// 	news.Images = append(news.Images, entities.NewsImage{
	// 		PublicID: resp.PublicID, URL: resp.SecureURL, ImageType: "inline",
	// 	})
	// }

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

func UpdateNewsService(id string, req entities.UpdateNewsRequest, newThumbnail []byte, db *gorm.DB, extReq request.ExternalRequest) (gin.H, int, error) {
	pdb := inst.InitDB(db)
	var news entities.News
	existing, err := news.GetByID(pdb, id)
	if err != nil {
		return gin.H{}, 404, errors.New("not found")
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Body != nil {
		updates["body"] = *req.Body
	}

	// 3. Update Image & Clean Old One
	if len(newThumbnail) > 0 {
		// Delete old
		if existing.ThumbnailID != "" {
			_ = deleteFromCloudinary(extReq, existing.ThumbnailID)
		}
		// Upload new
		resp, _ := uploadToCloudinary(extReq, newThumbnail, "news/thumbnails")
		updates["thumbnail_id"] = resp.PublicID
		updates["thumbnail_url"] = resp.SecureURL
	}

	updated, err := existing.UpdateByID(pdb, updates, id)
	return gin.H{"news": updated}, 200, err
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
