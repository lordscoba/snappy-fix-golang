package cronjobs

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/snappy-fix-golang/internal/adapters/repository"
	"github.com/snappy-fix-golang/internal/workers"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func StartCronJobs(db repository.DatabaseManager, logger *logutil.Logger) {
	s := gocron.NewScheduler(time.UTC)

	s.Every(1).Day().At("01:00").Do(func() {
		start := time.Now().UTC()
		duration := time.Since(start)

		err := workers.RunRolling30DayMetrics(db)
		if err != nil {
			logger.Error("metrics cron failed", "job", "metrics_cron", "error", err.Error(), "duration_ms", duration.Milliseconds())
			fmt.Println("metrics cron failed:", err)
		} else {
			logger.Info("metrics cron completed", "job", "metrics_cron", "duration_ms", duration.Milliseconds())
			fmt.Println("metrics cron completed")
		}
	})
	s.StartAsync()
}
