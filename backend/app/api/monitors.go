package api

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/enftaurus/stayup/config"
	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type CreateMonitorReq struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	IntervalSeconds *int   `json:"interval_seconds"`
	ExpectedStatus  *int   `json:"expected_status"`
}

type UpdateMonitorReq struct {
	Name            *string `json:"name"`
	URL             *string `json:"url"`
	IntervalSeconds *int    `json:"interval_seconds"`
	ExpectedStatus  *int    `json:"expected_status"`
}

func allowPrivateSSRF() bool {
	return strings.ToLower(config.GetEnv("SSRF_ALLOW_PRIVATE")) == "true"
}

func CreateMonitor(c *gin.Context) {
	userID := GetUserID(c)
	var req CreateMonitorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", "")
		return
	}

	name := strings.TrimSpace(req.Name)
	if len(name) < 1 || len(name) > 255 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Name must be between 1 and 255 characters", "name")
		return
	}

	rawURL := strings.TrimSpace(req.URL)
	if _, err := utils.ValidateURL(rawURL, allowPrivateSSRF()); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "url")
		return
	}

	interval := 60
	if req.IntervalSeconds != nil {
		interval = *req.IntervalSeconds
	}
	if interval < 60 || interval > 3600 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Interval must be between 60 and 3600 seconds", "interval_seconds")
		return
	}

	expectedStatus := 200
	if req.ExpectedStatus != nil {
		expectedStatus = *req.ExpectedStatus
	}
	if expectedStatus < 100 || expectedStatus > 599 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Expected status must be between 100 and 599", "expected_status")
		return
	}

	ctx := c.Request.Context()
	tx, err := utils.DB.Begin(ctx)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction begin error", "")
		return
	}
	defer tx.Rollback(ctx)

	var monitorID string
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx, `
		INSERT INTO monitors (
			user_id, name, url, method, expected_status, interval_seconds, timeout_seconds, enabled, created_at, updated_at
		)
		VALUES ($1::uuid, $2, $3, 'HEAD', $4, $5, 10, true, NOW(), NOW())
		RETURNING id::text, created_at, updated_at
	`, userID, name, rawURL, expectedStatus, interval).Scan(&monitorID, &createdAt, &updatedAt)

	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create monitor", "")
		return
	}

	var nextCheckAt time.Time
	var stateStatus string
	err = tx.QueryRow(ctx, `
		INSERT INTO monitor_state (monitor_id, status, next_check_at, updated_at)
		VALUES ($1::uuid, 'UNKNOWN', NOW(), NOW())
		RETURNING status::text, next_check_at
	`, monitorID).Scan(&stateStatus, &nextCheckAt)

	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to initialize monitor state", "")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to commit transaction", "")
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":               monitorID,
		"name":             name,
		"url":              rawURL,
		"method":           "HEAD",
		"expected_status":  expectedStatus,
		"interval_seconds": interval,
		"timeout_seconds":  10,
		"enabled":          true,
		"created_at":       createdAt,
		"updated_at":       updatedAt,
		"state": gin.H{
			"status":        stateStatus,
			"next_check_at": nextCheckAt,
		},
	})
}

