// controllers/event_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nzenitram/relay-esp/middleware"
	"github.com/nzenitram/relay-esp/models"
)

type EventController struct {
	DB *sql.DB
}

func NewEventController(db *sql.DB) *EventController {
	return &EventController{DB: db}
}

func (ec *EventController) GetEvents(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50 // Default limit
	}

	events, err := models.GetEventsByUserID(ec.DB, authUser.ID, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (ec *EventController) GetEventsByType(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	eventType := vars["type"]

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit == 0 {
		limit = 50 // Default limit
	}

	events, err := models.GetEventsByTypeAndUserID(ec.DB, authUser.ID, eventType, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}

func (ec *EventController) GetAvailableEventTypes(w http.ResponseWriter, r *http.Request) {
	eventTypes, err := models.GetAvailableEventTypes(ec.DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string][]string{"event_types": eventTypes})
}

func (ec *EventController) GetProviderEventStatsByType(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	providerName := vars["provider"]
	eventType := vars["event"]

	// Validate provider and event type
	if !isValidProvider(providerName) {
		http.Error(w, "Invalid provider", http.StatusBadRequest)
		return
	}
	if !isValidEventType(eventType) {
		http.Error(w, "Invalid event type", http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startTime, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Set end time to the end of the day
	endTime = endTime.Add(24*time.Hour - time.Second)

	stats, err := models.GetProviderEventStatsByType(ec.DB, authUser.ID, providerName, eventType, startTime, endTime)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// Helper functions to validate input
func isValidProvider(provider string) bool {
	validProviders := []string{"sendgrid", "postmark", "sparkpost", "socketlabs"} // Add all valid providers
	for _, p := range validProviders {
		if p == provider {
			return true
		}
	}
	return false
}

func isValidEventType(eventType string) bool {
	validTypes := []string{"processed", "delivered", "bounce", "deferred", "unique_open", "open", "dropped"}
	for _, t := range validTypes {
		if t == eventType {
			return true
		}
	}
	return false
}

func (ec *ESPController) GetProviderEventStats(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	providerName := vars["provider"]
	if providerName == "" {
		http.Error(w, "Provider name is required", http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	startTime, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	// Set end time to the end of the day
	endTime = endTime.Add(24*time.Hour - time.Second)

	stats, err := models.GetProviderEventStats(ec.DB, authUser.ID, providerName, startTime, endTime)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
