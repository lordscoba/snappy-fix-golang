package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	image "github.com/snappy-fix-golang/internal/handler/images"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
)

func AdminImages(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	extReq := request.ExternalRequest{}

	controller := image.Controller{
		Db:        db,
		Validator: validator,
		ExtReq:    extReq,
	}

	admin := r.Group(fmt.Sprintf("%v/admin/images", apiVersion), middleware.Authorize(db.Postgresql.DB()))

	{
		admin.POST("", controller.CreateImages)    // upload multiple
		admin.GET("", controller.GetAllImages)     // list + search + pagination
		admin.GET("/:id", controller.GetImageByID) // single image
		admin.DELETE("/:id", controller.DeleteImage)
	}

	return r
}
