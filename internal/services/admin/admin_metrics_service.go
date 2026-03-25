package adminmetricsservices

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/snappy-fix-golang/internal/domain/entities"
	"github.com/snappy-fix-golang/internal/inst"
	"gorm.io/gorm"
)

// GetPlatformMetricsService aggregates a metrics snapshot + SLA comparison
// into a ready-to-return response payload for the dashboard.
func GetPlatformMetricsService(db *gorm.DB, windowDays int) (gin.H, int, error) {
	pdb := inst.InitDB(db)

	sla := entities.DefaultPlatformSLA()
	if windowDays <= 0 {
		windowDays = sla.WindowDaysDefault
	}

	snapshot, err := entities.GetLatestMetrics(pdb, windowDays)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// If you have no data yet, return zeros but still include SLA targets.
			empty := gin.H{
				"uptime": gin.H{
					"metric":      "Uptime",
					"description": "Percentage of platform uptime.",
					"value":       0.0,
					"sla_target":  sla.UptimeTargetPercent,
					"status":      "UNKNOWN",
				},
				"latency": gin.H{
					"metric":      "API Latency",
					"description": "Average API response time (ms).",
					"value":       0.0,
					"sla_target":  sla.MaxLatencyMs,
					"status":      "UNKNOWN",
				},
				"error_rate": gin.H{
					"metric":      "Error Rate",
					"description": "Percentage of failed transactions.",
					"value":       0.0,
					"sla_target":  sla.MaxErrorRatePercent,
					"status":      "UNKNOWN",
				},
				"downtime_incidents": gin.H{
					"metric":      "Downtime Incidents",
					"description": "Number of downtime occurrences.",
					"value":       0,
					"sla_target":  sla.MaxDowntimeIncidents,
					"status":      "UNKNOWN",
				},
				"window": gin.H{
					"days": windowDays,
				},
			}
			return empty, http.StatusOK, nil
		}
		// real DB error
		return gin.H{}, http.StatusInternalServerError, err
	}

	// Compute statuses against SLA
	uptimeStatus := statusHigherIsBetter(snapshot.UptimePercent, sla.UptimeTargetPercent)
	latencyStatus := statusLowerIsBetter(snapshot.AvgLatencyMs, sla.MaxLatencyMs)
	errorRateStatus := statusLowerIsBetter(snapshot.ErrorRatePercent, sla.MaxErrorRatePercent)
	downtimeStatus := statusLowerIsBetterFloat(float64(snapshot.DowntimeIncidents), float64(sla.MaxDowntimeIncidents))

	payload := gin.H{
		"uptime": gin.H{
			"metric":      "Uptime",
			"description": "Percentage of platform uptime.",
			"value":       snapshot.UptimePercent,
			"sla_target":  sla.UptimeTargetPercent,
			"status":      uptimeStatus,
		},
		"latency": gin.H{
			"metric":      "API Latency",
			"description": "Average API response time (ms).",
			"value":       snapshot.AvgLatencyMs,
			"sla_target":  sla.MaxLatencyMs,
			"status":      latencyStatus,
		},
		"error_rate": gin.H{
			"metric":      "Error Rate",
			"description": "Percentage of failed transactions.",
			"value":       snapshot.ErrorRatePercent,
			"sla_target":  sla.MaxErrorRatePercent,
			"status":      errorRateStatus,
		},
		"downtime_incidents": gin.H{
			"metric":      "Downtime Incidents",
			"description": "Number of downtime occurrences.",
			"value":       snapshot.DowntimeIncidents,
			"sla_target":  sla.MaxDowntimeIncidents,
			"status":      downtimeStatus,
		},
		"window": gin.H{
			"start": snapshot.WindowStart,
			"end":   snapshot.WindowEnd,
			"days":  windowDays,
		},
	}

	return payload, http.StatusOK, nil
}

// ---- SLA helpers ----

func statusHigherIsBetter(value, target float64) string {
	if value == 0 && target == 0 {
		return "UNKNOWN"
	}
	if value >= target {
		return "GOOD"
	}
	return "BAD"
}

func statusLowerIsBetter(value, max float64) string {
	if value == 0 && max == 0 {
		return "UNKNOWN"
	}
	if value <= max {
		return "GOOD"
	}
	return "BAD"
}

func statusLowerIsBetterFloat(value, max float64) string {
	if value == 0 && max == 0 {
		return "UNKNOWN"
	}
	if value <= max {
		return "GOOD"
	}
	return "BAD"
}
