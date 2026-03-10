package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

type AlterColumn struct {
	Model     interface{}
	TableName string
	Column    string
	Type      string
}

func (a *AlterColumn) UpdateColumnType(db *gorm.DB) error {
	m := db.Migrator()

	// Check if column exists before trying to alter it
	if !m.HasColumn(a.Model, a.Column) {
		return fmt.Errorf("column %s does not exist on table %s", a.Column, a.TableName)
	}

	// Use GORM's native AlterColumn which handles the underlying SQL safely
	return m.AlterColumn(a.Model, a.Column)
}

// func (a *AlterColumn) UpdateColumnType(db *gorm.DB) error {
// 	if err := db.Exec(fmt.Sprintf(
// 		"ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s",
// 		a.TableName, a.Column, a.Type, a.Column, a.Type,
// 	)).Error; err != nil {
// 		return err
// 	}
// 	// keep GORM’s metadata aligned
// 	return db.Migrator().AlterColumn(a.Model, a.Column)
// }
