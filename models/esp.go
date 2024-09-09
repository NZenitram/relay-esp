// models/esp.go
package models

import (
	"database/sql"
	"strconv"

	"github.com/lib/pq"
)

type ESP struct {
	ESPID          int      `json:"esp_id"`
	UserID         int      `json:"user_id"`
	ProviderName   string   `json:"provider_name"`
	SendingDomains []string `json:"sending_domains"`
	CreatedAt      string   `json:"created_at"`
	UpdatedAt      string   `json:"updated_at"`
	Weight         int      `json:"weight"`
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