func ListMonitors(c *gin.Context) {
	userID := GetUserID(c)
	statusFilter := strings.ToUpper(strings.TrimSpace(c.Query("status")))

	ctx := c.Request.Context()
	query := `
		WITH sparklines AS (
			SELECT ch.monitor_id,
			       ARRAY_AGG(ch.response_time_ms ORDER BY ch.checked_at ASC) AS response_times
			FROM (
				SELECT monitor_id, response_time_ms, checked_at,
				       ROW_NUMBER() OVER (PARTITION BY monitor_id ORDER BY checked_at DESC) as rn
				FROM check_history
			) ch
			WHERE ch.rn <= 30
			GROUP BY ch.monitor_id
		),
		uptimes AS (
			SELECT monitor_id,
			       ROUND(COUNT(*) FILTER (WHERE status = 'UP')::numeric / GREATEST(1, COUNT(*)) * 100, 2) AS uptime_24h
			FROM check_history
			WHERE checked_at >= NOW() - INTERVAL '24 hours'
			GROUP BY monitor_id
		)
		SELECT m.id::text, m.name, m.url, m.method::text, m.expected_status, m.interval_seconds, m.enabled, m.created_at,
		       s.status::text, s.last_checked_at, s.last_response_time_ms, s.last_status_code, s.last_error, s.next_check_at,
		       COALESCE(u.uptime_24h, 100.0) AS uptime_24h,
		       COALESCE(sp.response_times, '{}'::int[]) AS sparkline
		FROM monitors m
		JOIN monitor_state s ON s.monitor_id = m.id
		LEFT JOIN uptimes u ON u.monitor_id = m.id
		LEFT JOIN sparklines sp ON sp.monitor_id = m.id
		WHERE m.user_id = $1::uuid
		  AND ($2 = '' OR s.status::text = $2)
		ORDER BY m.created_at DESC;
	`

	rows, err := utils.DB.Query(ctx, query, userID, statusFilter)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("Query error: %v", err), "")
		return
	}
	defer rows.Close()

	type MonitorItem struct {
		ID                 string    `json:"id"`
		Name               string    `json:"name"`
		URL                string    `json:"url"`
		Method             string    `json:"method"`
		ExpectedStatus     int       `json:"expected_status"`
		IntervalSeconds    int       `json:"interval_seconds"`
		Enabled            bool      `json:"enabled"`
		CreatedAt          time.Time `json:"created_at"`
		Status             string    `json:"status"`
		LastCheckedAt      *time.Time `json:"last_checked_at"`
		LastResponseTimeMs *int      `json:"last_response_time_ms"`
		LastStatusCode     *int      `json:"last_status_code"`
		LastError          *string   `json:"last_error"`
		NextCheckAt        *time.Time `json:"next_check_at"`
		Uptime24h          float64   `json:"uptime_24h"`
		Sparkline          []int     `json:"sparkline"`
	}

	monitors := []MonitorItem{}
	for rows.Next() {
		var item MonitorItem
		var rawSpark []int32
		if err := rows.Scan(
			&item.ID, &item.Name, &item.URL, &item.Method, &item.ExpectedStatus, &item.IntervalSeconds, &item.Enabled, &item.CreatedAt,
			&item.Status, &item.LastCheckedAt, &item.LastResponseTimeMs, &item.LastStatusCode, &item.LastError, &item.NextCheckAt,
			&item.Uptime24h, &rawSpark,
		); err != nil {
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error parsing row", "")
			return
		}
		item.Sparkline = make([]int, len(rawSpark))
		for i, v := range rawSpark {
			item.Sparkline[i] = int(v)
		}
		monitors = append(monitors, item)
	}

	c.JSON(http.StatusOK, gin.H{"monitors": monitors})
}

func GetMonitor(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")
	ctx := c.Request.Context()

	query := `
		SELECT m.id::text, m.name, m.url, m.method::text, m.expected_status, m.interval_seconds, m.timeout_seconds, m.enabled, m.created_at, m.updated_at,
		       s.status::text, s.last_checked_at, s.last_response_time_ms, s.last_status_code, s.last_error,
		       s.consecutive_failures, s.consecutive_successes, s.next_check_at,
		       COALESCE((SELECT ROUND(COUNT(*) FILTER (WHERE status = 'UP')::numeric / GREATEST(1, COUNT(*)) * 100, 2) FROM check_history WHERE monitor_id = m.id AND checked_at >= NOW() - INTERVAL '24 hours'), 100.0) as uptime_24h,
		       COALESCE((SELECT ROUND(COUNT(*) FILTER (WHERE status = 'UP')::numeric / GREATEST(1, COUNT(*)) * 100, 2) FROM check_history WHERE monitor_id = m.id AND checked_at >= NOW() - INTERVAL '7 days'), 100.0) as uptime_7d
		FROM monitors m
		JOIN monitor_state s ON s.monitor_id = m.id
		WHERE m.id = $1::uuid AND m.user_id = $2::uuid;
	`

	var id, name, targetURL, method, status string
	var expectedStatus, intervalSeconds, timeoutSeconds int
	var enabled bool
	var createdAt, updatedAt time.Time
	var lastCheckedAt, nextCheckAt *time.Time
	var lastResponseTimeMs, lastStatusCode *int
	var lastError *string
	var consecFailures, consecSuccesses int
	var uptime24h, uptime7d float64

	err := utils.DB.QueryRow(ctx, query, monitorID, userID).Scan(
		&id, &name, &targetURL, &method, &expectedStatus, &intervalSeconds, &timeoutSeconds, &enabled, &createdAt, &updatedAt,
		&status, &lastCheckedAt, &lastResponseTimeMs, &lastStatusCode, &lastError,
		&consecFailures, &consecSuccesses, &nextCheckAt,
		&uptime24h, &uptime7d,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
			return
		}
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Database error", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":               id,
		"name":             name,
		"url":              targetURL,
		"method":           method,
		"expected_status":  expectedStatus,
		"interval_seconds": intervalSeconds,
		"timeout_seconds":  timeoutSeconds,
		"enabled":          enabled,
		"created_at":       createdAt,
		"updated_at":       updatedAt,
		"uptime_24h":       uptime24h,
		"uptime_7d":        uptime7d,
		"state": gin.H{
			"status":                 status,
			"last_checked_at":        lastCheckedAt,
			"last_response_time_ms":  lastResponseTimeMs,
			"last_status_code":       lastStatusCode,
			"last_error":             lastError,
			"consecutive_failures":   consecFailures,
			"consecutive_successes":  consecSuccesses,
			"next_check_at":          nextCheckAt,
		},
	})
}

