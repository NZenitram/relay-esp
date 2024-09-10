package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	ID               int             `json:"id"`
	MessageID        string          `json:"message_id"`
	Processed        bool            `json:"processed"`
	ProcessedTime    sql.NullInt64   `json:"processed_time"`
	Delivered        bool            `json:"delivered"`
	DeliveredTime    sql.NullInt64   `json:"delivered_time"`
	Bounce           bool            `json:"bounce"`
	BounceType       sql.NullString  `json:"bounce_type"`
	BounceTime       sql.NullInt64   `json:"bounce_time"`
	Deferred         bool            `json:"deferred"`
	DeferredCount    int             `json:"deferred_count"`
	LastDeferralTime sql.NullInt64   `json:"last_deferral_time"`
	UniqueOpen       bool            `json:"unique_open"`
	UniqueOpenTime   sql.NullInt64   `json:"unique_open_time"`
	Open             bool            `json:"open"`
	OpenCount        int             `json:"open_count"`
	LastOpenTime     sql.NullInt64   `json:"last_open_time"`
	Dropped          bool            `json:"dropped"`
	DroppedTime      sql.NullInt64   `json:"dropped_time"`
	DroppedReason    sql.NullString  `json:"dropped_reason"`
	Provider         string          `json:"provider"`
	Metadata         json.RawMessage `json:"metadata"`
}

func (e Event) MarshalJSON() ([]byte, error) {
	type Alias Event
	return json.Marshal(&struct {
		ProcessedTime    *int64  `json:"processed_time"`
		DeliveredTime    *int64  `json:"delivered_time"`
		BounceType       *string `json:"bounce_type"`
		BounceTime       *int64  `json:"bounce_time"`
		LastDeferralTime *int64  `json:"last_deferral_time"`
		UniqueOpenTime   *int64  `json:"unique_open_time"`
		LastOpenTime     *int64  `json:"last_open_time"`
		DroppedTime      *int64  `json:"dropped_time"`
		DroppedReason    *string `json:"dropped_reason"`
		Alias
	}{
		ProcessedTime:    nullInt64ToPtr(e.ProcessedTime),
		DeliveredTime:    nullInt64ToPtr(e.DeliveredTime),
		BounceType:       nullStringToPtr(e.BounceType),
		BounceTime:       nullInt64ToPtr(e.BounceTime),
		LastDeferralTime: nullInt64ToPtr(e.LastDeferralTime),
		UniqueOpenTime:   nullInt64ToPtr(e.UniqueOpenTime),
		LastOpenTime:     nullInt64ToPtr(e.LastOpenTime),
		DroppedTime:      nullInt64ToPtr(e.DroppedTime),
		DroppedReason:    nullStringToPtr(e.DroppedReason),
		Alias:            (Alias)(e),
	})
}

func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

func nullStringToPtr(s sql.NullString) *string {
	if s.Valid {
		return &s.String
	}
	return nil
}

