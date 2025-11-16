package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                       int        `json:"id"`
	Name                     string     `json:"name"`
	Email                    *string    `json:"email,omitempty"`
	Phone                    *string    `json:"phone,omitempty"`
	PasswordHash             *string    `json:"-"`
	ExperienceLevel          string     `json:"experience_level"`
	IsVerified               bool       `json:"is_verified"`
	IsActive                 bool       `json:"is_active"`
	IsDeleted                bool       `json:"is_deleted"`
	VerificationToken        *string    `json:"-"`
	VerificationTokenExpires *time.Time `json:"-"`
	PasswordResetToken       *string    `json:"-"`
	PasswordResetExpires     *time.Time `json:"-"`
	ProfilePhoto             *string    `json:"profile_photo,omitempty"`
	AnonymousID              *string    `json:"anonymous_id,omitempty"`
	TermsAcceptedAt          time.Time  `json:"terms_accepted_at"`
	LastActivityAt           time.Time  `json:"last_activity_at"`
	DeactivatedAt            *time.Time `json:"deactivated_at,omitempty"`
	DeactivationReason       *string    `json:"deactivation_reason,omitempty"`
	ReactivatedAt            *time.Time `json:"reactivated_at,omitempty"`
	DeletedAt                *time.Time `json:"deleted_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
}

// RegisterRequest represents the registration payload
type RegisterRequest struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	AcceptTerms     bool   `json:"accept_terms"`
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token   string `json:"token"`
	User    *User  `json:"user"`
	IsAdmin bool   `json:"is_admin"`
}

// VerifyEmailRequest represents email verification payload
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ForgotPasswordRequest represents forgot password payload
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents password reset payload
type ResetPasswordRequest struct {
	Token           string `json:"token"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

// ChangePasswordRequest represents change password payload
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// UpdateProfileRequest represents profile update payload
type UpdateProfileRequest struct {
	Name  *string `json:"name,omitempty"`
	Email *string `json:"email,omitempty"`
	Phone *string `json:"phone,omitempty"`
}