func UpdateMonitor(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")
	var req UpdateMonitorReq
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid JSON payload", "")
		return
	}

	ctx := c.Request.Context()

	// Check monitor existence and ownership
	var exists bool
	err := utils.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitors WHERE id = $1::uuid AND user_id = $2::uuid)", monitorID, userID).Scan(&exists)
	if err != nil || !exists {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	sets := []string{}
	args := []interface{}{monitorID, userID}
	argIdx := 3

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if len(name) < 1 || len(name) > 255 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Name must be between 1 and 255 characters", "name")
			return
		}
		sets = append(sets, fmt.Sprintf("name = $%d", argIdx))
		args = append(args, name)
		argIdx++
	}

	if req.URL != nil {
		rawURL := strings.TrimSpace(*req.URL)
		if _, err := utils.ValidateURL(rawURL, allowPrivateSSRF()); err != nil {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), "url")
			return
		}
		sets = append(sets, fmt.Sprintf("url = $%d", argIdx))
		args = append(args, rawURL)
		argIdx++
	}

	if req.IntervalSeconds != nil {
		interval := *req.IntervalSeconds
		if interval < 60 || interval > 3600 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Interval must be between 60 and 3600 seconds", "interval_seconds")
			return
		}
		sets = append(sets, fmt.Sprintf("interval_seconds = $%d", argIdx))
		args = append(args, interval)
		argIdx++
	}

	if req.ExpectedStatus != nil {
		status := *req.ExpectedStatus
		if status < 100 || status > 599 {
			RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Expected status must be between 100 and 599", "expected_status")
			return
		}
		sets = append(sets, fmt.Sprintf("expected_status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	if len(sets) == 0 {
		RespondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "No fields provided to update", "")
		return
	}

	sets = append(sets, "updated_at = NOW()")
	query := fmt.Sprintf(`
		UPDATE monitors
		SET %s
		WHERE id = $1::uuid AND user_id = $2::uuid
	`, strings.Join(sets, ", "))

	_, err = utils.DB.Exec(ctx, query, args...)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update monitor", "")
		return
	}

	// Fetch updated monitor
	GetMonitor(c)
}

func DeleteMonitor(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")
	ctx := c.Request.Context()

	cmd, err := utils.DB.Exec(ctx, "DELETE FROM monitors WHERE id = $1::uuid AND user_id = $2::uuid", monitorID, userID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete monitor", "")
		return
	}

	if cmd.RowsAffected() == 0 {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	c.Status(http.StatusNoContent)
}

func PauseMonitor(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")
	ctx := c.Request.Context()

	tx, err := utils.DB.Begin(ctx)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction error", "")
		return
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, "UPDATE monitors SET enabled = false, updated_at = NOW() WHERE id = $1::uuid AND user_id = $2::uuid", monitorID, userID)
	if err != nil || cmd.RowsAffected() == 0 {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	_, err = tx.Exec(ctx, "UPDATE monitor_state SET status = 'PAUSED', updated_at = NOW() WHERE monitor_id = $1::uuid", monitorID)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "State update error", "")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Commit error", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "PAUSED", "enabled": false})
}

