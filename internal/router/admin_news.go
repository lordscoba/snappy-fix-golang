package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	news "github.com/snappy-fix-golang/internal/handler/news"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
)

func AdminNews(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	extReq := request.ExternalRequest{}

	controller := news.Controller{Db: db, Validator: validator, ExtReq: extReq}

	admin := r.Group(fmt.Sprintf("%v/admin/news", apiVersion), middleware.Authorize(db.Postgresql.DB()))

	{

		admin.POST("", controller.CreateNews)
		admin.PUT("/:id", controller.UpdateNews)
		admin.GET("", controller.GetAllNews)
		admin.GET("/:slug", controller.GetNewsBySlug)
		admin.DELETE("/:id", controller.DeleteNews)

	}

	return r
}
