package models

import "time"

// SystemSetting represents a system configuration setting
type SystemSetting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateSettingRequest represents a request to update a setting
type UpdateSettingRequest struct {
	Value string `json:"value"`
}

// Validate validates the update setting request
func (r *UpdateSettingRequest) Validate() error {
	if r.Value == "" {
		return &ValidationError{Field: "value", Message: "Value is required"}
	}

	return nil
}
