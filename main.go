package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/domain/migrations"
	"github.com/snappy-fix-golang/internal/inst"

	"github.com/snappy-fix-golang/internal/config"
	"github.com/snappy-fix-golang/internal/router"
	cronjobs "github.com/snappy-fix-golang/pkg/cron_jobs"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func main() {

	// Initialize the logger and handle the error
	logger, err := logutil.NewLogger() //Warning !!!!! Do not recreate this action anywhere on the app
	if err != nil {
		// If logger initialization fails, we can't use it.
		// Panic here so the application fails fast.
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	// Config
	configuration := config.Setup(logger, "./app")

	// DB connect
	postgresql.ConnectToDatabase(logger, configuration.Database)
	conn := db.Connection()

	// Run migrations (FAIL FAST)
	if configuration.Database.Migrate {
		log.Println("Starting DB migration...")
		if err := migrations.RunAllMigrations(context.Background(), conn); err != nil {
			log.Fatalf("Migration failed: %v", err) // abort startup
		}
		log.Println("DB migration completed.")
	}

	// Validator + Router
	validatorRef := validator.New()
	r := router.Setup(logger, validatorRef, conn, &configuration.App)

	// for cron jobs
	pgDB := conn.Postgresql.DB()
	pdb := inst.InitDB(pgDB)
	cronjobs.StartCronJobs(pdb, logger)

	fmt.Printf("Total APIs: %d\n", len(r.Routes()))
	logutil.LogAndPrint(logger, fmt.Sprintf("Server is starting at 127.0.0.1:%s", configuration.Server.Port))
	log.Fatal(r.Run(":" + configuration.Server.Port))
}
