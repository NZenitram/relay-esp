// models/esp.go
package models

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/lib/pq"
)

type ESP struct {
	ESPID                    int       `json:"esp_id"`
	UserID                   int       `json:"user_id"`
	ProviderName             string    `json:"provider_name"`
	SendingDomains           []string  `json:"sending_domains"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
	SendgridVerificationKey  string    `json:"sendgrid_verification_key,omitempty"`
	SparkpostWebhookUser     string    `json:"sparkpost_webhook_user,omitempty"`
	SparkpostWebhookPassword string    `json:"sparkpost_webhook_password,omitempty"`
	SocketlabsSecretKey      string    `json:"socketlabs_secret_key,omitempty"`
	PostmarkWebhookUser      string    `json:"postmark_webhook_user,omitempty"`
	PostmarkWebhookPassword  string    `json:"postmark_webhook_password,omitempty"`
	SocketlabsServerID       string    `json:"socketlabs_server_id,omitempty"`
	Weight                   int       `json:"weight"`
}

func GetESPsByUserID(db *sql.DB, userID int) ([]ESP, error) {
	query := `
        SELECT esp_id, user_id, provider_name, sending_domains, 
               created_at, updated_at, weight
        FROM email_service_providers
        WHERE user_id = $1
        ORDER BY provider_name
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var esps []ESP
	for rows.Next() {
		var esp ESP

		err := rows.Scan(
			&esp.ESPID,
			&esp.UserID,
			&esp.ProviderName,
			pq.Array(&esp.SendingDomains),
			&esp.CreatedAt,
			&esp.UpdatedAt,
			&esp.Weight,
		)
		if err != nil {
			return nil, err
		}

		esps = append(esps, esp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return esps, nil
}

func GetESPsByUserIDWithFilters(db *sql.DB, userID int, espID int, providerName string, sendingDomain string) ([]ESP, error) {
	query := `
        SELECT esp_id, user_id, provider_name, sending_domains, 
               created_at, updated_at, weight
        FROM email_service_providers
        WHERE 1=1
    `
	args := []interface{}{}
	argCount := 0

	// Use either userID or providerName, prioritizing userID if both are provided
	if userID != 0 {
		argCount++
		query += ` AND user_id = $` + strconv.Itoa(argCount)
		args = append(args, userID)
	} else if providerName != "" {
		argCount++
		query += ` AND LOWER(provider_name) LIKE LOWER($` + strconv.Itoa(argCount) + `)`
		args = append(args, "%"+providerName+"%")
	}

	if espID != 0 {
		argCount++
		query += ` AND esp_id = $` + strconv.Itoa(argCount)
		args = append(args, espID)
	}

	if sendingDomain != "" {
		argCount++
		query += ` AND $` + strconv.Itoa(argCount) + ` = ANY(sending_domains)`
		args = append(args, sendingDomain)
	}

	query += ` ORDER BY provider_name`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var esps []ESP
	for rows.Next() {
		var esp ESP
		err := rows.Scan(
			&esp.ESPID,
			&esp.UserID,
			&esp.ProviderName,
			pq.Array(&esp.SendingDomains),
			&esp.CreatedAt,
			&esp.UpdatedAt,
			&esp.Weight,
		)
		if err != nil {
			return nil, err
		}
		esps = append(esps, esp)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return esps, nil
}

func CreateESP(db *sql.DB, esp *ESP) error {
	query := `
        INSERT INTO email_service_providers (
            user_id, provider_name, sending_domains, sendgrid_verification_key,
            sparkpost_webhook_user, sparkpost_webhook_password, socketlabs_secret_key,
            postmark_webhook_user, postmark_webhook_password, socketlabs_server_id, weight
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING esp_id, created_at, updated_at`

	err := db.QueryRow(
		query,
		esp.UserID,
		esp.ProviderName,
		pq.Array(esp.SendingDomains),
		esp.SendgridVerificationKey,
		esp.SparkpostWebhookUser,
		esp.SparkpostWebhookPassword,
		esp.SocketlabsSecretKey,
		esp.PostmarkWebhookUser,
		esp.PostmarkWebhookPassword,
		esp.SocketlabsServerID,
		esp.Weight,
	).Scan(&esp.ESPID, &esp.CreatedAt, &esp.UpdatedAt)

	return err
}

func UpdateESP(db *sql.DB, esp *ESP) error {
	query := `
        UPDATE email_service_providers
        SET 
            user_id = $1,
            provider_name = $2,
            sending_domains = $3,
            sendgrid_verification_key = $4,
            sparkpost_webhook_user = $5,
            sparkpost_webhook_password = $6,
            socketlabs_secret_key = $7,
            postmark_webhook_user = $8,
            postmark_webhook_password = $9,
            socketlabs_server_id = $10,
            weight = $11,
            updated_at = CURRENT_TIMESTAMP
        WHERE esp_id = $12
        RETURNING updated_at`

	err := db.QueryRow(
		query,
		esp.UserID,
		esp.ProviderName,
		pq.Array(esp.SendingDomains),
		esp.SendgridVerificationKey,
		esp.SparkpostWebhookUser,
		esp.SparkpostWebhookPassword,
		esp.SocketlabsSecretKey,
		esp.PostmarkWebhookUser,
		esp.PostmarkWebhookPassword,
		esp.SocketlabsServerID,
		esp.Weight,
		esp.ESPID,
	).Scan(&esp.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Printf("no ESP found with ID %v", esp.ESPID)
		}
		fmt.Printf("error updating ESP: %v", err)
	}

	return nil
}

func DeleteESP(db *sql.DB, espID, userID int) error {
	query := `
        DELETE FROM email_service_providers
        WHERE esp_id = $1 AND user_id = $2
    `

	result, err := db.Exec(query, espID, userID)
	if err != nil {
		return fmt.Errorf("error deleting ESP: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no ESP found with ID %d for user %d", espID, userID)
	}

	return nil
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
	query := `SELECT
    date_trunc('hour', to_timestamp(CASE
        WHEN e.processed_time IS NOT NULL THEN e.processed_time
        WHEN e.delivered_time IS NOT NULL THEN e.delivered_time
        WHEN e.bounce_time IS NOT NULL THEN e.bounce_time
        WHEN e.last_deferral_time IS NOT NULL THEN e.last_deferral_time
        WHEN e.unique_open_time IS NOT NULL THEN e.unique_open_time
        WHEN e.last_open_time IS NOT NULL THEN e.last_open_time
        WHEN e.dropped_time IS NOT NULL THEN e.dropped_time
        ELSE NULL
    END / 1000)) AS time_bucket,
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
    AND to_timestamp(CASE
        WHEN e.processed_time IS NOT NULL THEN e.processed_time
        WHEN e.delivered_time IS NOT NULL THEN e.delivered_time
        WHEN e.bounce_time IS NOT NULL THEN e.bounce_time
        WHEN e.last_deferral_time IS NOT NULL THEN e.last_deferral_time
        WHEN e.unique_open_time IS NOT NULL THEN e.unique_open_time
        WHEN e.last_open_time IS NOT NULL THEN e.last_open_time
        WHEN e.dropped_time IS NOT NULL THEN e.dropped_time
        ELSE NULL
    END / 1000) BETWEEN $2 AND $3
GROUP BY
    time_bucket, e.provider
ORDER BY
    time_bucket, e.provider;
	`
	rows, err := db.Query(query, userID, startTime, endTime)
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
