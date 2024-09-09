// controllers/event_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nzenitram/relay-esp/models"
)

type EventController struct {
	DB *sql.DB
}

func NewEventController(db *sql.DB) *EventController {
	return &EventController{DB: db}
}

func (ec *EventController) GetEvents(w http.ResponseWriter, r *http.Request) {
	authUser, ok := r.Context().Value("authUser").(*models.User)
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
	authUser, ok := r.Context().Value("authUser").(*models.User)
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