func ResumeMonitor(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")
	ctx := c.Request.Context()

	tx, err := utils.DB.Begin(ctx)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Transaction error", "")
		return
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, "UPDATE monitors SET enabled = true, updated_at = NOW() WHERE id = $1::uuid AND user_id = $2::uuid", monitorID, userID)
	if err != nil || cmd.RowsAffected() == 0 {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	now := time.Now().UTC()
	_, err = tx.Exec(ctx, "UPDATE monitor_state SET status = 'UNKNOWN', next_check_at = $2, updated_at = NOW() WHERE monitor_id = $1::uuid", monitorID, now)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "State update error", "")
		return
	}

	if err := tx.Commit(ctx); err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Commit error", "")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "UNKNOWN", "enabled": true, "next_check_at": now})
}

func GetMonitorChecks(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")
	rangeParam := c.DefaultQuery("range", "24h")

	var bucketInterval string
	var lookbackInterval string
	switch rangeParam {
	case "7d":
		bucketInterval = "1 hour"
		lookbackInterval = "7 days"
	default:
		rangeParam = "24h"
		bucketInterval = "5 minutes"
		lookbackInterval = "24 hours"
	}

	ctx := c.Request.Context()

	// Verify ownership
	var exists bool
	err := utils.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitors WHERE id = $1::uuid AND user_id = $2::uuid)", monitorID, userID).Scan(&exists)
	if err != nil || !exists {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	query := `
		SELECT date_bin($1::interval, checked_at, TIMESTAMPTZ '2000-01-01 00:00:00Z') AS ts,
		       ROUND(AVG(response_time_ms))::int AS avg_response_ms,
		       COUNT(*)::int AS checks,
		       COUNT(*) FILTER (WHERE status = 'DOWN')::int AS failures
		FROM check_history
		WHERE monitor_id = $2::uuid
		  AND checked_at >= NOW() - $3::interval
		GROUP BY ts
		ORDER BY ts ASC;
	`

	rows, err := utils.DB.Query(ctx, query, bucketInterval, monitorID, lookbackInterval)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query check series", "")
		return
	}
	defer rows.Close()

	type Bucket struct {
		TS            time.Time `json:"ts"`
		AvgResponseMs int       `json:"avg_response_ms"`
		Checks        int       `json:"checks"`
		Failures      int       `json:"failures"`
	}

	buckets := []Bucket{}
	for rows.Next() {
		var b Bucket
		if err := rows.Scan(&b.TS, &b.AvgResponseMs, &b.Checks, &b.Failures); err != nil {
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error scanning buckets", "")
			return
		}
		buckets = append(buckets, b)
	}

	c.JSON(http.StatusOK, gin.H{
		"range":   rangeParam,
		"buckets": buckets,
	})
}

func GetRecentChecks(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()

	// Verify ownership
	var exists bool
	err = utils.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitors WHERE id = $1::uuid AND user_id = $2::uuid)", monitorID, userID).Scan(&exists)
	if err != nil || !exists {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	query := `
		SELECT id, status::text, response_time_ms, status_code, error_message, checked_at
		FROM check_history
		WHERE monitor_id = $1::uuid
		ORDER BY checked_at DESC
		LIMIT $2;
	`

	rows, err := utils.DB.Query(ctx, query, monitorID, limit)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Query error", "")
		return
	}
	defer rows.Close()

	type RawCheck struct {
		ID             int64     `json:"id"`
		Status         string    `json:"status"`
		ResponseTimeMs *int      `json:"response_time_ms"`
		StatusCode     *int      `json:"status_code"`
		ErrorMessage   *string   `json:"error_message"`
		CheckedAt      time.Time `json:"checked_at"`
	}

	checks := []RawCheck{}
	for rows.Next() {
		var rc RawCheck
		if err := rows.Scan(&rc.ID, &rc.Status, &rc.ResponseTimeMs, &rc.StatusCode, &rc.ErrorMessage, &rc.CheckedAt); err != nil {
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Scan error", "")
			return
		}
		checks = append(checks, rc)
	}

	c.JSON(http.StatusOK, gin.H{"checks": checks})
}
