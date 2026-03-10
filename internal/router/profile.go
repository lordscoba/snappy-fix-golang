package router

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"

	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/handler/users"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func Profile(r *gin.Engine, ApiVersion string, validator *validator.Validate, db *db.Database, logger *logutil.Logger) *gin.Engine {
	extReq := request.ExternalRequest{Logger: logger, Test: false}
	profile := users.Controller{Db: db, Validator: validator, Logger: logger, ExtReq: extReq}

	ProfileUrl := r.Group(fmt.Sprintf("%v/profile", ApiVersion), middleware.Authorize(db.Postgresql.DB()))
	{

		ProfileUrl.Use(middleware.UserActiveTracker(db.Postgresql.DB(), 60*time.Second))
		ProfileUrl.GET("", profile.GetProfile)
		ProfileUrl.PATCH("", profile.UpdateProfile)
		ProfileUrl.PATCH("/location", profile.UpdateUserLocation)
		ProfileUrl.PATCH("/change-password", profile.ChangePassword)
		// ProfileUrl.GET("/insurance", profile.GetInsurance)

		// for uploading logo and getting logo
		ProfileUrl.GET("/get-image", profile.GetUserImage)
		ProfileUrl.PUT("/update-image", profile.UpdateUserImage)

		// NEW: delete account
		ProfileUrl.DELETE("/delete-account", profile.DeleteMyAccount)

	}
	return r
}
