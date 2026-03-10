package entities

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/db/postgresql"
	"github.com/snappy-fix-golang/internal/adapters/repository"
)

type UserActivityLog struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	ActorType string    `gorm:"type:varchar(96)"                                json:"actor_type"`
	UserID    uuid.UUID `gorm:"type:uuid;index;not null" json:"actor_id"`
	Subject   string    `gorm:"type:varchar(64);index" json:"subject"` // e.g., "APPOINTMENT"
	Message   string    `gorm:"type:text" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (m *UserActivityLog) CreateUserActivityLog(db repository.DatabaseManager) error {
	err := db.CreateOneRecord(&m)
	if err != nil {
		return err
	}
	return nil
}

// List (paginated newest-first)
func (m *UserActivityLog) ListByUserID(db repository.DatabaseManager, userID string, c *gin.Context) ([]UserActivityLog, repository.PaginationResponse, error) {
	var out []UserActivityLog
	p := postgresql.GetPagination(c)
	pag, err := db.SelectAllFromDbOrderByPaginated(
		"created_at", "desc", "",
		p, &out, "user_id = ?", userID,
	)
	return out, pag, err
}

// TimeUTC is a thin wrapper if you use custom JSON formatting; otherwise you can drop it.
type TimeUTC = time.Time // or time.Time if you don't have a custom type
