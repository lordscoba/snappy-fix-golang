package entities

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/repository"
)

type ApiRequestMetric struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Path       string    `gorm:"index;not null" json:"path"`
	Method     string    `gorm:"type:varchar(8);index;not null" json:"method"`
	StatusCode int       `gorm:"index;not null" json:"status_code"`
	LatencyMs  float64   `json:"latency_ms"`
	IsError    bool      `gorm:"index" json:"is_error"`
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
}

func (m *ApiRequestMetric) Create(db repository.DatabaseManager) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.Must(uuid.NewV4())
	}
	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now().UTC()
	}
	return db.CreateOneRecord(m)
}
