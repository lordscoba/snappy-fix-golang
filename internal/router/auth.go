package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/external/request"

	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/handler/users"
	middleware "github.com/snappy-fix-golang/internal/midlleware"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func Auth(r *gin.Engine, ApiVersion string, validator *validator.Validate, db *db.Database, logger *logutil.Logger) *gin.Engine {
	extReq := request.ExternalRequest{Logger: logger, Test: false}
	auth := users.Controller{Db: db, Validator: validator, Logger: logger, ExtReq: extReq}

	AuthUrl := r.Group(fmt.Sprintf("%v/auth", ApiVersion))
	{
		AuthUrl.POST("/register", auth.RegisterHandler)
		AuthUrl.POST("/login", auth.LoginHandler)
		AuthUrl.POST("/refresh-token", auth.RefreshHandler)

		AuthUrl.POST("/forgot-password", auth.ForgotPasswordHandler)
		AuthUrl.POST("/reset-password", auth.ResetPasswordHandler)

		// email verification with url
		AuthUrl.POST("/verify-email", auth.VerifyEmailHandler)
		AuthUrl.POST("/resend-verification", auth.ResendVerifyEmailHandler)

	}

	AuthUrlSecure := r.Group(fmt.Sprintf("%v/auth", ApiVersion))
	{
		AuthUrlSecure.Use(middleware.Authorize(db.Postgresql.DB()))
		AuthUrlSecure.POST("/logout", auth.LogoutHandler)

	}

	return r
}
