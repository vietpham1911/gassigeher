package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/tranm/gassigeher/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (
			name, email, phone, password_hash, experience_level, is_verified,
			is_active, verification_token, verification_token_expires,
			terms_accepted_at, last_activity_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(
		query,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.ExperienceLevel,
		user.IsVerified,
		user.IsActive,
		user.VerificationToken,
		user.VerificationTokenExpires,
		user.TermsAcceptedAt,
		user.LastActivityAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	user.ID = int(id)
	return nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, experience_level,
		       is_verified, is_active, is_deleted, verification_token,
		       verification_token_expires, password_reset_token,
		       password_reset_expires, profile_photo, anonymous_id,
		       terms_accepted_at, last_activity_at, deactivated_at,
		       deactivation_reason, reactivated_at, deleted_at,
		       created_at, updated_at
		FROM users
		WHERE email = ? AND is_deleted = 0
	`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.ExperienceLevel,
		&user.IsVerified,
		&user.IsActive,
		&user.IsDeleted,
		&user.VerificationToken,
		&user.VerificationTokenExpires,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.ProfilePhoto,
		&user.AnonymousID,
		&user.TermsAcceptedAt,
		&user.LastActivityAt,
		&user.DeactivatedAt,
		&user.DeactivationReason,
		&user.ReactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id int) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, experience_level,
		       is_verified, is_active, is_deleted, verification_token,
		       verification_token_expires, password_reset_token,
		       password_reset_expires, profile_photo, anonymous_id,
		       terms_accepted_at, last_activity_at, deactivated_at,
		       deactivation_reason, reactivated_at, deleted_at,
		       created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.ExperienceLevel,
		&user.IsVerified,
		&user.IsActive,
		&user.IsDeleted,
		&user.VerificationToken,
		&user.VerificationTokenExpires,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.ProfilePhoto,
		&user.AnonymousID,
		&user.TermsAcceptedAt,
		&user.LastActivityAt,
		&user.DeactivatedAt,
		&user.DeactivationReason,
		&user.ReactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// FindByVerificationToken finds a user by verification token
func (r *UserRepository) FindByVerificationToken(token string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, experience_level,
		       is_verified, is_active, is_deleted, verification_token,
		       verification_token_expires, password_reset_token,
		       password_reset_expires, profile_photo, anonymous_id,
		       terms_accepted_at, last_activity_at, deactivated_at,
		       deactivation_reason, reactivated_at, deleted_at,
		       created_at, updated_at
		FROM users
		WHERE verification_token = ? AND is_deleted = 0
	`

	user := &models.User{}
	err := r.db.QueryRow(query, token).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.ExperienceLevel,
		&user.IsVerified,
		&user.IsActive,
		&user.IsDeleted,
		&user.VerificationToken,
		&user.VerificationTokenExpires,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.ProfilePhoto,
		&user.AnonymousID,
		&user.TermsAcceptedAt,
		&user.LastActivityAt,
		&user.DeactivatedAt,
		&user.DeactivationReason,
		&user.ReactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// FindByPasswordResetToken finds a user by password reset token
func (r *UserRepository) FindByPasswordResetToken(token string) (*models.User, error) {
	query := `
		SELECT id, name, email, phone, password_hash, experience_level,
		       is_verified, is_active, is_deleted, verification_token,
		       verification_token_expires, password_reset_token,
		       password_reset_expires, profile_photo, anonymous_id,
		       terms_accepted_at, last_activity_at, deactivated_at,
		       deactivation_reason, reactivated_at, deleted_at,
		       created_at, updated_at
		FROM users
		WHERE password_reset_token = ? AND is_deleted = 0
	`

	user := &models.User{}
	err := r.db.QueryRow(query, token).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Phone,
		&user.PasswordHash,
		&user.ExperienceLevel,
		&user.IsVerified,
		&user.IsActive,
		&user.IsDeleted,
		&user.VerificationToken,
		&user.VerificationTokenExpires,
		&user.PasswordResetToken,
		&user.PasswordResetExpires,
		&user.ProfilePhoto,
		&user.AnonymousID,
		&user.TermsAcceptedAt,
		&user.LastActivityAt,
		&user.DeactivatedAt,
		&user.DeactivationReason,
		&user.ReactivatedAt,
		&user.DeletedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	return user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	query := `
		UPDATE users SET
			name = ?,
			email = ?,
			phone = ?,
			password_hash = ?,
			experience_level = ?,
			is_verified = ?,
			is_active = ?,
			is_deleted = ?,
			verification_token = ?,
			verification_token_expires = ?,
			password_reset_token = ?,
			password_reset_expires = ?,
			profile_photo = ?,
			anonymous_id = ?,
			last_activity_at = ?,
			deactivated_at = ?,
			deactivation_reason = ?,
			reactivated_at = ?,
			deleted_at = ?,
			updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.Exec(
		query,
		user.Name,
		user.Email,
		user.Phone,
		user.PasswordHash,
		user.ExperienceLevel,
		user.IsVerified,
		user.IsActive,
		user.IsDeleted,
		user.VerificationToken,
		user.VerificationTokenExpires,
		user.PasswordResetToken,
		user.PasswordResetExpires,
		user.ProfilePhoto,
		user.AnonymousID,
		user.LastActivityAt,
		user.DeactivatedAt,
		user.DeactivationReason,
		user.ReactivatedAt,
		user.DeletedAt,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// UpdateLastActivity updates the last activity timestamp
func (r *UserRepository) UpdateLastActivity(userID int) error {
	query := `UPDATE users SET last_activity_at = ? WHERE id = ?`
	_, err := r.db.Exec(query, time.Now(), userID)
	return err
}
