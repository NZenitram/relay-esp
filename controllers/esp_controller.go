// controllers/esp_controller.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

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
	authUser, ok := r.Context().Value("authUser").(*models.User)
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
