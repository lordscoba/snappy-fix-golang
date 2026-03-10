package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/handler/category"
)

func BlogCategory(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	controller := category.Controller{Db: db, Validator: validator}

	blog := r.Group(fmt.Sprintf("%v/blog/categories", apiVersion))

	{
		blog.GET("", controller.GetAllCategories)
		blog.GET("/top", controller.GetTopLevelCategories)
		blog.GET("/:slug", controller.GetCategoryBySlug)
	}

	return r
}
