package models

import "time"

// BlockedDate represents a date that is blocked from bookings
type BlockedDate struct {
	ID        int       `json:"id"`
	Date      string    `json:"date"` // YYYY-MM-DD format
	Reason    string    `json:"reason"`
	CreatedBy int       `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateBlockedDateRequest represents a request to block a date
type CreateBlockedDateRequest struct {
	Date   string `json:"date"`
	Reason string `json:"reason"`
}

// Validate validates the create blocked date request
func (r *CreateBlockedDateRequest) Validate() error {
	if r.Date == "" {
		return &ValidationError{Field: "date", Message: "Date is required"}
	}

	// Validate date format (YYYY-MM-DD)
	if _, err := time.Parse("2006-01-02", r.Date); err != nil {
		return &ValidationError{Field: "date", Message: "Date must be in YYYY-MM-DD format"}
	}

	if r.Reason == "" {
		return &ValidationError{Field: "reason", Message: "Reason is required"}
	}

	return nil
}
