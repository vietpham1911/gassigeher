package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tranm/gassigeher/internal/config"
	"github.com/tranm/gassigeher/internal/models"
	"github.com/tranm/gassigeher/internal/repository"
)

// SettingsHandler handles system settings-related HTTP requests
type SettingsHandler struct {
	db           *sql.DB
	cfg          *config.Config
	settingsRepo *repository.SettingsRepository
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(db *sql.DB, cfg *config.Config) *SettingsHandler {
	return &SettingsHandler{
		db:           db,
		cfg:          cfg,
		settingsRepo: repository.NewSettingsRepository(db),
	}
}

// GetAllSettings gets all system settings (admin only)
func (h *SettingsHandler) GetAllSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.settingsRepo.GetAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get settings")
		return
	}

	respondJSON(w, http.StatusOK, settings)
}

// UpdateSetting updates a system setting (admin only)
func (h *SettingsHandler) UpdateSetting(w http.ResponseWriter, r *http.Request) {
	// Get key from URL
	vars := mux.Vars(r)
	key := vars["key"]

	// Parse request
	var req models.UpdateSettingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update setting
	if err := h.settingsRepo.Update(key, req.Value); err != nil {
		if err.Error() == "setting not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update setting")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Setting updated successfully"})
}
