package migrations

import (
	"gorm.io/gorm"
)

// DropTables drops all given models if they exist.
// Pass child tables BEFORE parent tables when FKs exist.
func DropTables(db *gorm.DB, models []interface{}) error {
	for _, m := range models {
		if db.Migrator().HasTable(m) {
			if err := db.Migrator().DropTable(m); err != nil {
				return err
			}
		}
	}
	return nil
}

// Optional convenience: drop by table name (raw)
func DropTableByName(db *gorm.DB, table string) error {
	return db.Exec("DROP TABLE IF EXISTS " + table + " CASCADE").Error
}
