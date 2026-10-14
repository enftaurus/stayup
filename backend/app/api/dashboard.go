package api

import (
	"net/http"

	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
)

func GetDashboardSummary(c *gin.Context) {
	userID := GetUserID(c)
	ctx := c.Request.Context()

	query := `
		WITH user_monitors AS (
			SELECT m.id, s.status
			FROM monitors m
			JOIN monitor_state s ON s.monitor_id = m.id
			WHERE m.user_id = $1::uuid
		),
		stats_24h AS (
			SELECT
				ROUND(AVG(ch.response_time_ms))::int AS avg_response_ms,
				ROUND(COUNT(*) FILTER (WHERE ch.status = 'UP')::numeric / GREATEST(1, COUNT(*)) * 100, 2) AS uptime_24h
			FROM check_history ch
			JOIN monitors m ON m.id = ch.monitor_id
			WHERE m.user_id = $1::uuid
			  AND ch.checked_at >= NOW() - INTERVAL '24 hours'
		),
		open_inc AS (
			SELECT COUNT(*)::int AS count
			FROM incidents i
			JOIN monitors m ON m.id = i.monitor_id
			WHERE m.user_id = $1::uuid
			  AND i.resolved_at IS NULL
		)
		SELECT
			(SELECT COUNT(*)::int FROM user_monitors) AS total_monitors,
			(SELECT COUNT(*) FILTER (WHERE status = 'UP')::int FROM user_monitors) AS up_count,
			(SELECT COUNT(*) FILTER (WHERE status = 'DOWN')::int FROM user_monitors) AS down_count,
			(SELECT COUNT(*) FILTER (WHERE status = 'PAUSED')::int FROM user_monitors) AS paused_count,
			(SELECT COUNT(*) FILTER (WHERE status = 'UNKNOWN')::int FROM user_monitors) AS unknown_count,
			(SELECT count FROM open_inc) AS open_incidents,
			COALESCE((SELECT avg_response_ms FROM stats_24h), 0) AS avg_response_time_ms,
			COALESCE((SELECT uptime_24h FROM stats_24h), 100.0) AS uptime_24h;
	`

	var total, up, down, paused, unknown, openIncidents, avgResponse int
	var uptime24h float64

	err := utils.DB.QueryRow(ctx, query, userID).Scan(
		&total, &up, &down, &paused, &unknown, &openIncidents, &avgResponse, &uptime24h,
	)

	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query dashboard summary", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_monitors":       total,
		"up_count":             up,
		"down_count":           down,
		"paused_count":         paused,
		"unknown_count":        unknown,
		"open_incidents":       openIncidents,
		"avg_response_time_ms": avgResponse,
		"uptime_24h":           uptime24h,
	})
}