func GetEventsByUserID(db *sql.DB, userID int, limit, offset int) ([]Event, error) {
	query := `
        SELECT e.* 
        FROM events e
        JOIN message_user_associations mua ON e.message_id = mua.message_id
        WHERE mua.user_id = $1
        ORDER BY e.id DESC
        LIMIT $2 OFFSET $3
    `
	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID, &e.MessageID, &e.Processed, &e.ProcessedTime, &e.Delivered, &e.DeliveredTime,
			&e.Bounce, &e.BounceType, &e.BounceTime, &e.Deferred, &e.DeferredCount, &e.LastDeferralTime,
			&e.UniqueOpen, &e.UniqueOpenTime, &e.Open, &e.OpenCount, &e.LastOpenTime,
			&e.Dropped, &e.DroppedTime, &e.DroppedReason, &e.Provider, &e.Metadata,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func GetEventsByTypeAndUserID(db *sql.DB, userID int, eventType string, limit, offset int) ([]Event, error) {
	query := `
        SELECT e.* 
        FROM events e
        JOIN message_user_associations mua ON e.message_id = mua.message_id
        WHERE mua.user_id = $1 AND %s
        ORDER BY e.id DESC
        LIMIT $2 OFFSET $3
    `

	var condition string
	switch eventType {
	case "delivered":
		condition = "e.delivered = true"
	case "bounce":
		condition = "e.bounce = true"
	case "deferred":
		condition = "e.deferred = true"
	case "open":
		condition = "e.open = true"
	case "dropped":
		condition = "e.dropped = true"
	case "processed":
		condition = "e.processed = true"
	case "unique_open":
		condition = "e.unique_open = true"
	default:
		return nil, fmt.Errorf("invalid event type")
	}

	query = fmt.Sprintf(query, condition)
	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		err := rows.Scan(
			&e.ID, &e.MessageID, &e.Processed, &e.ProcessedTime, &e.Delivered, &e.DeliveredTime,
			&e.Bounce, &e.BounceType, &e.BounceTime, &e.Deferred, &e.DeferredCount, &e.LastDeferralTime,
			&e.UniqueOpen, &e.UniqueOpenTime, &e.Open, &e.OpenCount, &e.LastOpenTime,
			&e.Dropped, &e.DroppedTime, &e.DroppedReason, &e.Provider, &e.Metadata,
		)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func GetAvailableEventTypes(db *sql.DB) ([]string, error) {
	query := `
        SELECT 'processed' AS event_type
        UNION ALL
        SELECT 'delivered'
        UNION ALL
        SELECT 'bounce'
        UNION ALL
        SELECT 'deferred'
        UNION ALL
        SELECT 'unique_open'
        UNION ALL
        SELECT 'open'
        UNION ALL
        SELECT 'dropped'
        ORDER BY event_type
    `

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var eventTypes []string
	for rows.Next() {
		var eventType string
		if err := rows.Scan(&eventType); err != nil {
			return nil, err
		}
		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes, nil
}

type EventStats struct {
	TimeBucket      time.Time
	Provider        string
	TotalEvents     int
	ProcessedCount  int
	DeliveredCount  int
	BounceCount     int
	DeferredCount   int
	UniqueOpenCount int
	OpenCount       int
	DroppedCount    int
}

func GetUserEventStats(db *sql.DB, userID int, startTime, endTime time.Time) ([]EventStats, error) {
	startUnix := startTime.Unix()
	endUnix := endTime.Unix()
	query := `
    WITH event_times AS (
        SELECT
            CASE
                WHEN COALESCE(e.processed_time, e.delivered_time, e.bounce_time,
                              e.last_deferral_time, e.unique_open_time, e.last_open_time,
                              e.dropped_time) BETWEEN $2 AND $3
                THEN date_trunc('hour', to_timestamp(
                    COALESCE(e.processed_time, e.delivered_time, e.bounce_time,
                             e.last_deferral_time, e.unique_open_time, e.last_open_time,
                             e.dropped_time, 0)
                ))
                ELSE NULL
            END AS time_bucket,
            e.provider,
            e.processed,
            e.delivered,
            e.bounce,
            e.deferred,
            e.unique_open,
            e.open,
            e.dropped
        FROM events e
        JOIN message_user_associations mua ON e.message_id = mua.message_id
        WHERE mua.user_id = $1
    )
    SELECT
        time_bucket,
        provider,
        COUNT(*) AS total_events,
        SUM(CASE WHEN processed THEN 1 ELSE 0 END) AS processed_count,
        SUM(CASE WHEN delivered THEN 1 ELSE 0 END) AS delivered_count,
        SUM(CASE WHEN bounce THEN 1 ELSE 0 END) AS bounce_count,
        SUM(CASE WHEN deferred THEN 1 ELSE 0 END) AS deferred_count,
        SUM(CASE WHEN unique_open THEN 1 ELSE 0 END) AS unique_open_count,
        SUM(CASE WHEN open THEN 1 ELSE 0 END) AS open_count,
        SUM(CASE WHEN dropped THEN 1 ELSE 0 END) AS dropped_count
    FROM event_times
    WHERE time_bucket IS NOT NULL
    GROUP BY time_bucket, provider
    ORDER BY time_bucket, provider;
	`
	rows, err := db.Query(query, userID, startUnix, endUnix)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []EventStats
	for rows.Next() {
		var s EventStats
		err := rows.Scan(
			&s.TimeBucket,
			&s.Provider,
			&s.TotalEvents,
			&s.ProcessedCount,
			&s.DeliveredCount,
			&s.BounceCount,
			&s.DeferredCount,
			&s.UniqueOpenCount,
			&s.OpenCount,
			&s.DroppedCount,
		)
		if err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func GetProviderEventStats(db *sql.DB, userID int, providerName string, startTime, endTime time.Time) ([]EventStats, error) {
	query := `
    SELECT 
        date_trunc('hour', to_timestamp(COALESCE(
            e.processed_time,
            e.delivered_time,
            e.bounce_time,
            e.last_deferral_time,
            e.unique_open_time,
            e.last_open_time,
            e.dropped_time,
            0
        ))) AS time_bucket,
        e.provider,
        COUNT(*) AS total_events,
        SUM(CASE WHEN e.processed THEN 1 ELSE 0 END) AS processed_count,
        SUM(CASE WHEN e.delivered THEN 1 ELSE 0 END) AS delivered_count,
        SUM(CASE WHEN e.bounce THEN 1 ELSE 0 END) AS bounce_count,
        SUM(CASE WHEN e.deferred THEN 1 ELSE 0 END) AS deferred_count,
        SUM(CASE WHEN e.unique_open THEN 1 ELSE 0 END) AS unique_open_count,
        SUM(CASE WHEN e.open THEN 1 ELSE 0 END) AS open_count,
        SUM(CASE WHEN e.dropped THEN 1 ELSE 0 END) AS dropped_count
    FROM 
        events e
    JOIN 
        message_user_associations mua ON e.message_id = mua.message_id
    JOIN 
        email_service_providers esp ON mua.esp_id = esp.esp_id
    WHERE 
        esp.user_id = $1
        AND e.provider = $2
        AND COALESCE(
            e.processed_time,
            e.delivered_time,
            e.bounce_time,
            e.last_deferral_time,
            e.unique_open_time,
            e.last_open_time,
            e.dropped_time,
            0
        ) BETWEEN $3 AND $4
    GROUP BY 
        time_bucket, e.provider
    ORDER BY 
        time_bucket, e.provider;
    `

	rows, err := db.Query(query, userID, providerName, startTime.Unix(), endTime.Unix())
	if err != nil {
		return nil, fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var stats []EventStats
	for rows.Next() {
		var s EventStats
		err := rows.Scan(
			&s.TimeBucket,
			&s.Provider,
			&s.TotalEvents,
			&s.ProcessedCount,
			&s.DeliveredCount,
			&s.BounceCount,
			&s.DeferredCount,
			&s.UniqueOpenCount,
			&s.OpenCount,
			&s.DroppedCount,
		)
		if err != nil {
			return nil, fmt.Errorf("row scan error: %v", err)
		}
		stats = append(stats, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return stats, nil
}
