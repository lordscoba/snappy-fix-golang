package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/handler/contact"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
)

func Contact(r *gin.Engine, apiVersion string, validator *validator.Validate, db *db.Database) *gin.Engine {

	controller := contact.Controller{
		Db:        db,
		Validator: validator,
	}

	contactPublicRoutes := r.Group(fmt.Sprintf("%v/public/contact", apiVersion))
	{
		contactPublicRoutes.POST("", controller.CreateMessage)
	}

	contactRoutes := r.Group(fmt.Sprintf("%v/contact", apiVersion), middleware.Authorize(db.Postgresql.DB()))
	{

		// Admin / dashboard
		contactRoutes.GET("", controller.GetAllMessages)
		contactRoutes.GET("/:id", controller.GetMessageByID)
		contactRoutes.PATCH("/:id/read", controller.MarkAsRead)
		contactRoutes.DELETE("/:id", controller.DeleteMessage)

		// Stats
		contactRoutes.GET("/count/all", controller.CountAllMessages)
	}

	return r
}
