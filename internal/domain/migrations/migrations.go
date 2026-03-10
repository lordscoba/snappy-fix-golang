package migrations

import "github.com/snappy-fix-golang/internal/domain/entities"

func AuthMigrationModels() []interface{} {
	return []interface{}{
		// Users and patient-side first (referenced by FKs)
		entities.User{},
		entities.AccessToken{},
		entities.PasswordReset{},
		entities.EmailVerification{},
		entities.UserActivityLog{},
		entities.ApiRequestMetric{},
		entities.PlatformMetricsSnapshot{},
	}

}

func ContentMigrationModels() []interface{} {
	return []interface{}{
		entities.Category{},
		entities.News{},
		entities.NewsImage{},
	}
}

// Only models you want to DROP go here.
// If you add child tables later, list them BEFORE the parent (MedicalHistory).
func DropModels() []interface{} {
	return []interface{}{
		// entities.Role{},
	}
}

func AlterColumnModels() []AlterColumn {
	return []AlterColumn{}
}
