package entities

import "time"

type MigrationLog struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex;not null"`
	Status    string    `gorm:"size:20"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}
