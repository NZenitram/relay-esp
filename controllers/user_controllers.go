// controllers/user_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/nzenitram/relay-esp/middleware"
	"github.com/nzenitram/relay-esp/models"
	"github.com/nzenitram/relay-esp/utils"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/exp/rand"
)

type UserController struct {
	DB *sql.DB
}

var (
	sendgridAPIKey string
)

func NewUserController(db *sql.DB) *UserController {
	return &UserController{DB: db}
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Get SendGrid API key from environment variable
	sendgridAPIKey = os.Getenv("SENDGRID_API_KEY")
	if sendgridAPIKey == "" {
		log.Fatal("SENDGRID_API_KEY is not set in the environment")
	}
}

func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := models.GetUserByEmail(uc.DB, loginData.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if !user.CheckPassword(loginData.Password) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Create a session
	err = createSession(uc.DB, user.ID, token)
	if err != nil {
		http.Error(w, "Error creating session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token":    token,
		"email":    user.Email,
		"username": user.Username,
	})
}

func (uc *UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the authenticated user is requesting their own data
	if authUser.ID != id {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]string{"error": "Forbidden: You can only access your own data"})
		return
	}

	user, err := models.GetUserByID(uc.DB, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (uc *UserController) UpdateUser(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the authenticated user is updating their own data
	if authUser.ID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var user models.User
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	user.ID = id

	if err := models.UpdateUser(uc.DB, &user); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (uc *UserController) DeleteUser(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the authenticated user is deleting their own account
	if authUser.ID != id {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := models.DeleteUser(uc.DB, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (uc *UserController) GetUsers(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Instead of getting all users, just return the authenticated user
	user, err := models.GetUserByID(uc.DB, authUser.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode([]*models.User{user}) // Wrap in array to maintain consistent response format
}

func (uc *UserController) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var requestData struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	resetToken := generateResetToken()
	expiryTime := time.Now().UTC().Add(1 * time.Hour)

	err := storeResetToken(uc.DB, requestData.Email, resetToken, expiryTime)
	if err != nil {
		log.Printf("Failed to store reset token: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = sendResetEmail(requestData.Email, resetToken)
	if err != nil {
		log.Printf("Failed to send reset email: %v", err)
		http.Error(w, "Failed to send reset email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset requested. Check your email for instructions."})
}

func (uc *UserController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var resetData struct {
		ResetToken  string `json:"reset_token"`
		NewPassword string `json:"new_password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&resetData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	valid, email := verifyResetToken(uc.DB, resetData.ResetToken)
	if !valid {
		http.Error(w, "Invalid or expired reset token", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(resetData.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing new password", http.StatusInternalServerError)
		return
	}

	err = updatePassword(uc.DB, email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Error updating password", http.StatusInternalServerError)
		return
	}

	clearResetToken(uc.DB, email)
	logPasswordChange(email)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Password reset successfully"})
}

func generateResetToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

func storeResetToken(db *sql.DB, email, token string, expiry time.Time) error {
	_, err := db.Exec("UPDATE users SET reset_token = $1, reset_token_expiry = $2 WHERE email = $3",
		token, expiry, email)
	return err
}

func sendResetEmail(email, token string) error {
	from := mail.NewEmail("ESP Relay", "noreply@esprelay.com")
	subject := "Password Reset Request"
	to := mail.NewEmail("", email)

	resetLink := fmt.Sprintf("http://localhost:8081/reset-password?token=%s", token)
	plainTextContent := fmt.Sprintf("Click the following link to reset your password: %s", resetLink)
	htmlContent := fmt.Sprintf("<p>Click the following link to reset your password:</p><p><a href=\"%s\">Reset Password</a></p>", resetLink)

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(sendgridAPIKey)
	response, err := client.Send(message)
	if err != nil {
		log.Printf("Error sending email: %v", err)
		return err
	}
	log.Printf("Email sent. Status: %d, Body: %s", response.StatusCode, response.Body)
	return nil
}

func verifyResetToken(db *sql.DB, token string) (bool, string) {
	var email string
	var expiry time.Time
	err := db.QueryRow("SELECT email, reset_token_expiry FROM users WHERE reset_token = $1", token).
		Scan(&email, &expiry)
	if err != nil {
		log.Printf("Error querying reset token: %v", err)
		return false, ""
	}

	now := time.Now().UTC()
	log.Printf("Current time: %v", now)
	log.Printf("Token expiry: %v", expiry)
	log.Printf("Difference: %v", expiry.Sub(now))

	if now.After(expiry) {
		log.Printf("Token expired. Current time: %v, Expiry time: %v", now, expiry)
		return false, ""
	}
	return true, email
}

func updatePassword(db *sql.DB, email, hashedPassword string) error {
	_, err := db.Exec("UPDATE users SET password = $1 WHERE email = $2", hashedPassword, email)
	return err
}

func clearResetToken(db *sql.DB, email string) {
	db.Exec("UPDATE users SET reset_token = NULL, reset_token_expiry = NULL WHERE email = $1", email)
}

// Function to log password change events
func logPasswordChange(username string) {
	log.Printf("Password changed successfully for user: %s", username)
	// In a production environment, you might want to use a more robust logging system
}

func createSession(db *sql.DB, userID int, token string) error {
	query := `INSERT INTO sessions (user_id, token, created_at, expires_at) 
              VALUES ($1, $2, $3, $4)`
	_, err := db.Exec(query, userID, token, time.Now(), time.Now().Add(24*time.Hour))
	return err
}
