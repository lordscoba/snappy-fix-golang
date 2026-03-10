package entities

import (
	"time"

	"github.com/gofrs/uuid"
	"github.com/snappy-fix-golang/internal/adapters/repository"
)

// PlatformMetricsSnapshot represents one aggregated metrics window
// e.g. "last 30 days".
type PlatformMetricsSnapshot struct {
	ID                 uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	WindowStart        time.Time `gorm:"index;not null" json:"window_start"`
	WindowEnd          time.Time `gorm:"index;not null" json:"window_end"`
	UptimePercent      float64   `json:"uptime_percent"`
	AvgLatencyMs       float64   `json:"avg_latency_ms"`
	ErrorRatePercent   float64   `json:"error_rate_percent"`
	DowntimeIncidents  int64     `json:"downtime_incidents"`
	TotalRequests      int64     `json:"total_requests"`
	TotalErrorRequests int64     `json:"total_error_requests"`
	CreatedAt          time.Time `json:"created_at"`
}

func (PlatformMetricsSnapshot) TableName() string {
	return "platform_metrics_snapshots"
}

// ----- SLA Config (static for now – can later move to DB if you want) -----

type PlatformSLAConfig struct {
	UptimeTargetPercent  float64 // e.g. 99.5
	MaxLatencyMs         float64 // e.g. 300
	MaxErrorRatePercent  float64 // e.g. 1.0
	MaxDowntimeIncidents int64   // e.g. 3
	WindowDaysDefault    int     // e.g. 30
}

func DefaultPlatformSLA() PlatformSLAConfig {
	return PlatformSLAConfig{
		UptimeTargetPercent:  99.5,
		MaxLatencyMs:         300,
		MaxErrorRatePercent:  1.0,
		MaxDowntimeIncidents: 3,
		WindowDaysDefault:    30,
	}
}

// ----- Repository helpers -----

// GetLatestMetrics returns the most recent snapshot (optionally scoped by windowStart).
// If no record exists it returns gorm.ErrRecordNotFound.
func GetLatestMetrics(db repository.DatabaseManager, windowDays int) (PlatformMetricsSnapshot, error) {
	var out PlatformMetricsSnapshot

	query := db.DB().Order("window_end DESC")

	if windowDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -windowDays)
		query = query.Where("window_end >= ?", cutoff)
	}

	if err := query.First(&out).Error; err != nil {
		return out, err
	}
	return out, nil
}

func UpsertMetricsSnapshot(
	db repository.DatabaseManager,
	windowStart, windowEnd time.Time,
	uptime, latency, errorRate float64,
	downtimeIncidents, totalReq, totalErr int64,
) (PlatformMetricsSnapshot, error) {

	var existing PlatformMetricsSnapshot
	err := db.DB().
		Where("window_start = ? AND window_end = ?", windowStart, windowEnd).
		First(&existing).Error

	now := time.Now().UTC()

	if err != nil {
		// Create new snapshot
		snap := PlatformMetricsSnapshot{
			ID:                 uuid.Must(uuid.NewV4()),
			WindowStart:        windowStart,
			WindowEnd:          windowEnd,
			UptimePercent:      uptime,
			AvgLatencyMs:       latency,
			ErrorRatePercent:   errorRate,
			DowntimeIncidents:  downtimeIncidents,
			TotalRequests:      totalReq,
			TotalErrorRequests: totalErr,
			CreatedAt:          now,
		}
		if err := db.CreateOneRecord(&snap); err != nil {
			return snap, err
		}
		return snap, nil
	}

	// Update existing
	updates := map[string]interface{}{
		"uptime_percent":       uptime,
		"avg_latency_ms":       latency,
		"error_rate_percent":   errorRate,
		"downtime_incidents":   downtimeIncidents,
		"total_requests":       totalReq,
		"total_error_requests": totalErr,
	}
	if _, err := db.UpdateFields(&existing, updates, existing.ID.String()); err != nil {
		return existing, err
	}

	return existing, nil
}
