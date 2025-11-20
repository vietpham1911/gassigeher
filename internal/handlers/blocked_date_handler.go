package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tranm/gassigeher/internal/config"
	"github.com/tranm/gassigeher/internal/middleware"
	"github.com/tranm/gassigeher/internal/models"
	"github.com/tranm/gassigeher/internal/repository"
)

// BlockedDateHandler handles blocked date-related HTTP requests
type BlockedDateHandler struct {
	db              *sql.DB
	cfg             *config.Config
	blockedDateRepo *repository.BlockedDateRepository
}

// NewBlockedDateHandler creates a new blocked date handler
func NewBlockedDateHandler(db *sql.DB, cfg *config.Config) *BlockedDateHandler {
	return &BlockedDateHandler{
		db:              db,
		cfg:             cfg,
		blockedDateRepo: repository.NewBlockedDateRepository(db),
	}
}

// ListBlockedDates lists all blocked dates
func (h *BlockedDateHandler) ListBlockedDates(w http.ResponseWriter, r *http.Request) {
	blockedDates, err := h.blockedDateRepo.FindAll()
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get blocked dates")
		return
	}

	respondJSON(w, http.StatusOK, blockedDates)
}

// CreateBlockedDate creates a new blocked date (admin only)
func (h *BlockedDateHandler) CreateBlockedDate(w http.ResponseWriter, r *http.Request) {
	// Get admin user ID from context
	userID, _ := r.Context().Value(middleware.UserIDKey).(int)

	// Parse request
	var req models.CreateBlockedDateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create blocked date
	blockedDate := &models.BlockedDate{
		Date:      req.Date,
		Reason:    req.Reason,
		CreatedBy: userID,
	}

	if err := h.blockedDateRepo.Create(blockedDate); err != nil {
		if err.Error() == "date is already blocked" {
			respondError(w, http.StatusConflict, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to create blocked date")
		return
	}

	respondJSON(w, http.StatusCreated, blockedDate)
}

// DeleteBlockedDate deletes a blocked date (admin only)
func (h *BlockedDateHandler) DeleteBlockedDate(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid blocked date ID")
		return
	}

	// Delete blocked date
	if err := h.blockedDateRepo.Delete(id); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete blocked date")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Blocked date deleted successfully"})
}
