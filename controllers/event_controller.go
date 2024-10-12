// controllers/event_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	// Get the authenticated user from the context
	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	eventType := vars["type"]

	// Parse query parameters
	startTime := r.URL.Query().Get("start")
	endTime := r.URL.Query().Get("end")
	bucketSize := r.URL.Query().Get("bucket")

	// Validate and set default values
	if startTime == "" || endTime == "" {
		http.Error(w, "Missing required parameters: start, end", http.StatusBadRequest)
		return
	}
	if bucketSize == "" {
		bucketSize = "1 hour" // Default bucket size
	}

	// Construct the SQL query based on the event type
	var query string
	switch eventType {
	case "processed":
		query = `SELECT time_bucket($1, time) AS time, provider, COUNT(DISTINCT message_id) AS count
                 FROM processed_events
                 WHERE time BETWEEN $2 AND $3 AND user_id = $4
                 GROUP BY 1, 2 ORDER BY 1, 2`
	case "delivered":
		query = `SELECT time_bucket($1, time) AS time, provider, COUNT(DISTINCT message_id) AS count
                 FROM delivered_events
                 WHERE time BETWEEN $2 AND $3 AND user_id = $4
                 GROUP BY 1, 2 ORDER BY 1, 2`
	case "bounce":
		query = `SELECT time_bucket($1, time) AS time, provider, COUNT(DISTINCT message_id) AS count
                 FROM bounce_events
                 WHERE time BETWEEN $2 AND $3 AND user_id = $4
                 GROUP BY 1, 2 ORDER BY 1, 2`
	case "deferred":
		query = `SELECT time_bucket($1, time) AS time, provider, COUNT(DISTINCT message_id) AS count
                 FROM deferred_events
                 WHERE time BETWEEN $2 AND $3 AND user_id = $4
                 GROUP BY 1, 2 ORDER BY 1, 2`
	case "open":
		query = `SELECT time_bucket($1, time) AS time, provider, COUNT(DISTINCT message_id) AS count
                 FROM open_events
                 WHERE time BETWEEN $2 AND $3 AND user_id = $4
                 GROUP BY 1, 2 ORDER BY 1, 2`
	case "dropped":
		query = `SELECT time_bucket($1, time) AS time, provider, COUNT(DISTINCT message_id) AS count
                 FROM dropped_events
                 WHERE time BETWEEN $2 AND $3 AND user_id = $4
                 GROUP BY 1, 2 ORDER BY 1, 2`
	default:
		http.Error(w, "Invalid event type", http.StatusBadRequest)
		return
	}

	// Execute the query
	rows, err := ec.DB.Query(query, bucketSize, startTime, endTime, authUser.ID)
	if err != nil {
		http.Error(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Process the results
	type Result struct {
		Time     time.Time `json:"time"`
		Provider string    `json:"provider"`
		Count    int       `json:"count"`
	}
	var results []Result

	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Time, &r.Provider, &r.Count); err != nil {
			http.Error(w, "Error scanning row: "+err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, r)
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
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

// func (ec *EventController) GetProviderEventStatsByType(w http.ResponseWriter, r *http.Request) {
// 	authUser, ok := r.Context().Value(middleware.AuthUserKey).(*models.User)
// 	if !ok {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	vars := mux.Vars(r)
// 	providerName := vars["provider"]
// 	eventType := vars["event"]

// 	// Validate provider and event type
// 	if !isValidProvider(providerName) {
// 		http.Error(w, "Invalid provider", http.StatusBadRequest)
// 		return
// 	}
// 	if !isValidEventType(eventType) {
// 		http.Error(w, "Invalid event type", http.StatusBadRequest)
// 		return
// 	}

// 	startDateStr := r.URL.Query().Get("start_date")
// 	endDateStr := r.URL.Query().Get("end_date")

// 	startTime, err := time.Parse("2006-01-02", startDateStr)
// 	if err != nil {
// 		http.Error(w, "Invalid start_date format. Use YYYY-MM-DD", http.StatusBadRequest)
// 		return
// 	}

// 	endTime, err := time.Parse("2006-01-02", endDateStr)
// 	if err != nil {
// 		http.Error(w, "Invalid end_date format. Use YYYY-MM-DD", http.StatusBadRequest)
// 		return
// 	}

// 	// Set end time to the end of the day
// 	endTime = endTime.Add(24*time.Hour - time.Second)

// 	stats, err := models.GetProviderEventStatsByType(ec.DB, authUser.ID, providerName, eventType, startTime, endTime)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
// 		return
// 	}

// 	w.Header().Set("Content-Type", "application/json")
// 	json.NewEncoder(w).Encode(stats)
// }

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

	eventType := r.URL.Query().Get("event_type")
	if eventType == "" {
		http.Error(w, "Event type is required", http.StatusBadRequest)
		return
	}
	if !isValidEventType(eventType) {
		http.Error(w, "Invalid event type", http.StatusBadRequest)
		return
	}

	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	timeBucket := r.URL.Query().Get("time_bucket")

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

	// Validate and set default time bucket
	if timeBucket == "" {
		timeBucket = "1 day"
	} else {
		// Remove any spaces from the time bucket
		timeBucket = strings.ReplaceAll(timeBucket, " ", "")
		validBuckets := map[string]bool{
			"1minute": true, "5minutes": true, "15minutes": true, "30minutes": true,
			"1hour": true, "1day": true, "1week": true, "1month": true,
		}
		if !validBuckets[timeBucket] {
			http.Error(w, "Invalid time_bucket. Valid values are: 1 minute, 5 minutes, 15 minutes, 30 minutes, 1 hour, 1 day, 1 week, 1 month", http.StatusBadRequest)
			return
		}
		// Add space before the time unit for PostgreSQL time_bucket function
		if len(timeBucket) > 1 {
			timeBucket = timeBucket[:1] + " " + timeBucket[1:]
		}
	}

	stats, err := models.GetProviderEventStatsByType(ec.DB, authUser.ID, providerName, eventType, startTime, endTime, timeBucket)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
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
