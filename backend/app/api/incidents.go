package api

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/enftaurus/stayup/utils"
	"github.com/gin-gonic/gin"
)

type IncidentItem struct {
	ID              string     `json:"id"`
	MonitorID       string     `json:"monitor_id"`
	MonitorName     string     `json:"monitor_name,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	ResolvedAt      *time.Time `json:"resolved_at"`
	DurationSeconds *int       `json:"duration_seconds"`
	Status          string     `json:"status"`
}

func GetMonitorIncidents(c *gin.Context) {
	userID := GetUserID(c)
	monitorID := c.Param("id")

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	cursor := c.Query("cursor")
	var cursorTime time.Time
	if cursor != "" {
		if raw, err := base64.StdEncoding.DecodeString(cursor); err == nil {
			cursorTime, _ = time.Parse(time.RFC3339Nano, string(raw))
		}
	}

	ctx := c.Request.Context()

	// Verify ownership
	var exists bool
	err = utils.DB.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM monitors WHERE id = $1::uuid AND user_id = $2::uuid)", monitorID, userID).Scan(&exists)
	if err != nil || !exists {
		RespondError(c, http.StatusNotFound, "NOT_FOUND", "Monitor not found", "")
		return
	}

	var query string
	var args []interface{}
	if cursorTime.IsZero() {
		query = `
			SELECT id::text, monitor_id::text, started_at, resolved_at, duration_seconds, status::text
			FROM incidents
			WHERE monitor_id = $1::uuid
			ORDER BY started_at DESC
			LIMIT $2;
		`
		args = []interface{}{monitorID, limit + 1}
	} else {
		query = `
			SELECT id::text, monitor_id::text, started_at, resolved_at, duration_seconds, status::text
			FROM incidents
			WHERE monitor_id = $1::uuid AND started_at < $2
			ORDER BY started_at DESC
			LIMIT $3;
		`
		args = []interface{}{monitorID, cursorTime, limit + 1}
	}

	rows, err := utils.DB.Query(ctx, query, args...)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query incidents", "")
		return
	}
	defer rows.Close()

	incidents := []IncidentItem{}
	for rows.Next() {
		var inc IncidentItem
		if err := rows.Scan(&inc.ID, &inc.MonitorID, &inc.StartedAt, &inc.ResolvedAt, &inc.DurationSeconds, &inc.Status); err != nil {
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error parsing row", "")
			return
		}
		incidents = append(incidents, inc)
	}

	var nextCursor *string
	if len(incidents) > limit {
		nextCursorVal := base64.StdEncoding.EncodeToString([]byte(incidents[limit-1].StartedAt.Format(time.RFC3339Nano)))
		nextCursor = &nextCursorVal
		incidents = incidents[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents":   incidents,
		"next_cursor": nextCursor,
	})
}

func ListIncidents(c *gin.Context) {
	userID := GetUserID(c)
	statusFilter := strings.ToLower(strings.TrimSpace(c.Query("status")))

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	cursor := c.Query("cursor")
	var cursorTime time.Time
	if cursor != "" {
		if raw, err := base64.StdEncoding.DecodeString(cursor); err == nil {
			cursorTime, _ = time.Parse(time.RFC3339Nano, string(raw))
		}
	}

	ctx := c.Request.Context()

	whereClauses := []string{"m.user_id = $1::uuid"}
	args := []interface{}{userID}
	argIdx := 2

	if statusFilter == "open" {
		whereClauses = append(whereClauses, "i.resolved_at IS NULL")
	} else if statusFilter == "resolved" {
		whereClauses = append(whereClauses, "i.resolved_at IS NOT NULL")
	}

	if !cursorTime.IsZero() {
		whereClauses = append(whereClauses, fmt.Sprintf("i.started_at < $%d", argIdx))
		args = append(args, cursorTime)
		argIdx++
	}

	args = append(args, limit+1)
	query := fmt.Sprintf(`
		SELECT i.id::text, i.monitor_id::text, m.name, i.started_at, i.resolved_at, i.duration_seconds, i.status::text
		FROM incidents i
		JOIN monitors m ON m.id = i.monitor_id
		WHERE %s
		ORDER BY i.started_at DESC
		LIMIT $%d;
	`, strings.Join(whereClauses, " AND "), argIdx)

	rows, err := utils.DB.Query(ctx, query, args...)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to query incidents", "")
		return
	}
	defer rows.Close()

	incidents := []IncidentItem{}
	for rows.Next() {
		var inc IncidentItem
		if err := rows.Scan(&inc.ID, &inc.MonitorID, &inc.MonitorName, &inc.StartedAt, &inc.ResolvedAt, &inc.DurationSeconds, &inc.Status); err != nil {
			RespondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Error parsing row", "")
			return
		}
		incidents = append(incidents, inc)
	}

	var nextCursor *string
	if len(incidents) > limit {
		nextCursorVal := base64.StdEncoding.EncodeToString([]byte(incidents[limit-1].StartedAt.Format(time.RFC3339Nano)))
		nextCursor = &nextCursorVal
		incidents = incidents[:limit]
	}

	c.JSON(http.StatusOK, gin.H{
		"incidents":   incidents,
		"next_cursor": nextCursor,
	})
}
