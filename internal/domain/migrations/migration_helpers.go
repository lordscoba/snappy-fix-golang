package migrations

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func Ping(db *gorm.DB) error {
	var one int
	return db.Raw("SELECT 1").Scan(&one).Error
}

func LogCurrentDB(db *gorm.DB) error {
	var currentDB, currentSchema string
	row := db.Raw(`SELECT current_database(), current_schema()`).Row()
	if err := row.Scan(&currentDB, &currentSchema); err != nil {
		return err
	}
	log.Printf("Migrating on database=%q schema=%q", currentDB, currentSchema)
	return nil
}

func EnsureUUIDExt(db *gorm.DB) error {
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf(`ensuring "uuid-ossp" extension failed: %w`, err)
	}
	return nil
}
