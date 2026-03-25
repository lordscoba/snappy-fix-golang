package migrations

import (
	"context"

	"gorm.io/gorm"
)

// Define a custom type for clarity
type MigrationStep struct {
	Name string
	Fn   func(tx *gorm.DB) error
}

func GetMigrationSteps(ctx context.Context) []MigrationStep {
	return []MigrationStep{
		{
			Name: "20260310_001_initial_setup",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				// CALLING THE HELPER HERE:
				// This is where your Auth models actually get created/synced
				return MigrateModels(ctx, tx, AuthMigrationModels(), AlterColumnModels())
			},
		},
		{
			Name: "20260310_002_drop_legacy_roles",
			Fn: func(tx *gorm.DB) error {
				return DropTables(tx, DropModels())
			},
		},
		{
			Name: "20260310_003_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContentMigrationModels(), nil)
			},
		},
		{
			Name: "20260310_004_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContentMigrationModels(), nil)
			},
		},
		{
			Name: "20260310_04_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContentMigrationModels(), nil)
			},
		},
		{
			Name: "20260310_005_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContentMigrationModels(), nil)
			},
		},
		{
			Name: "20260310_006_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContentMigrationModels(), nil)
			},
		},
		{
			Name: "20260310_007_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContentMigrationModels(), nil)
			},
		},
		{
			Name: "20260310_008_content_tables",
			Fn: func(tx *gorm.DB) error {
				if err := EnsureUUIDExt(tx); err != nil {
					return err
				}

				return MigrateModels(ctx, tx, ContactMigrationModels(), nil)
			},
		},
	}
}
