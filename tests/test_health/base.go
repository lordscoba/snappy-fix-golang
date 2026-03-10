package testhealth

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/snappy-fix-golang/internal/adapters/db"
	httphealth "github.com/snappy-fix-golang/internal/handler/http_health"

	"github.com/snappy-fix-golang/tests"
)

var (
	setupOnce        sync.Once
	sharedRouter     *gin.Engine
	sharedController *httphealth.Controller
)

func SetupTestRouter() (*gin.Engine, *httphealth.Controller) {
	setupOnce.Do(func() {
		gin.SetMode(gin.TestMode)

		logger := tests.Setup()
		dbConn := db.Connection() // Called only once
		validator := validator.New()

		sharedController = &httphealth.Controller{
			Db:        dbConn,
			Validator: validator,
			Logger:    logger,
		}

		r := gin.Default()
		registerHealthRoutes(r, sharedController)
		sharedRouter = r
	})

	return sharedRouter, sharedController
}

func registerHealthRoutes(router *gin.Engine, controller *httphealth.Controller) {
	v1 := router.Group("/api/v1")
	{
		v1.POST("/health", controller.Post)
		v1.GET("/health", controller.Get)
	}
}
