package activity

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/domain/enums"
)

func UserActivityLog(db repository.DatabaseManager, userID uuid.UUID, actorType enums.UserType, subject, message string) error {
	entry := &entities.UserActivityLog{
		ID:        uuid.Must(uuid.NewV4()),
		UserID:    userID,
		ActorType: string(actorType),
		Subject:   subject,
		Message:   message,
		CreatedAt: time.Now().UTC(),
	}
	return db.CreateOneRecord(entry)
}
