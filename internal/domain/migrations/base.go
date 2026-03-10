package migrations

import (
	"context"
	"fmt"
	"log"

	"github.com/snappy-fix-golang/internal/adapters/db"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"gorm.io/gorm"
)

func RunAllMigrations(ctx context.Context, database *db.Database) error {
	pg := database.Postgresql.DB()

	// 1. Sanity checks
	if err := Ping(pg); err != nil {
		return fmt.Errorf("db ping failed: %w", err)
	}
	if err := LogCurrentDB(pg); err != nil {
		return fmt.Errorf("failed to read current db/schema: %w", err)
	}

	// 2. Initialize tracking table
	if err := pg.AutoMigrate(&entities.MigrationLog{}); err != nil {
		return fmt.Errorf("failed to create migration log table: %w", err)
	}

	steps := GetMigrationSteps(ctx)

	// 4. Execution Loop (Exactly as you had it)
	for _, step := range steps {
		var logEntry entities.MigrationLog
		err := pg.Where("name = ? AND status = ?", step.Name, "Success").First(&logEntry).Error

		if err == nil {
			continue // Skip if already done
		}

		log.Printf("🚀 Applying migration: %s", step.Name)

		err = pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			if err := step.Fn(tx); err != nil {
				return err
			}
			return tx.Create(&entities.MigrationLog{Name: step.Name, Status: "Success"}).Error
		})

		if err != nil {
			return fmt.Errorf("migration failed at step %s: %w", step.Name, err)
		}
	}

	log.Println("✅ All migrations processed successfully")
	return nil
}
func MigrateModels(ctx context.Context, tx *gorm.DB, models []interface{}, alterCols []AlterColumn) error {
	if tx == nil {
		return fmt.Errorf("migration failed: database transaction is nil")
	}

	// Ensure we respect the context (timeout/cancelation)
	tx = tx.WithContext(ctx)

	// Auto migrate models
	if len(models) > 0 {
		if err := tx.AutoMigrate(models...); err != nil {
			return fmt.Errorf("automigrate failed: %w", err)
		}
	}

	// Run manual alterations
	for _, col := range alterCols {
		log.Printf("Altering column %s in table %s...", col.Column, col.TableName)

		if err := col.UpdateColumnType(tx); err != nil {
			return fmt.Errorf(
				"failed altering column %s in table %s: %w",
				col.Column,
				col.TableName,
				err,
			)
		}
	}

	return nil
}

// func MigrateModels(ctx context.Context, tx *gorm.DB, models []interface{}, alterCols []AlterColumn) error {
// 	// 1. Bulk migrate models first
// 	if err := tx.AutoMigrate(models...); err != nil {
// 		return fmt.Errorf("automigrate failed: %w", err)
// 	}

// 	// 2. Run manual alterations
// 	for _, col := range alterCols {
// 		log.Printf("Altering column %s in table %s...", col.Column, col.TableName)
// 		if err := col.UpdateColumnType(tx); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// RunAllMigrations runs the full migration pipeline and FAILS FAST on error.
// func RunAllMigrations(ctx context.Context, database *db.Database) error {
// 	pg := database.Postgresql.DB()

// 	// Sanity check connectivity
// 	if err := Ping(pg); err != nil {
// 		return fmt.Errorf("db ping failed: %w", err)
// 	}

// 	// Log where we are migrating (helps catch DSN/schema mismatches)
// 	if err := LogCurrentDB(pg); err != nil {
// 		return fmt.Errorf("failed to read current db/schema: %w", err)
// 	}

// 	// Enable uuid extension BEFORE any table using uuid_generate_v4()
// 	if err := EnsureUUIDExt(pg); err != nil {
// 		return err
// 	}

// 	return pg.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
// 		if err := EnsureUUIDExt(tx); err != nil {
// 			return err
// 		}

// 		// Drop (optional; order-aware if you pass children before parents)
// 		if err := DropTables(pg, DropModels()); err != nil {
// 			return fmt.Errorf("drop tables failed: %w", err)
// 		}

// 		if err := MigrateModels(ctx, tx, AuthMigrationModels(), AlterColumnModels()); err != nil {
// 			return err
// 		}

// 		log.Println("migrations completed successfully")

// 		return nil // Commits if nil is returned
// 	})

// }
