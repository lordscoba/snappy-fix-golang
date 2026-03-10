package workers

import (
	"time"

	"github.com/snappy-fix-golang/internal/adapters/repository"
	metricssvc "github.com/snappy-fix-golang/internal/services/metrics"
)

func RunRolling30DayMetrics(db repository.DatabaseManager) error {
	windowEnd := time.Now().UTC()
	windowStart := windowEnd.AddDate(0, 0, -30)

	_, err := metricssvc.AggregateAndUpsertPlatformMetrics(db, windowStart, windowEnd)
	return err
}
