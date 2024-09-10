// controllers/esp_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/nzenitram/relay-esp/middleware"
	"github.com/nzenitram/relay-esp/models"
)

type ESPController struct {
	DB *sql.DB
}

func NewESPController(db *sql.DB) *ESPController {
	return &ESPController{DB: db}
}

func (ec *ESPController) GetESPs(w http.ResponseWriter, r *http.Request) {
	var userID int
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if ok {
		userID = authUser.ID
	}

	espID, _ := strconv.Atoi(r.URL.Query().Get("esp_id"))
	providerName := r.URL.Query().Get("provider_name")
	sendingDomain := r.URL.Query().Get("sending_domain")

	// If no userID (not authenticated) and no providerName, return an error
	if userID == 0 && providerName == "" {
		http.Error(w, "Unauthorized: Must provide either authentication or a provider name", http.StatusUnauthorized)
		return
	}

	esps, err := models.GetESPsByUserIDWithFilters(ec.DB, userID, espID, providerName, sendingDomain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]models.ESP{"esps": esps})
}

func (ec *ESPController) CreateESP(w http.ResponseWriter, r *http.Request) {
	// Get the authenticated user from the context
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var esp models.ESP
	err := json.NewDecoder(r.Body).Decode(&esp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the UserID to the authenticated user's ID
	esp.UserID = authUser.ID

	// Create the ESP
	err = models.CreateESP(ec.DB, &esp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(esp)
}

func (ec *ESPController) UpdateESP(w http.ResponseWriter, r *http.Request) {
	// Get the esp_id from the URL parameters
	vars := mux.Vars(r)
	espID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ESP ID", http.StatusBadRequest)
		return
	}

	// Get the authenticated user from the context
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Decode the request body into an ESP struct
	var esp models.ESP
	err = json.NewDecoder(r.Body).Decode(&esp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Set the ESP ID and User ID
	esp.ESPID = espID
	esp.UserID = authUser.ID

	// Update the ESP
	err = models.UpdateESP(ec.DB, &esp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the updated ESP
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(esp)
}

func (ec *ESPController) DeleteESP(w http.ResponseWriter, r *http.Request) {
	// Get the esp_id from the URL parameters
	vars := mux.Vars(r)
	espID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ESP ID", http.StatusBadRequest)
		return
	}

	// Get the authenticated user from the context
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Delete the ESP
	err = models.DeleteESP(ec.DB, espID, authUser.ID)
	if err != nil {
		if strings.Contains(err.Error(), "no ESP found") {
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Return a success message
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "ESP deleted successfully"})
}

func (ec *ESPController) GetUserEventStats(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse start and end times from query parameters
	startTimeStr := r.URL.Query().Get("start_date")
	endTimeStr := r.URL.Query().Get("end_date")

	// Use a more user-friendly date format
	const dateFormat = "2006-01-02"

	startTime, err := time.Parse(dateFormat, startTimeStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(dateFormat, endTimeStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Set the end time to the end of the day
	endTime = endTime.Add(24*time.Hour - time.Second)

	stats, err := models.GetUserEventStats(ec.DB, authUser.ID, startTime, endTime)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
