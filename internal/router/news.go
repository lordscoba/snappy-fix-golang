package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/internal/adapters/db"
	news "github.com/snappy-fix-golang/internal/handler/news"
)

func BlogNews(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	extReq := request.ExternalRequest{}

	controller := news.Controller{Db: db, Validator: validator, ExtReq: extReq}

	blog := r.Group(fmt.Sprintf("%v/blog", apiVersion))

	{
		blog.GET("/news", controller.GetAllNews)
		blog.GET("/news/:slug", controller.GetNewsBySlug)
		blog.GET("/news/featured", controller.GetFeaturedNews)
		blog.GET("/news/exclusive", controller.GetExclusiveNews)
		blog.GET("/news/category", controller.GetNewsByCategory)
		blog.GET("/news/search", controller.SearchNews)
	}

	return r
}
