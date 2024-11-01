// models/user.go
package models

import (
	"database/sql"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	APIKey    string    `json:"api_key"`
}

func CreateUser(db *sql.DB, user *User) error {
	query := `
        INSERT INTO users (username, email, first_name, last_name, api_key, password)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at`

	return db.QueryRow(query, user.Username, user.Email, user.FirstName, user.LastName, user.APIKey).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

func GetUsers(db *sql.DB) ([]*User, error) {
	query := `SELECT SELECT id, username, email, api_key FROM users`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}
	for rows.Next() {
		user := &User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.APIKey)
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

func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}

// Get User By

func GetUserByEmail(db *sql.DB, email string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, api_key, password
              FROM users WHERE email = $1`
	err := db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &user.APIKey, &user.Password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(db *sql.DB, id int) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, api_key FROM users WHERE id = $1`
	err := db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.APIKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByAPIKey(db *sql.DB, apiKey string) (*User, error) {
	user := &User{}
	query := `SELECT id, username, email, api_key FROM users WHERE api_key = $1`
	err := db.QueryRow(query, apiKey).Scan(
		&user.ID, &user.Username, &user.Email, &user.APIKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(db *sql.DB, username string) (*User, error) {
	user := &User{}
	err := db.QueryRow("SELECT id, username, email, api_key FROM users WHERE username = $1", username).
		Scan(&user.ID, &user.Username, &user.Email, &user.APIKey)
	if err != nil {
		return nil, err
	}
	return user, nil
}
