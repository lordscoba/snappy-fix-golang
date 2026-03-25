package router

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	// _ "github.com/snappy-fix-golang/docs"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/config"

	middleware "github.com/snappy-fix-golang/internal/midlleware"
	logutil "github.com/snappy-fix-golang/pkg/logger"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Nactive Health API
//	@version		1.0
//	@description	API documentation for Nactive Health
//	@termsOfService	http://nactive-health.com/terms/

//	@contact.name	Support
//	@contact.email	support@nactive-health.com

// @host		localhost:8019
// @BasePath	/api/v1
func Setup(logger *logutil.Logger, validator *validator.Validate, db *db.Database, appConfiguration *config.App) *gin.Engine {
	if appConfiguration.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()

	// Middlewares
	// r.Use(gin.Logger())
	r.ForwardedByClientIP = true
	r.SetTrustedProxies(config.GetConfig().Server.TrustedProxies)
	r.Use(middleware.CORS())
	r.Use(middleware.RateLimiter())
	r.Use(middleware.Security())
	r.Use(middleware.Logger())
	r.Use(middleware.MetricsLogger(db.Postgresql.DB()))
	r.Use(gin.Recovery())
	r.Use(gzip.Gzip(gzip.DefaultCompression))
	r.MaxMultipartMemory = 1 << 20 // 1MB

	// Static files
	wd, _ := os.Getwd() // ignore error; fall back to relative if needed
	assetsDir := filepath.Join(wd, "templates")

	// Serve files under {abs}/templates at /assets
	r.Static("/assets", assetsDir)
	//Then your email can point to https://<host>/assets/logo.png.

	// routers
	ApiVersion := "api/v1"
	// routers for health
	HttpHealth(r, ApiVersion, validator, db, logger)

	// routers for  users
	Auth(r, ApiVersion, validator, db, logger)
	Profile(r, ApiVersion, validator, db, logger)

	// admin
	AdminCategory(r, ApiVersion, validator, db)
	AdminNews(r, ApiVersion, validator, db)
	AdminSettings(r, ApiVersion, validator, db, logger)

	// Blog
	BlogCategory(r, ApiVersion, validator, db)
	BlogNews(r, ApiVersion, validator, db)
	AdminImages(r, ApiVersion, validator, db)

	UsageLog(r, ApiVersion, validator, db)
	Contact(r, ApiVersion, validator, db)

	// for testing
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "Nactive Health Boilerplate",
			"status":  http.StatusOK,
		})
	})

	// Swagger endpoint
	r.GET("/api/docs/*any", middleware.SwaggerMiddleware(), ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"name":    "Not Found",
			"message": "Page not found.",
			"code":    404,
			"status":  http.StatusNotFound,
		})
	})

	return r
}

func LogRouteCounts(r *gin.Engine, base string) {
	rts := r.Routes()
	total := 0
	byMethod := map[string]int{}
	byArea := map[string]int{} // provider/, doctor/, admin/, etc.

	for _, rt := range rts {
		if base == "" || strings.HasPrefix(rt.Path, "/"+base) || strings.HasPrefix(rt.Path, base) {
			total++
			byMethod[rt.Method]++

			// crude “area” bucket after /api/v1/
			area := "root"
			p := strings.TrimPrefix(rt.Path, "/"+base+"/")
			if i := strings.Index(p, "/"); i > -1 {
				area = p[:i] // e.g. provider, doctor, admin
			} else if p != "" {
				area = p
			}
			byArea[area]++
		}
	}

	log.Printf("[Routes] Total under %s: %d", base, total)
	for m, c := range byMethod {
		log.Printf("[Routes]   %-6s %d", m, c)
	}
	for a, c := range byArea {
		log.Printf("[Routes]   %-12s %d", a, c)
	}
}
