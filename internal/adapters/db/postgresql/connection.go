package postgresql

import (
	"fmt"
	"os"

	"log"

	"github.com/snappy-fix-golang/internal/adapters"
	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/config"
	logutil "github.com/snappy-fix-golang/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	lg "gorm.io/gorm/logger"
)

func ConnectToDatabase(logger *logutil.Logger, configDatabases config.Database) *gorm.DB {
	dbsCV := configDatabases

	logutil.LogAndPrint(logger, "connecting to database")
	connectedDB := connectToDb(dbsCV.DB_HOST, dbsCV.USERNAME, dbsCV.PASSWORD, dbsCV.DB_NAME, dbsCV.DB_PORT, dbsCV.SSLMODE, dbsCV.TIMEZONE, logger)

	logutil.LogAndPrint(logger, "connected to database")
	db.DB.Postgresql = NewPostgresqlConnection(connectedDB)
	return connectedDB
}

func connectToDb(host, user, password, dbname, port, sslmode, timezone string, logger *logutil.Logger) *gorm.DB {
	port = adapters.ResolvePortParsing(port, logger)

	// 1. Added prepare_threshold=0 to bypass driver-level caching
	dsn := fmt.Sprintf("host=%v user=%v password=%v dbname=%v port=%v sslmode=%v TimeZone=%v prepare_threshold=0",
		host, user, password, dbname, port, sslmode, timezone)

	newLogger := lg.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		lg.Config{
			LogLevel:                  lg.Error,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
		// 2. Explicitly disable GORM-level prepared statement caching
		PrepareStmt: false,
	})

	if err != nil {
		logutil.LogAndPrint(logger, fmt.Sprintf("connection to %v db failed with: %v", dbname, err))
		panic(err)
	}

	logutil.LogAndPrint(logger, fmt.Sprintf("connected to %v db", dbname))
	return db
}
