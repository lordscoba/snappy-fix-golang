package metrics

import (
	"time"

	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/domain/entities"
)

// AggregateAndUpsertPlatformMetrics aggregates request logs within [windowStart, windowEnd)
// and stores a snapshot using UpsertMetricsSnapshot.
func AggregateAndUpsertPlatformMetrics(
	db repository.DatabaseManager,
	windowStart, windowEnd time.Time,
) (entities.PlatformMetricsSnapshot, error) {

	gdb := db.DB()

	// 1) Total requests in window
	var totalReq int64
	if err := gdb.Model(&entities.ApiRequestMetric{}).
		Where("created_at >= ? AND created_at < ?", windowStart, windowEnd).
		Count(&totalReq).Error; err != nil {
		return entities.PlatformMetricsSnapshot{}, err
	}

	// Default "empty" snapshot
	if totalReq == 0 {
		return entities.UpsertMetricsSnapshot(
			db,
			windowStart,
			windowEnd,
			100.0, // assume 100% uptime if no requests
			0.0,
			0.0,
			0,
			0,
			0,
		)
	}

	// 2) Average latency + error count
	type aggRow struct {
		AvgLatency float64
		ErrorCount int64
	}

	var agg aggRow
	if err := gdb.Model(&entities.ApiRequestMetric{}).
		Select(`
			AVG(latency_ms) AS avg_latency,
			SUM(CASE WHEN is_error THEN 1 ELSE 0 END) AS error_count
		`).
		Where("created_at >= ? AND created_at < ?", windowStart, windowEnd).
		Scan(&agg).Error; err != nil {
		return entities.PlatformMetricsSnapshot{}, err
	}

	// 3) Error rate
	errorRate := 0.0
	if totalReq > 0 {
		errorRate = (float64(agg.ErrorCount) * 100.0) / float64(totalReq)
	}

	// 4) Uptime (simple approximation: 100% - errorRate)
	uptime := 100.0
	if errorRate > 0 {
		uptime = 100.0 - errorRate
	}

	// 5) Downtime incidents (distinct minutes that had any error)
	var downtimeIncidents int64
	if err := gdb.Model(&entities.ApiRequestMetric{}).
		Select("COUNT(DISTINCT date_trunc('minute', created_at))").
		Where("created_at >= ? AND created_at < ? AND is_error = TRUE", windowStart, windowEnd).
		Count(&downtimeIncidents).Error; err != nil {
		return entities.PlatformMetricsSnapshot{}, err
	}

	// 6) Upsert snapshot
	return entities.UpsertMetricsSnapshot(
		db,
		windowStart,
		windowEnd,
		uptime,
		agg.AvgLatency,
		errorRate,
		downtimeIncidents,
		totalReq,
		agg.ErrorCount,
	)
}
