package db

import (
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/config"
	logutil "github.com/snappy-fix-golang/pkg/logger"
	"gorm.io/gorm"
)

type DbConnection interface {
	NewDatabaseConnection(db *gorm.DB, logger *logutil.Logger, config *config.Database) *Database
}

type Database struct {
	Postgresql repository.DatabaseManager
	Redis      repository.CacheManager
}

var DB *Database = &Database{}

func Connection() *Database {
	return DB
}
