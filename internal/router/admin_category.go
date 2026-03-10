package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/handler/category"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
)

func AdminCategory(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	controller := category.Controller{Db: db, Validator: validator}

	admin := r.Group(fmt.Sprintf("%v/admin/categories", apiVersion), middleware.Authorize(db.Postgresql.DB()), middleware.Authorize(db.Postgresql.DB()))

	{
		admin.POST("", controller.CreateCategory)
		admin.GET("", controller.GetAllCategories)
		admin.GET("/top", controller.GetTopLevelCategories)
		admin.PUT("/:id", controller.UpdateCategory)
		admin.DELETE("/:id", controller.DeleteCategory)
		admin.GET("/:slug", controller.GetCategoryBySlug)
	}

	return r
}
