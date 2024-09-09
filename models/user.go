// models/user.go
package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	APIKey    string    `json:"api_key"`
}

func CreateUser(db *sql.DB, user *User) error {
	query := `
        INSERT INTO users (username, email, first_name, last_name, api_key)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, created_at, updated_at`

	return db.QueryRow(query, user.Username, user.Email, user.FirstName, user.LastName, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func GetUser(db *sql.DB, id int) (*User, error) {
	user := &User{}
	query := `SELECT * FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
		&user.CreatedAt, &user.UpdatedAt, &user.APIKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUsers(db *sql.DB) ([]*User, error) {
	query := `SELECT * FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.FirstName, &user.LastName,
			&user.CreatedAt, &user.UpdatedAt, &user.APIKey)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, nil
}

func UpdateUser(db *sql.DB, user *User) error {
	query := `
        UPDATE users
        SET username = $2, email = $3, first_name = $4, last_name = $5, api_key = $6, updated_at = CURRENT_TIMESTAMP
        WHERE id = $1
        RETURNING updated_at`

	return db.QueryRow(query, user.ID, user.Username, user.Email, user.FirstName, user.LastName, user.APIKey).
		Scan(&user.UpdatedAt)
}

func DeleteUser(db *sql.DB, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := db.Exec(query, id)
	return err
}
