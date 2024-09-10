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
