# Installation and Self-Services - Implementation Plan

## Overview

This document provides a detailed, phase-by-phase implementation plan for the automatic installation system, Super Admin management, and self-service admin operations described in [InstallationAndSelfServices.md](InstallationAndSelfServices.md).

**Goals:**
1. Zero-config first-time installation with automatic seed data
2. Super Admin system for managing other administrators
3. File-based Super Admin password management (simple, no complex recovery)
4. Database-driven admin privileges (replacing config-based ADMIN_EMAILS)
5. Self-service admin promotion/demotion without developer assistance

**Design Philosophy:** Keep it simple. All operations via Web UI. Super Admin password management via file-based system.

---

## Table of Contents

1. [Phase 1: Database Schema & Models](#phase-1-database-schema--models)
2. [Phase 2: Super Admin Service & Seed Data](#phase-2-super-admin-service--seed-data)
3. [Phase 3: Authentication Updates](#phase-3-authentication-updates)
4. [Phase 4: Backend API Endpoints](#phase-4-backend-api-endpoints)
5. [Phase 5: Frontend UI Updates](#phase-5-frontend-ui-updates)
6. [Phase 6: Protection & Security](#phase-6-protection--security)
7. [Phase 7: Testing & Documentation](#phase-7-testing--documentation)
8. [Implementation Checklist](#implementation-checklist)
9. [Testing Strategy](#testing-strategy)
10. [Rollout Plan](#rollout-plan)

---

## Phase 1: Database Schema & Models

**Objective:** Add database support for admin and super admin flags, update models to reflect new structure.

### 1.1 Database Migration

**File:** `internal/database/migrations.go`

**Tasks:**

1. **Add admin flag columns to users table**
   - `is_admin BOOLEAN DEFAULT FALSE`
   - `is_super_admin BOOLEAN DEFAULT FALSE`

2. **Create indexes for performance**
   - `idx_users_admin` on `is_admin`
   - `idx_users_super_admin` on `is_super_admin`

3. **Add unique constraint**
   - `idx_one_super_admin` to ensure only one super admin exists

**SQL for each database:**

```sql
-- SQLite
ALTER TABLE users ADD COLUMN is_admin INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN is_super_admin INTEGER DEFAULT 0;
CREATE INDEX idx_users_admin ON users(is_admin);
CREATE INDEX idx_users_super_admin ON users(is_super_admin);
CREATE UNIQUE INDEX idx_one_super_admin ON users(is_super_admin) WHERE is_super_admin = 1;

-- MySQL
ALTER TABLE users ADD COLUMN is_admin TINYINT(1) DEFAULT 0;
ALTER TABLE users ADD COLUMN is_super_admin TINYINT(1) DEFAULT 0;
CREATE INDEX idx_users_admin ON users(is_admin);
CREATE INDEX idx_users_super_admin ON users(is_super_admin);
-- Note: MySQL doesn't support partial unique indexes, handle in application logic

-- PostgreSQL
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN is_super_admin BOOLEAN DEFAULT FALSE;
CREATE INDEX idx_users_admin ON users(is_admin);
CREATE INDEX idx_users_super_admin ON users(is_super_admin);
CREATE UNIQUE INDEX idx_one_super_admin ON users(is_super_admin) WHERE is_super_admin = TRUE;
```

**Implementation:**

```go
// internal/database/migrations.go

func RunMigrations(db *sql.DB) error {
    // ... existing migrations ...

    // Migration: Add admin flags
    _, err := db.Exec(`
        ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;
    `)
    if err != nil && !strings.Contains(err.Error(), "duplicate column") {
        log.Printf("Warning: Error adding is_admin column: %v", err)
    }

    _, err = db.Exec(`
        ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN DEFAULT FALSE;
    `)
    if err != nil && !strings.Contains(err.Error(), "duplicate column") {
        log.Printf("Warning: Error adding is_super_admin column: %v", err)
    }

    // Create indexes
    db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_admin ON users(is_admin)`)
    db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_super_admin ON users(is_super_admin)`)

    // Unique constraint (SQLite/PostgreSQL only)
    db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_one_super_admin ON users(is_super_admin) WHERE is_super_admin = TRUE`)

    log.Println("✓ Admin flags migration completed")
    return nil
}
```

### 1.2 Update User Model

**File:** `internal/models/user.go`

**Tasks:**

1. Add `IsAdmin` field to User struct
2. Add `IsSuperAdmin` field to User struct
3. Update JSON tags for API responses
4. Update validation if needed

**Implementation:**

```go
// internal/models/user.go

type User struct {
    ID               int       `json:"id"`
    Name             string    `json:"name"`
    Email            string    `json:"email"`
    PasswordHash     string    `json:"-"` // Never expose in JSON
    Phone            string    `json:"phone"`
    ExperienceLevel  string    `json:"experience_level"`
    ProfilePhoto     string    `json:"profile_photo,omitempty"`

    // Admin flags (NEW)
    IsAdmin          bool      `json:"is_admin"`
    IsSuperAdmin     bool      `json:"is_super_admin"`

    // Account status
    IsActive         bool      `json:"is_active"`
    IsVerified       bool      `json:"is_verified"`
    IsDeleted        bool      `json:"is_deleted"`

    // Tracking
    LastActivityAt   time.Time `json:"last_activity_at"`
    CreatedAt        time.Time `json:"created_at"`
    DeactivatedAt    *time.Time `json:"deactivated_at,omitempty"`
    DeactivationReason string  `json:"deactivation_reason,omitempty"`
    AnonymousID      string    `json:"anonymous_id,omitempty"`

    // Verification
    VerificationToken       string    `json:"-"`
    VerificationTokenExpiry time.Time `json:"-"`
    ResetToken              string    `json:"-"`
    ResetTokenExpiry        time.Time `json:"-"`
}
```

### 1.3 Update Repository Methods

**File:** `internal/repository/user_repository.go`

**Tasks:**

1. Update `FindByID` to select admin flags
2. Update `FindByEmail` to select admin flags
3. Update `GetAllUsers` to select admin flags
4. Add `PromoteToAdmin` method
5. Add `DemoteAdmin` method
6. Update any other methods that return User objects

**Implementation:**

```go
// internal/repository/user_repository.go

func (r *UserRepository) FindByID(id int) (*models.User, error) {
    user := &models.User{}
    query := `
        SELECT id, name, email, password_hash, phone, experience_level, profile_photo,
               is_admin, is_super_admin, is_active, is_verified, is_deleted,
               last_activity_at, created_at, deactivated_at, deactivation_reason, anonymous_id
        FROM users
        WHERE id = ?
    `
    err := r.db.QueryRow(query, id).Scan(
        &user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Phone,
        &user.ExperienceLevel, &user.ProfilePhoto,
        &user.IsAdmin, &user.IsSuperAdmin, // NEW
        &user.IsActive, &user.IsVerified, &user.IsDeleted,
        &user.LastActivityAt, &user.CreatedAt, &user.DeactivatedAt,
        &user.DeactivationReason, &user.AnonymousID,
    )
    if err != nil {
        return nil, err
    }
    return user, nil
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    // Similar to FindByID, include is_admin and is_super_admin
}

func (r *UserRepository) PromoteToAdmin(userID int) error {
    query := `UPDATE users SET is_admin = ? WHERE id = ?`
    _, err := r.db.Exec(query, true, userID)
    return err
}

func (r *UserRepository) DemoteAdmin(userID int) error {
    query := `UPDATE users SET is_admin = ? WHERE id = ?`
    _, err := r.db.Exec(query, false, userID)
    return err
}

func (r *UserRepository) IsSuperAdmin(userID int) (bool, error) {
    var isSuperAdmin bool
    query := `SELECT is_super_admin FROM users WHERE id = ?`
    err := r.db.QueryRow(query, userID).Scan(&isSuperAdmin)
    return isSuperAdmin, err
}
```

**Deliverables:**
- ✅ Migration script for all 3 databases (SQLite, MySQL, PostgreSQL)
- ✅ Updated User model with admin flags
- ✅ Updated repository methods
- ✅ Database indexes created

**// DONE - Phase 1 Complete**

---

## Phase 2: Super Admin Service & Seed Data

**Objective:** Create automatic seed data generation for first-time installations and file-based Super Admin password management.

### 2.1 Seed Data Generation

**File:** `internal/database/seed.go` (new file)

**Tasks:**

1. Create `SeedDatabase()` function
2. Check if database is empty (count users)
3. Generate Super Admin with random password
4. Generate 3 test users (green, blue, orange levels)
5. Generate 5 test dogs (mix of categories)
6. Generate 3 test bookings (past, present, future)
7. Initialize default system settings
8. Write credentials to file and console

**Implementation Structure:**

```go
// internal/database/seed.go

package database

import (
    "database/sql"
    "fmt"
    "log"
    "math/rand"
    "os"
    "time"
    "gassigeher/internal/config"
    "gassigeher/internal/models"
    "golang.org/x/crypto/bcrypt"
)

type TestUser struct {
    Name     string
    Email    string
    Password string
    Level    string
}

func SeedDatabase(db *sql.DB, cfg *config.Config) error {
    // 1. Check if database is empty
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        return fmt.Errorf("failed to check users count: %w", err)
    }

    if count > 0 {
        log.Println("Database already seeded, skipping seed data generation")
        return nil
    }

    log.Println("Empty database detected, generating seed data...")

    // 2. Validate Super Admin email
    if cfg.SuperAdminEmail == "" {
        return fmt.Errorf("SUPER_ADMIN_EMAIL not set in .env - cannot create Super Admin")
    }

    // 3. Generate Super Admin
    superAdminPassword := generateSecurePassword(20)
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(superAdminPassword), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    _, err = db.Exec(`
        INSERT INTO users (
            id, name, email, password_hash, experience_level,
            is_admin, is_super_admin, is_active, is_verified,
            last_activity_at, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, 1, "Super Admin", cfg.SuperAdminEmail, string(hashedPassword), "orange",
       true, true, true, true, time.Now(), time.Now())

    if err != nil {
        return fmt.Errorf("failed to create Super Admin: %w", err)
    }

    // 4. Generate test users
    testUsers := generateTestUsers(db)

    // 5. Generate dogs
    err = generateDogs(db)
    if err != nil {
        return err
    }

    // 6. Generate bookings
    err = generateBookings(db)
    if err != nil {
        return err
    }

    // 7. Initialize default settings
    err = initializeSystemSettings(db)
    if err != nil {
        return err
    }

    // 8. Write credentials to file
    err = writeCredentialsFile(cfg.SuperAdminEmail, superAdminPassword)
    if err != nil {
        log.Printf("Warning: Failed to write credentials file: %v", err)
    }

    // 9. Print setup complete message
    printSetupComplete(cfg.SuperAdminEmail, superAdminPassword, testUsers)

    log.Println("✓ Seed data generation completed successfully")
    return nil
}

func generateSecurePassword(length int) string {
    // Character sets for password generation
    lowercase := "abcdefghijklmnopqrstuvwxyz"
    uppercase := "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    numbers := "0123456789"
    special := "!@#$%^&*"
    allChars := lowercase + uppercase + numbers + special

    rand.Seed(time.Now().UnixNano())

    password := make([]byte, length)
    // Ensure at least one of each type
    password[0] = lowercase[rand.Intn(len(lowercase))]
    password[1] = uppercase[rand.Intn(len(uppercase))]
    password[2] = numbers[rand.Intn(len(numbers))]
    password[3] = special[rand.Intn(len(special))]

    // Fill rest randomly
    for i := 4; i < length; i++ {
        password[i] = allChars[rand.Intn(len(allChars))]
    }

    // Shuffle
    rand.Shuffle(len(password), func(i, j int) {
        password[i], password[j] = password[j], password[i]
    })

    return string(password)
}

func generateTestUsers(db *sql.DB) []TestUser {
    users := []TestUser{
        {Name: "Test Walker (Green)", Email: "green-walker@test.com", Level: "green"},
        {Name: "Test Walker (Blue)", Email: "blue-walker@test.com", Level: "blue"},
        {Name: "Test Walker (Orange)", Email: "orange-walker@test.com", Level: "orange"},
    }

    for i := range users {
        users[i].Password = generateSecurePassword(12)
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(users[i].Password), bcrypt.DefaultCost)

        db.Exec(`
            INSERT INTO users (name, email, password_hash, experience_level,
                             is_admin, is_super_admin, is_active, is_verified,
                             last_activity_at, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, users[i].Name, users[i].Email, string(hashedPassword), users[i].Level,
           false, false, true, true, time.Now(), time.Now())
    }

    return users
}

func generateDogs(db *sql.DB) error {
    dogs := []struct {
        Name     string
        Category string
        Breed    string
        Age      int
        Gender   string
        Desc     string
    }{
        {"Bella", "green", "Labrador Retriever", 3, "female", "Freundlicher und ruhiger Hund, perfekt für Anfänger"},
        {"Max", "green", "Golden Retriever", 5, "male", "Sehr gutmütig und leicht zu führen"},
        {"Luna", "blue", "Deutscher Schäferhund", 4, "female", "Aktiv und intelligent, braucht erfahrene Führung"},
        {"Charlie", "blue", "Border Collie", 2, "male", "Energiegeladen und verspielt, benötigt Erfahrung"},
        {"Rocky", "orange", "Belgischer Malinois", 6, "male", "Sehr anspruchsvoll, nur für erfahrene Hundeführer"},
    }

    for _, dog := range dogs {
        _, err := db.Exec(`
            INSERT INTO dogs (name, category, breed, age, gender,
                            description, special_needs, is_available, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, dog.Name, dog.Category, dog.Breed, dog.Age, dog.Gender,
           dog.Desc, "Keine besonderen Bedürfnisse", true, time.Now())
        if err != nil {
            return err
        }
    }

    return nil
}

func generateBookings(db *sql.DB) error {
    today := time.Now()
    yesterday := today.AddDate(0, 0, -1)
    tomorrow := today.AddDate(0, 0, 1)

    bookings := []struct {
        UserID int
        DogID  int
        Date   time.Time
        Time   string
        Status string
        Type   string
    }{
        {2, 1, yesterday, "09:00", "completed", "short"},
        {3, 2, today, "14:00", "scheduled", "long"},
        {4, 3, tomorrow, "10:30", "scheduled", "short"},
    }

    for _, booking := range bookings {
        _, err := db.Exec(`
            INSERT INTO bookings (user_id, dog_id, date, scheduled_time,
                                walk_type, status, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        `, booking.UserID, booking.DogID,
           booking.Date.Format("2006-01-02"), booking.Time,
           booking.Type, booking.Status, time.Now())
        if err != nil {
            return err
        }
    }

    return nil
}

func initializeSystemSettings(db *sql.DB) error {
    settings := []struct {
        Key   string
        Value string
    }{
        {"booking_advance_days", "14"},
        {"cancellation_notice_hours", "12"},
        {"auto_deactivation_days", "365"},
    }

    for _, setting := range settings {
        _, err := db.Exec(`
            INSERT INTO system_settings (setting_key, setting_value, updated_at)
            VALUES (?, ?, ?)
            ON CONFLICT(setting_key) DO UPDATE SET setting_value = ?, updated_at = ?
        `, setting.Key, setting.Value, time.Now(), setting.Value, time.Now())
        if err != nil {
            return err
        }
    }

    return nil
}

func writeCredentialsFile(email, password string) error {
    now := time.Now().Format("2006-01-02 15:04:05")
    content := fmt.Sprintf(`=============================================================
GASSIGEHER - SUPER ADMIN CREDENTIALS
=============================================================

EMAIL: %s
PASSWORD: %s

CREATED: %s
LAST UPDATED: %s

=============================================================
HOW TO CHANGE PASSWORD:
=============================================================

1. Edit the PASSWORD line above with your new password
2. Save this file
3. Restart the Gassigeher server
4. Server will hash and save the new password
5. This file will be updated with confirmation

IMPORTANT:
- Keep this file secure (never commit to git)
- This is the ONLY way to change Super Admin password
- Super Admin email cannot be changed (defined in .env)

=============================================================
`, email, password, now, now)

    err := os.WriteFile("SUPER_ADMIN_CREDENTIALS.txt", []byte(content), 0600)
    if err != nil {
        return err
    }

    log.Println("✓ Credentials file written: SUPER_ADMIN_CREDENTIALS.txt")
    return nil
}

func printSetupComplete(superAdminEmail, superAdminPassword string, testUsers []TestUser) {
    fmt.Println("\n=============================================================")
    fmt.Println("  GASSIGEHER - INSTALLATION COMPLETE")
    fmt.Println("=============================================================")
    fmt.Println()
    fmt.Println("SUPER ADMIN CREDENTIALS (SAVE THESE!):")
    fmt.Printf("  Email:    %s\n", superAdminEmail)
    fmt.Printf("  Password: %s\n", superAdminPassword)
    fmt.Println()
    fmt.Println("TEST USER CREDENTIALS:")
    for i, user := range testUsers {
        fmt.Printf("  %d. %s / %s\n", i+1, user.Email, user.Password)
    }
    fmt.Println()
    fmt.Println("IMPORTANT:")
    fmt.Println("- Super Admin password saved to: SUPER_ADMIN_CREDENTIALS.txt")
    fmt.Println("- Change Super Admin password: Edit file and restart server")
    fmt.Println("- Test users can be deleted after setup")
    fmt.Println()
    fmt.Println("=============================================================")
    fmt.Println()
}
```

### 2.2 Super Admin Password File Service

**File:** `internal/services/super_admin_service.go` (new file)

**Tasks:**

1. Create service to read credentials file
2. Parse email and password from file
3. Compare password hash to detect changes
4. Update database when password changes
5. Rewrite file with confirmation

**Implementation:**

```go
// internal/services/super_admin_service.go

package services

import (
    "database/sql"
    "errors"
    "fmt"
    "log"
    "os"
    "strings"
    "time"
    "gassigeher/internal/config"
    "golang.org/x/crypto/bcrypt"
)

type SuperAdminService struct {
    db  *sql.DB
    cfg *config.Config
}

func NewSuperAdminService(db *sql.DB, cfg *config.Config) *SuperAdminService {
    return &SuperAdminService{
        db:  db,
        cfg: cfg,
    }
}

// CheckAndUpdatePassword reads credentials file and updates password if changed
func (s *SuperAdminService) CheckAndUpdatePassword() error {
    filePath := "SUPER_ADMIN_CREDENTIALS.txt"

    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        log.Println("Super Admin credentials file not found (okay for existing installations)")
        return nil
    }

    // Read file
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read credentials file: %w", err)
    }

    // Parse EMAIL and PASSWORD lines
    email, password, createdTime, err := parseCredentialsFile(string(content))
    if err != nil {
        return fmt.Errorf("failed to parse credentials file: %w", err)
    }

    // Verify email matches config
    if email != s.cfg.SuperAdminEmail {
        return fmt.Errorf("email in credentials file (%s) doesn't match SUPER_ADMIN_EMAIL in .env (%s)",
            email, s.cfg.SuperAdminEmail)
    }

    // Get current Super Admin password hash from database
    var currentHash string
    err = s.db.QueryRow("SELECT password_hash FROM users WHERE id = 1").Scan(&currentHash)
    if err != nil {
        return fmt.Errorf("failed to get Super Admin from database: %w", err)
    }

    // Check if password has changed
    if bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(password)) == nil {
        // Password unchanged, no action needed
        log.Println("Super Admin password unchanged")
        return nil
    }

    // Password changed! Hash new password and update database
    log.Println("Super Admin password change detected, updating...")
    newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return fmt.Errorf("failed to hash new password: %w", err)
    }

    _, err = s.db.Exec("UPDATE users SET password_hash = ? WHERE id = 1", string(newHash))
    if err != nil {
        return fmt.Errorf("failed to update password in database: %w", err)
    }

    // Update file with confirmation and new timestamp
    err = s.writeUpdatedCredentialsFile(email, password, createdTime, true)
    if err != nil {
        return fmt.Errorf("failed to update credentials file: %w", err)
    }

    log.Println("✓ Super Admin password updated successfully")

    return nil
}

func parseCredentialsFile(content string) (email, password, createdTime string, err error) {
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "EMAIL:") {
            email = strings.TrimSpace(strings.TrimPrefix(line, "EMAIL:"))
        }
        if strings.HasPrefix(line, "PASSWORD:") {
            password = strings.TrimSpace(strings.TrimPrefix(line, "PASSWORD:"))
        }
        if strings.HasPrefix(line, "CREATED:") {
            createdTime = strings.TrimSpace(strings.TrimPrefix(line, "CREATED:"))
        }
    }

    if email == "" || password == "" {
        return "", "", "", errors.New("invalid credentials file format: missing EMAIL or PASSWORD")
    }

    return email, password, createdTime, nil
}

func (s *SuperAdminService) writeUpdatedCredentialsFile(email, password, createdTime string, changed bool) error {
    if createdTime == "" {
        createdTime = time.Now().Format("2006-01-02 15:04:05")
    }

    changeConfirmation := ""
    if changed {
        changeConfirmation = "\nPASSWORD CHANGE CONFIRMED: ✓\n"
    }

    content := fmt.Sprintf(`=============================================================
GASSIGEHER - SUPER ADMIN CREDENTIALS
=============================================================

EMAIL: %s
PASSWORD: %s

CREATED: %s
LAST UPDATED: %s%s

=============================================================
HOW TO CHANGE PASSWORD:
=============================================================

1. Edit the PASSWORD line above with your new password
2. Save this file
3. Restart the Gassigeher server
4. Server will hash and save the new password
5. This file will be updated with confirmation

IMPORTANT:
- Keep this file secure (never commit to git)
- This is the ONLY way to change Super Admin password
- Super Admin email cannot be changed (defined in .env)

=============================================================
`, email, password, createdTime, time.Now().Format("2006-01-02 15:04:05"), changeConfirmation)

    return os.WriteFile("SUPER_ADMIN_CREDENTIALS.txt", []byte(content), 0600)
}
```

### 2.3 Configuration Update

**File:** `internal/config/config.go`

**Tasks:**

1. Add `SuperAdminEmail` field
2. Load from environment variable
3. Validate presence on startup
4. Remove `AdminEmails` field and `IsAdmin()` method

**Implementation:**

```go
// internal/config/config.go

type Config struct {
    // ... existing fields ...

    // Super Admin (NEW - replaces ADMIN_EMAILS)
    SuperAdminEmail string

    // REMOVE: AdminEmails []string

    // ... rest of fields ...
}

func LoadConfig() (*Config, error) {
    // ... existing loading ...

    cfg := &Config{
        // ... existing fields ...

        SuperAdminEmail: os.Getenv("SUPER_ADMIN_EMAIL"),

        // REMOVE: AdminEmails: strings.Split(os.Getenv("ADMIN_EMAILS"), ","),
    }

    // Validate SuperAdminEmail (only for fresh installs)
    // Don't fail on existing installations
    if cfg.SuperAdminEmail == "" {
        log.Println("Warning: SUPER_ADMIN_EMAIL not set in .env")
    }

    return cfg, nil
}

// REMOVE this method entirely:
// func (c *Config) IsAdmin(email string) bool { ... }
```

### 2.4 Integration in main.go

**File:** `cmd/server/main.go`

**Tasks:**

1. Call `SeedDatabase()` after migrations
2. Call `CheckAndUpdatePassword()` on startup
3. Handle errors appropriately

**Implementation:**

```go
// cmd/server/main.go

func main() {
    // ... existing config and database setup ...

    // Run migrations
    err = database.RunMigrations(db)
    if err != nil {
        log.Fatal("Failed to run migrations:", err)
    }

    // NEW: Run seed data (first-time installations)
    err = database.SeedDatabase(db, cfg)
    if err != nil {
        log.Fatal("Failed to seed database:", err)
    }

    // NEW: Check and update Super Admin password
    superAdminService := services.NewSuperAdminService(db, cfg)
    err = superAdminService.CheckAndUpdatePassword()
    if err != nil {
        log.Printf("Warning: Failed to check Super Admin password: %v", err)
        // Don't exit - allow server to start
    }

    // ... rest of server setup ...
}
```

**Deliverables:**
- ✅ Seed data generation system
- ✅ Super Admin service with file-based password management
- ✅ Configuration updates
- ✅ Integration in main.go
- ✅ Credentials file generation

**// DONE - Phase 2 Complete**

---

## Phase 3: Authentication Updates

**Objective:** Update JWT authentication system to use database-driven admin flags instead of config-based system.

**// DONE - Phase 3 Complete**

### 3.1 Update JWT Claims

**File:** `internal/services/auth_service.go`

**Tasks:**

1. Add `IsAdmin` field to Claims struct
2. Add `IsSuperAdmin` field to Claims struct
3. Update `GenerateJWT()` to include admin flags from User model
4. Update `ValidateJWT()` to parse new claims

**Implementation:**

```go
// internal/services/auth_service.go

type Claims struct {
    UserID       int    `json:"user_id"`
    Email        string `json:"email"`
    IsAdmin      bool   `json:"is_admin"`       // NEW
    IsSuperAdmin bool   `json:"is_super_admin"` // NEW
    jwt.RegisteredClaims
}

func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
    claims := &Claims{
        UserID:       user.ID,
        Email:        user.Email,
        IsAdmin:      user.IsAdmin,      // NEW: From database
        IsSuperAdmin: user.IsSuperAdmin, // NEW: From database
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "gassigeher",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(s.jwtSecret))
    if err != nil {
        return "", err
    }

    return tokenString, nil
}

// ValidateJWT already handles Claims parsing, no changes needed
```

### 3.2 Update Middleware

**File:** `internal/middleware/middleware.go`

**Tasks:**

1. Add `IsSuperAdminKey` context key
2. Update `AuthMiddleware` to extract and inject super admin flag
3. Create new `RequireSuperAdmin` middleware
4. Keep existing `RequireAdmin` middleware (unchanged)

**Implementation:**

```go
// internal/middleware/middleware.go

type contextKey string

const (
    UserIDKey       contextKey = "userID"
    EmailKey        contextKey = "email"
    IsAdminKey      contextKey = "isAdmin"
    IsSuperAdminKey contextKey = "isSuperAdmin" // NEW
)

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Authorization header required", http.StatusUnauthorized)
                return
            }

            tokenString := strings.TrimPrefix(authHeader, "Bearer ")
            if tokenString == authHeader {
                http.Error(w, "Bearer token required", http.StatusUnauthorized)
                return
            }

            // Validate token
            token, err := authService.ValidateJWT(tokenString)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            claims, ok := token.Claims.(*services.Claims)
            if !ok {
                http.Error(w, "Invalid token claims", http.StatusUnauthorized)
                return
            }

            // Add to context
            ctx := r.Context()
            ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
            ctx = context.WithValue(ctx, EmailKey, claims.Email)
            ctx = context.WithValue(ctx, IsAdminKey, claims.IsAdmin)
            ctx = context.WithValue(ctx, IsSuperAdminKey, claims.IsSuperAdmin) // NEW

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func RequireAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        isAdmin, ok := r.Context().Value(IsAdminKey).(bool)
        if !ok || !isAdmin {
            http.Error(w, "Forbidden: Admin access required", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// NEW: Super Admin middleware
func RequireSuperAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        isSuperAdmin, ok := r.Context().Value(IsSuperAdminKey).(bool)
        if !ok || !isSuperAdmin {
            http.Error(w, "Forbidden: Super Admin access required", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### 3.3 Update Login Handler

**File:** `internal/handlers/auth_handler.go`

**Tasks:**

1. Update `Login()` to return admin flags in response
2. Ensure JWT generation includes admin flags (already done via GenerateJWT)
3. Update response struct if needed

**Implementation:**

```go
// internal/handlers/auth_handler.go

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    // ... existing login logic ...

    // Generate JWT (already includes admin flags from Phase 3.1)
    token, err := h.authService.GenerateJWT(user)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to generate token")
        return
    }

    // Update last activity
    h.userRepo.UpdateLastActivity(user.ID)

    // Return response with admin flags
    respondJSON(w, http.StatusOK, map[string]interface{}{
        "token":          token,
        "user": map[string]interface{}{
            "id":              user.ID,
            "name":            user.Name,
            "email":           user.Email,
            "experience_level": user.ExperienceLevel,
            "is_admin":        user.IsAdmin,       // NEW
            "is_super_admin":  user.IsSuperAdmin,  // NEW
            "is_active":       user.IsActive,
            "is_verified":     user.IsVerified,
        },
    })
}
```

### 3.4 Update /me Endpoint

**File:** `internal/handlers/user_handler.go`

**Tasks:**

1. Ensure `GetMe()` returns admin flags
2. Already returns full User object, just verify

**Implementation:**

```go
// internal/handlers/user_handler.go

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
    userID, _ := r.Context().Value(middleware.UserIDKey).(int)

    user, err := h.userRepo.FindByID(userID)
    if err != nil {
        respondError(w, http.StatusNotFound, "User not found")
        return
    }

    // User model already includes IsAdmin and IsSuperAdmin fields
    // No changes needed, JSON response will include these automatically
    respondJSON(w, http.StatusOK, user)
}
```

**Deliverables:**
- ✅ Updated JWT claims with admin flags
- ✅ Updated middleware with super admin support
- ✅ New RequireSuperAdmin middleware
- ✅ Updated login and /me endpoints
- ✅ Removed config-based admin system

---

## Phase 4: Backend API Endpoints

**Objective:** Create API endpoints for promoting and demoting users to/from admin role.

**// DONE - Phase 4 Complete**

### 4.1 User Handler Updates

**File:** `internal/handlers/user_handler.go`

**Tasks:**

1. Add `PromoteToAdmin()` handler method
2. Add `DemoteAdmin()` handler method
3. Validate operations (can't demote super admin, can't promote if already admin, etc.)
4. Return updated user object

**Implementation:**

```go
// internal/handlers/user_handler.go

// PromoteToAdmin promotes a user to admin role (Super Admin only)
func (h *UserHandler) PromoteToAdmin(w http.ResponseWriter, r *http.Request) {
    // Extract super admin from context (middleware already verified)
    isSuperAdmin, _ := r.Context().Value(middleware.IsSuperAdminKey).(bool)
    if !isSuperAdmin {
        respondError(w, http.StatusForbidden, "Only Super Admin can promote users")
        return
    }

    // Get user ID from URL
    vars := mux.Vars(r)
    userIDStr := vars["id"]
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    // Get target user
    targetUser, err := h.userRepo.FindByID(userID)
    if err != nil {
        respondError(w, http.StatusNotFound, "User not found")
        return
    }

    // Validation checks
    if targetUser.IsSuperAdmin {
        respondError(w, http.StatusBadRequest, "Cannot modify Super Admin")
        return
    }

    if targetUser.IsAdmin {
        respondError(w, http.StatusBadRequest, "User is already an admin")
        return
    }

    // Promote user
    err = h.userRepo.PromoteToAdmin(userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to promote user")
        return
    }

    // Get updated user
    updatedUser, err := h.userRepo.FindByID(userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to retrieve updated user")
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "message": "User promoted to admin successfully",
        "user":    updatedUser,
    })
}

// DemoteAdmin revokes admin privileges (Super Admin only)
func (h *UserHandler) DemoteAdmin(w http.ResponseWriter, r *http.Request) {
    // Extract super admin from context
    isSuperAdmin, _ := r.Context().Value(middleware.IsSuperAdminKey).(bool)
    if !isSuperAdmin {
        respondError(w, http.StatusForbidden, "Only Super Admin can demote admins")
        return
    }

    // Get user ID from URL
    vars := mux.Vars(r)
    userIDStr := vars["id"]
    userID, err := strconv.Atoi(userIDStr)
    if err != nil {
        respondError(w, http.StatusBadRequest, "Invalid user ID")
        return
    }

    // Get target user
    targetUser, err := h.userRepo.FindByID(userID)
    if err != nil {
        respondError(w, http.StatusNotFound, "User not found")
        return
    }

    // Validation checks
    if targetUser.IsSuperAdmin {
        respondError(w, http.StatusBadRequest, "Cannot demote Super Admin")
        return
    }

    if !targetUser.IsAdmin {
        respondError(w, http.StatusBadRequest, "User is not an admin")
        return
    }

    // Demote user
    err = h.userRepo.DemoteAdmin(userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to demote admin")
        return
    }

    // Get updated user
    updatedUser, err := h.userRepo.FindByID(userID)
    if err != nil {
        respondError(w, http.StatusInternalServerError, "Failed to retrieve updated user")
        return
    }

    respondJSON(w, http.StatusOK, map[string]interface{}{
        "message": "Admin privileges revoked successfully",
        "user":    updatedUser,
    })
}
```

### 4.2 Route Registration

**File:** `cmd/server/main.go`

**Tasks:**

1. Create super admin subrouter with `RequireSuperAdmin` middleware
2. Register promote and demote endpoints
3. Ensure proper ordering (super admin routes after admin routes)

**Implementation:**

```go
// cmd/server/main.go

func main() {
    // ... existing setup ...

    // Protected routes (authenticated users)
    protected := router.PathPrefix("/api").Subrouter()
    protected.Use(middleware.AuthMiddleware(authService))

    // ... existing protected routes ...

    // Admin routes (authenticated + admin)
    admin := protected.PathPrefix("").Subrouter()
    admin.Use(middleware.RequireAdmin)

    // ... existing admin routes ...

    // NEW: Super Admin routes (authenticated + admin + super admin)
    superAdmin := admin.PathPrefix("").Subrouter()
    superAdmin.Use(middleware.RequireSuperAdmin)
    superAdmin.HandleFunc("/admin/users/{id}/promote", userHandler.PromoteToAdmin).Methods("POST")
    superAdmin.HandleFunc("/admin/users/{id}/demote", userHandler.DemoteAdmin).Methods("POST")

    // ... rest of server setup ...
}
```

**Deliverables:**
- ✅ PromoteToAdmin handler
- ✅ DemoteAdmin handler
- ✅ Route registration with RequireSuperAdmin middleware
- ✅ Input validation and error handling

---

## Phase 5: Frontend UI Updates

**Objective:** Update admin user management page to show admin management controls for Super Admin.

### 5.1 API Client Updates

**File:** `frontend/js/api.js`

**Tasks:**

1. Add `promoteToAdmin()` method
2. Add `demoteAdmin()` method
3. Update `getMe()` to return admin flags (already done)

**Implementation:**

```javascript
// frontend/js/api.js

class APIClient {
    // ... existing methods ...

    // NEW: Promote user to admin (Super Admin only)
    async promoteToAdmin(userId) {
        const response = await fetch(`${this.baseURL}/admin/users/${userId}/promote`, {
            method: 'POST',
            headers: this.getHeaders(),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to promote user');
        }

        return response.json();
    }

    // NEW: Revoke admin privileges (Super Admin only)
    async demoteAdmin(userId) {
        const response = await fetch(`${this.baseURL}/admin/users/${userId}/demote`, {
            method: 'POST',
            headers: this.getHeaders(),
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to demote admin');
        }

        return response.json();
    }

    // ... rest of methods ...
}
```

### 5.2 Admin Users Page Updates

**File:** `frontend/admin-users.html`

**Tasks:**

1. Update user table to show admin badges
2. Add promote/demote buttons (conditionally for Super Admin)
3. Implement button click handlers
4. Handle success/error states
5. Refresh user list after operations

**Implementation:**

```html
<!-- frontend/admin-users.html -->

<!DOCTYPE html>
<html lang="de">
<head>
    <!-- ... existing head ... -->
</head>
<body>
    <!-- ... existing navigation ... -->

    <main class="container">
        <h1 data-i18n="admin.users.title">Benutzerverwaltung</h1>

        <div class="users-container">
            <div class="filters">
                <!-- ... existing filters ... -->
            </div>

            <div id="users-table-container">
                <table class="users-table">
                    <thead>
                        <tr>
                            <th data-i18n="admin.users.name">Name</th>
                            <th data-i18n="admin.users.email">Email</th>
                            <th data-i18n="admin.users.level">Level</th>
                            <th data-i18n="admin.users.status">Status</th>
                            <th data-i18n="admin.users.role">Rolle</th>
                            <th data-i18n="admin.users.actions">Aktionen</th>
                        </tr>
                    </thead>
                    <tbody id="users-table-body">
                        <!-- Populated by JavaScript -->
                    </tbody>
                </table>
            </div>
        </div>
    </main>

    <script src="/js/api.js"></script>
    <script src="/js/i18n.js"></script>
    <script>
        let currentUser = null;
        let allUsers = [];

        async function init() {
            try {
                // Load translations
                await window.i18n.load();
                window.i18n.updateElement(document.documentElement);

                // Get current user info
                currentUser = await window.api.getMe();

                // Load users
                await loadUsers();
            } catch (error) {
                console.error('Initialization error:', error);
                alert('Fehler beim Laden der Seite: ' + error.message);
            }
        }

        async function loadUsers(activeOnly = false) {
            try {
                allUsers = await window.api.getUsers(activeOnly);
                renderUsersTable(allUsers);
            } catch (error) {
                console.error('Error loading users:', error);
                alert('Fehler beim Laden der Benutzer: ' + error.message);
            }
        }

        function renderUsersTable(users) {
            const tbody = document.getElementById('users-table-body');
            tbody.innerHTML = '';

            if (users.length === 0) {
                tbody.innerHTML = '<tr><td colspan="6" style="text-align: center;">Keine Benutzer gefunden</td></tr>';
                return;
            }

            users.forEach(user => {
                const row = createUserRow(user);
                tbody.appendChild(row);
            });
        }

        function createUserRow(user) {
            const tr = document.createElement('tr');

            // Name
            const tdName = document.createElement('td');
            tdName.textContent = user.name;
            tr.appendChild(tdName);

            // Email
            const tdEmail = document.createElement('td');
            tdEmail.textContent = user.email || '—';
            tr.appendChild(tdEmail);

            // Experience Level
            const tdLevel = document.createElement('td');
            tdLevel.innerHTML = `<span class="badge badge-${user.experience_level}">${user.experience_level.toUpperCase()}</span>`;
            tr.appendChild(tdLevel);

            // Status
            const tdStatus = document.createElement('td');
            tdStatus.innerHTML = user.is_active ?
                '<span class="badge badge-success">Aktiv</span>' :
                '<span class="badge badge-inactive">Inaktiv</span>';
            tr.appendChild(tdStatus);

            // Role (NEW)
            const tdRole = document.createElement('td');
            if (user.is_super_admin) {
                tdRole.innerHTML = '<span class="badge badge-super-admin">Super Admin</span>';
            } else if (user.is_admin) {
                tdRole.innerHTML = '<span class="badge badge-admin">Admin</span>';
            } else {
                tdRole.textContent = 'Benutzer';
            }
            tr.appendChild(tdRole);

            // Actions
            const tdActions = document.createElement('td');
            tdActions.className = 'actions-cell';

            // View button
            const btnView = document.createElement('button');
            btnView.className = 'btn btn-sm btn-view';
            btnView.textContent = 'Ansehen';
            btnView.onclick = () => viewUser(user.id);
            tdActions.appendChild(btnView);

            // Deactivate button (not for admins or super admin)
            if (!user.is_admin && !user.is_super_admin && user.is_active) {
                const btnDeactivate = document.createElement('button');
                btnDeactivate.className = 'btn btn-sm btn-danger';
                btnDeactivate.textContent = 'Deaktivieren';
                btnDeactivate.onclick = () => deactivateUser(user.id);
                tdActions.appendChild(btnDeactivate);
            }

            // Admin management buttons (NEW - only for Super Admin)
            if (currentUser.is_super_admin && !user.is_super_admin) {
                if (user.is_admin) {
                    // Demote button
                    const btnDemote = document.createElement('button');
                    btnDemote.className = 'btn btn-sm btn-demote';
                    btnDemote.textContent = 'Admin entfernen';
                    btnDemote.onclick = () => demoteAdmin(user.id, user.name);
                    tdActions.appendChild(btnDemote);
                } else {
                    // Promote button
                    const btnPromote = document.createElement('button');
                    btnPromote.className = 'btn btn-sm btn-promote';
                    btnPromote.textContent = 'Zu Admin ernennen';
                    btnPromote.onclick = () => promoteToAdmin(user.id, user.name);
                    tdActions.appendChild(btnPromote);
                }
            }

            tr.appendChild(tdActions);
            return tr;
        }

        async function promoteToAdmin(userId, userName) {
            if (!confirm(`Möchten Sie ${userName} wirklich zum Admin ernennen?\n\nAdmins haben Zugriff auf alle Verwaltungsfunktionen.`)) {
                return;
            }

            try {
                const result = await window.api.promoteToAdmin(userId);
                alert(`${userName} wurde erfolgreich zum Admin ernannt.`);
                await loadUsers(); // Refresh list
            } catch (error) {
                console.error('Error promoting user:', error);
                alert('Fehler beim Ernennen: ' + error.message);
            }
        }

        async function demoteAdmin(userId, userName) {
            if (!confirm(`Möchten Sie ${userName} wirklich die Admin-Rechte entziehen?\n\nDer Benutzer wird zu einem normalen Benutzer herabgestuft.`)) {
                return;
            }

            try {
                const result = await window.api.demoteAdmin(userId);
                alert(`Admin-Rechte von ${userName} wurden erfolgreich entzogen.`);
                await loadUsers(); // Refresh list
            } catch (error) {
                console.error('Error demoting admin:', error);
                alert('Fehler beim Herabstufen: ' + error.message);
            }
        }

        function viewUser(userId) {
            // ... existing view user logic ...
        }

        async function deactivateUser(userId) {
            // ... existing deactivate user logic ...
        }

        // Initialize on page load
        document.addEventListener('DOMContentLoaded', init);
    </script>

    <style>
        /* ... existing styles ... */

        .badge-super-admin {
            background-color: #d32f2f;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
        }

        .badge-admin {
            background-color: #1976d2;
            color: white;
            padding: 4px 8px;
            border-radius: 4px;
            font-size: 12px;
            font-weight: bold;
        }

        .btn-promote {
            background-color: #1976d2;
            color: white;
        }

        .btn-promote:hover {
            background-color: #1565c0;
        }

        .btn-demote {
            background-color: #f57c00;
            color: white;
        }

        .btn-demote:hover {
            background-color: #ef6c00;
        }

        .actions-cell {
            display: flex;
            gap: 8px;
            flex-wrap: wrap;
        }
    </style>
</body>
</html>
```

### 5.3 Translations Update

**File:** `frontend/i18n/de.json`

**Tasks:**

1. Add translations for admin management features

**Implementation:**

```json
{
  "admin": {
    "users": {
      "title": "Benutzerverwaltung",
      "name": "Name",
      "email": "Email",
      "level": "Level",
      "status": "Status",
      "role": "Rolle",
      "actions": "Aktionen",
      "superAdmin": "Super Admin",
      "admin": "Admin",
      "user": "Benutzer",
      "promoteButton": "Zu Admin ernennen",
      "demoteButton": "Admin entfernen",
      "promoteConfirm": "Möchten Sie diesen Benutzer wirklich zum Admin ernennen?\n\nAdmins haben Zugriff auf alle Verwaltungsfunktionen.",
      "demoteConfirm": "Möchten Sie diesem Benutzer wirklich die Admin-Rechte entziehen?\n\nDer Benutzer wird zu einem normalen Benutzer herabgestuft.",
      "promoteSuccess": "Benutzer wurde erfolgreich zum Admin ernannt.",
      "demoteSuccess": "Admin-Rechte wurden erfolgreich entzogen.",
      "promoteError": "Fehler beim Ernennen des Benutzers zum Admin.",
      "demoteError": "Fehler beim Entziehen der Admin-Rechte."
    }
  }
}
```

**Deliverables:**
- ✅ API client methods for promote/demote
- ✅ Updated admin users page with admin management UI
- ✅ German translations
- ✅ Success/error handling
- ✅ Conditional display based on Super Admin status

**// DONE - Phase 5 Complete**

---

## Phase 6: Protection & Security

**Objective:** Ensure Super Admin and admins cannot be auto-deactivated, deleted, or otherwise compromised.

### 6.1 Cron Job Protection

**File:** `internal/cron/cron.go`

**Tasks:**

1. Update `AutoDeactivateUsers()` to exclude admins
2. Update query to check `is_admin = FALSE` and `is_super_admin = FALSE`

**Implementation:**

```go
// internal/cron/cron.go

func (s *CronService) AutoDeactivateUsers() {
    settings, err := s.settingsRepo.GetAllSettings()
    if err != nil {
        log.Printf("Error getting system settings for auto-deactivation: %v", err)
        return
    }

    autoDeactivationDays := 365 // default
    if val, ok := settings["auto_deactivation_days"]; ok {
        if days, err := strconv.Atoi(val); err == nil {
            autoDeactivationDays = days
        }
    }

    cutoffDate := time.Now().AddDate(0, 0, -autoDeactivationDays)

    // NEW: Exclude admins and super admin from auto-deactivation
    query := `
        SELECT id, name, email, last_activity_at
        FROM users
        WHERE is_active = ?
          AND is_deleted = ?
          AND is_admin = ?       -- Exclude admins
          AND is_super_admin = ? -- Exclude super admin
          AND last_activity_at < ?
    `

    rows, err := s.db.Query(query, true, false, false, false, cutoffDate)
    if err != nil {
        log.Printf("Error querying users for auto-deactivation: %v", err)
        return
    }
    defer rows.Close()

    deactivatedCount := 0
    for rows.Next() {
        var user struct {
            ID             int
            Name           string
            Email          string
            LastActivityAt time.Time
        }

        err := rows.Scan(&user.ID, &user.Name, &user.Email, &user.LastActivityAt)
        if err != nil {
            log.Printf("Error scanning user row: %v", err)
            continue
        }

        // Deactivate user
        _, err = s.db.Exec(`
            UPDATE users
            SET is_active = ?,
                deactivated_at = ?,
                deactivation_reason = ?
            WHERE id = ?
        `, false, time.Now(), fmt.Sprintf("Auto-deactivated after %d days of inactivity", autoDeactivationDays), user.ID)

        if err != nil {
            log.Printf("Error deactivating user %d (%s): %v", user.ID, user.Name, err)
            continue
        }

        deactivatedCount++
        log.Printf("Auto-deactivated user: %s (%s) - Last activity: %s", user.Name, user.Email, user.LastActivityAt.Format("2006-01-02"))
    }

    if deactivatedCount > 0 {
        log.Printf("Auto-deactivation completed: %d users deactivated", deactivatedCount)
    }
}
```

### 6.2 User Deletion Protection

**File:** `internal/repository/user_repository.go`

**Tasks:**

1. Update `DeleteAccount()` to prevent deletion of Super Admin (ID = 1)
2. Optionally prevent deletion of all admins (business decision)

**Implementation:**

```go
// internal/repository/user_repository.go

func (r *UserRepository) DeleteAccount(userID int) error {
    // NEW: Prevent Super Admin deletion
    if userID == 1 {
        return errors.New("cannot delete Super Admin account")
    }

    // Optional: Prevent all admin deletions (uncomment if desired)
    /*
    var isAdmin bool
    err := r.db.QueryRow("SELECT is_admin FROM users WHERE id = ?", userID).Scan(&isAdmin)
    if err != nil {
        return err
    }
    if isAdmin {
        return errors.New("cannot delete admin accounts")
    }
    */

    // GDPR-compliant anonymization (existing logic)
    anonymousID := fmt.Sprintf("anonymous_user_%d", time.Now().Unix())
    query := `
        UPDATE users
        SET name = ?,
            email = NULL,
            phone = NULL,
            password_hash = NULL,
            is_deleted = ?,
            anonymous_id = ?,
            profile_photo = NULL
        WHERE id = ?
    `

    _, err := r.db.Exec(query, "Deleted User", true, anonymousID, userID)
    return err
}
```

### 6.3 Frontend Protection

**File:** `frontend/admin-users.html`

**Tasks:**

1. Hide deactivate button for admins and super admin (already done in Phase 5.2)
2. Add visual indicators (badges)

**Already implemented in Phase 5.2.**

### 6.4 .gitignore Update

**File:** `.gitignore`

**Tasks:**

1. Add `SUPER_ADMIN_CREDENTIALS.txt` to prevent accidental commits

**Implementation:**

```gitignore
# ... existing entries ...

# Super Admin credentials (NEVER commit!)
SUPER_ADMIN_CREDENTIALS.txt

# Database files
*.db
*.db-shm
*.db-wal

# ... rest of .gitignore ...
```

**Deliverables:**
- ✅ Cron job protection for admins
- ✅ Super Admin deletion protection
- ✅ Frontend UI protection
- ✅ .gitignore updated

**// DONE - Phase 6 Complete**

---

## Phase 7: Testing & Documentation

**Objective:** Thoroughly test all features and update documentation.

### 7.1 Testing Checklist

#### 7.1.1 Database & Migration Tests

- [ ] Fresh database migration creates `is_admin` and `is_super_admin` columns
- [ ] Indexes are created successfully
- [ ] Unique constraint on `is_super_admin` works (SQLite/PostgreSQL)
- [ ] Existing database migration works without errors
- [ ] Test on all three databases: SQLite, MySQL, PostgreSQL

#### 7.1.2 Seed Data Tests

- [ ] Seed runs automatically on empty database
- [ ] Super Admin created with correct flags (ID=1, is_admin=true, is_super_admin=true)
- [ ] 3 test users created with different levels
- [ ] 5 dogs created with different categories
- [ ] 3 bookings created (past, present, future)
- [ ] System settings initialized
- [ ] Credentials file created with correct format
- [ ] Credentials file has 600 permissions (Linux)
- [ ] Console output shows credentials clearly
- [ ] Seed does NOT run on database with existing users

#### 7.1.3 Super Admin Password Management Tests

- [ ] Password file read successfully on startup
- [ ] Unchanged password detected (no database update)
- [ ] Changed password detected and hashed
- [ ] Database updated with new password hash
- [ ] File rewritten with confirmation message
- [ ] Can login with new password after restart
- [ ] Email mismatch between file and .env detected

#### 7.1.4 Authentication Tests

- [ ] Login returns JWT with `is_admin` and `is_super_admin` claims
- [ ] Super Admin login includes `is_super_admin: true`
- [ ] Regular admin login includes `is_admin: true, is_super_admin: false`
- [ ] Regular user login includes `is_admin: false, is_super_admin: false`
- [ ] `/api/me` endpoint returns admin flags
- [ ] JWT validation extracts admin flags correctly
- [ ] Middleware injects admin flags into context

#### 7.1.5 Middleware Tests

- [ ] `RequireAdmin` blocks non-admin users (403)
- [ ] `RequireAdmin` allows admin users
- [ ] `RequireAdmin` allows super admin users
- [ ] `RequireSuperAdmin` blocks non-super-admin users (403)
- [ ] `RequireSuperAdmin` blocks regular admins (403)
- [ ] `RequireSuperAdmin` allows super admin users

#### 7.1.6 API Endpoint Tests

- [ ] `POST /api/admin/users/:id/promote` - Success case
- [ ] Promote endpoint requires Super Admin (403 for regular admin)
- [ ] Cannot promote user who is already admin (400)
- [ ] Cannot promote Super Admin (400)
- [ ] Database updated correctly after promotion
- [ ] `POST /api/admin/users/:id/demote` - Success case
- [ ] Demote endpoint requires Super Admin (403 for regular admin)
- [ ] Cannot demote user who is not admin (400)
- [ ] Cannot demote Super Admin (400)
- [ ] Database updated correctly after demotion

#### 7.1.7 Frontend Tests

- [ ] Admin users page shows role column
- [ ] Super Admin badge displays correctly
- [ ] Admin badge displays correctly
- [ ] Regular users show "Benutzer"
- [ ] Promote button visible only to Super Admin
- [ ] Promote button NOT visible on Super Admin's row
- [ ] Promote button NOT visible on admin rows
- [ ] Demote button visible only to Super Admin
- [ ] Demote button visible on admin rows (not super admin)
- [ ] Clicking promote shows confirmation dialog
- [ ] Successful promotion refreshes user list
- [ ] Successful promotion shows success message
- [ ] Error handling shows error message
- [ ] Demote operations work identically to promote

#### 7.1.8 Protection Tests

- [ ] Auto-deactivation cron skips admins
- [ ] Auto-deactivation cron skips super admin
- [ ] Deactivate button hidden for admins in UI
- [ ] Deactivate button hidden for super admin in UI
- [ ] Cannot delete Super Admin via API (ID=1 protected)
- [ ] `SUPER_ADMIN_CREDENTIALS.txt` in .gitignore

#### 7.1.9 Migration Tests (Existing Installations)

- [ ] Migration script works on existing database
- [ ] Existing users unchanged (is_admin=false by default)
- [ ] Manually set Super Admin works
- [ ] Old admins in ADMIN_EMAILS can be promoted via UI
- [ ] Old ADMIN_EMAILS config no longer affects permissions

### 7.2 Manual Testing Scenarios

#### Scenario 1: Fresh Installation

1. Delete `gassigeher.db` (or drop MySQL/PostgreSQL database)
2. Set `SUPER_ADMIN_EMAIL=admin@test.com` in `.env`
3. Start server
4. Verify console output shows credentials
5. Verify `SUPER_ADMIN_CREDENTIALS.txt` created
6. Login with Super Admin credentials
7. Verify admin users page accessible
8. Verify all 8 admin pages accessible
9. Verify "Super Admin" badge visible
10. Verify promote/demote buttons visible on test users
11. Delete test users (optional)

#### Scenario 2: Super Admin Password Change

1. Open `SUPER_ADMIN_CREDENTIALS.txt`
2. Change password to `NewPassword123!`
3. Save file
4. Restart server
5. Verify file updated with confirmation
6. Login with new password
7. Verify old password no longer works

#### Scenario 3: Promote User to Admin

1. Login as Super Admin
2. Create a new regular user via registration
3. Go to admin users page
4. Find the new user
5. Click "Zu Admin ernennen"
6. Confirm dialog
7. Verify success message
8. Verify user now shows "Admin" badge
9. Verify "Admin entfernen" button now visible
10. Logout and login as new admin
11. Verify access to admin pages
12. Verify promote/demote buttons NOT visible (not super admin)

#### Scenario 4: Demote Admin

1. Login as Super Admin
2. Go to admin users page
3. Find an admin user (not super admin)
4. Click "Admin entfernen"
5. Confirm dialog
6. Verify success message
7. Verify user now shows "Benutzer"
8. Verify "Zu Admin ernennen" button now visible
9. Logout and login as demoted user
10. Verify NO access to admin pages (403)

#### Scenario 5: Protection Tests

1. Login as regular admin (not super admin)
2. Go to admin users page
3. Verify promote/demote buttons NOT visible
4. Manually try `POST /api/admin/users/2/promote` via curl
5. Verify 403 Forbidden response
6. Login as Super Admin
7. Find your own row in user table
8. Verify NO promote/demote buttons on own row
9. Verify "Super Admin" badge visible
10. Try to demote self via API (should fail)

### 7.3 Documentation Updates

#### 7.3.1 README.md Updates

**File:** `README.md`

**Tasks:**

1. Update setup instructions to mention `SUPER_ADMIN_EMAIL`
2. Add note about credentials file
3. Update environment variables section

**Add to .env.example:**

```bash
# Super Admin Configuration (Required)
SUPER_ADMIN_EMAIL=admin@yourshelter.com

# REMOVED: ADMIN_EMAILS (no longer used, replaced by database flags)
```

#### 7.3.2 CLAUDE.md Updates

**File:** `CLAUDE.md`

**Tasks:**

1. Update authentication section
2. Update admin management section
3. Add Super Admin section
4. Update seed data section

**Add section:**

```markdown
### Super Admin System

**Super Admin vs Regular Admin:**

- **Super Admin** (ID=1): Can promote/demote other admins, cannot be deleted/deactivated
- **Regular Admin**: All admin functions except user promotion, can be demoted by Super Admin
- **Admin privileges stored in database** (not config file)

**First-time installation:**
- Automatic seed data generation
- Super Admin created automatically
- Credentials in `SUPER_ADMIN_CREDENTIALS.txt` and console

**Change Super Admin password:**
1. Edit `SUPER_ADMIN_CREDENTIALS.txt`
2. Restart server
3. File updated with confirmation

**Promote user to admin:**
- Login as Super Admin
- Go to admin-users.html
- Click "Zu Admin ernennen" button

**Authentication changes:**
- JWT includes `is_admin` and `is_super_admin` claims
- Middleware: `RequireAdmin` and `RequireSuperAdmin`
- Config-based `ADMIN_EMAILS` removed
```

#### 7.3.3 API.md Updates

**File:** `docs/API.md`

**Tasks:**

1. Add promote and demote endpoints
2. Update authentication response schemas

**Add endpoints:**

```markdown
### POST /api/admin/users/:id/promote

Promote user to admin role (Super Admin only).

**Authorization:** Bearer token (Super Admin)

**Response 200:**
```json
{
  "message": "User promoted to admin successfully",
  "user": {
    "id": 123,
    "name": "Anna Schmidt",
    "email": "anna@shelter.com",
    "is_admin": true,
    "is_super_admin": false
  }
}
```

**Response 403:** Forbidden (not Super Admin)
**Response 400:** User already admin, or trying to modify Super Admin

---

### POST /api/admin/users/:id/demote

Revoke admin privileges (Super Admin only).

**Authorization:** Bearer token (Super Admin)

**Response 200:**
```json
{
  "message": "Admin privileges revoked successfully",
  "user": {
    "id": 123,
    "name": "Anna Schmidt",
    "email": "anna@shelter.com",
    "is_admin": false,
    "is_super_admin": false
  }
}
```

**Response 403:** Forbidden (not Super Admin)
**Response 400:** User not admin, or trying to demote Super Admin
```

#### 7.3.4 ADMIN_GUIDE.md Updates

**File:** `docs/ADMIN_GUIDE.md`

**Tasks:**

1. Add Super Admin section
2. Add admin management guide
3. Update first-time setup section

**Add section:**

```markdown
## Admin Management (Super Admin Only)

Only the Super Admin can promote users to admin or revoke admin privileges.

### Promoting a User to Admin

1. Navigate to **Benutzerverwaltung** (admin-users.html)
2. Find the user you want to promote
3. Click **"Zu Admin ernennen"** button
4. Confirm the action
5. User now has full admin access to all features

### Revoking Admin Privileges

1. Navigate to **Benutzerverwaltung**
2. Find the admin you want to demote
3. Click **"Admin entfernen"** button
4. Confirm the action
5. User reverted to regular user (no admin access)

**Note:** You cannot demote the Super Admin. You cannot demote yourself.

### Changing Your Super Admin Password

1. Locate the file `SUPER_ADMIN_CREDENTIALS.txt` in the application directory
2. Open with a text editor
3. Change the line `PASSWORD: ...` to your new password
4. Save the file
5. Restart the Gassigeher server
6. File will be updated with confirmation
7. Login with new password
```

**Deliverables:**
- ✅ Comprehensive test checklist
- ✅ Manual testing scenarios
- ✅ Updated README.md
- ✅ Updated CLAUDE.md
- ✅ Updated API.md
- ✅ Updated ADMIN_GUIDE.md
- ✅ Updated .env.example

**// DONE - Phase 7 Complete**

---

## Implementation Checklist

Use this checklist to track implementation progress:

### Phase 1: Database Schema & Models
- [ ] Add `is_admin` column to users table (all DBs)
- [ ] Add `is_super_admin` column to users table (all DBs)
- [ ] Create indexes
- [ ] Create unique constraint (SQLite/PostgreSQL)
- [ ] Update User model struct
- [ ] Update `FindByID` repository method
- [ ] Update `FindByEmail` repository method
- [ ] Add `PromoteToAdmin` repository method
- [ ] Add `DemoteAdmin` repository method
- [ ] Test migration on all 3 databases

### Phase 2: Super Admin Service & Seed Data
- [ ] Create `internal/database/seed.go`
- [ ] Implement `SeedDatabase()` function
- [ ] Implement `generateSecurePassword()`
- [ ] Implement `generateTestUsers()`
- [ ] Implement `generateDogs()`
- [ ] Implement `generateBookings()`
- [ ] Implement `writeCredentialsFile()`
- [ ] Implement `printSetupComplete()`
- [ ] Create `internal/services/super_admin_service.go`
- [ ] Implement `CheckAndUpdatePassword()`
- [ ] Implement `parseCredentialsFile()`
- [ ] Implement `writeUpdatedCredentialsFile()`
- [ ] Update `internal/config/config.go` (add SuperAdminEmail)
- [ ] Remove `AdminEmails` from config
- [ ] Remove `IsAdmin()` method from config
- [ ] Integrate seed in `main.go`
- [ ] Integrate password check in `main.go`
- [ ] Test fresh installation
- [ ] Test password change flow

### Phase 3: Authentication Updates
- [ ] Add `IsAdmin` to JWT Claims struct
- [ ] Add `IsSuperAdmin` to JWT Claims struct
- [ ] Update `GenerateJWT()` method
- [ ] Add `IsSuperAdminKey` to middleware
- [ ] Update `AuthMiddleware` to inject super admin flag
- [ ] Create `RequireSuperAdmin` middleware
- [ ] Update login handler to return admin flags
- [ ] Verify `/api/me` returns admin flags
- [ ] Test JWT generation
- [ ] Test middleware protection

### Phase 4: Backend API Endpoints
- [ ] Implement `PromoteToAdmin()` handler
- [ ] Implement `DemoteAdmin()` handler
- [ ] Add validation logic (cannot promote/demote super admin)
- [ ] Register routes in `main.go`
- [ ] Apply `RequireSuperAdmin` middleware
- [ ] Test promote endpoint
- [ ] Test demote endpoint
- [ ] Test error cases (403, 400)

### Phase 5: Frontend UI Updates
- [ ] Add `promoteToAdmin()` to API client
- [ ] Add `demoteAdmin()` to API client
- [ ] Update admin-users.html table structure
- [ ] Add role column
- [ ] Add badge rendering
- [ ] Add promote button (conditional)
- [ ] Add demote button (conditional)
- [ ] Implement `promoteToAdmin()` JavaScript function
- [ ] Implement `demoteAdmin()` JavaScript function
- [ ] Add German translations
- [ ] Add CSS for badges and buttons
- [ ] Test UI visibility rules
- [ ] Test promote/demote operations

### Phase 6: Protection & Security
- [ ] Update cron auto-deactivation to exclude admins
- [ ] Add Super Admin deletion protection
- [ ] Verify UI hides deactivate button for admins
- [ ] Add `SUPER_ADMIN_CREDENTIALS.txt` to .gitignore
- [ ] Test cron protection
- [ ] Test deletion protection

### Phase 7: Testing & Documentation
- [ ] Run all database & migration tests
- [ ] Run all seed data tests
- [ ] Run all password management tests
- [ ] Run all authentication tests
- [ ] Run all middleware tests
- [ ] Run all API endpoint tests
- [ ] Run all frontend tests
- [ ] Run all protection tests
- [ ] Run all migration tests (existing installations)
- [ ] Complete Scenario 1: Fresh Installation
- [ ] Complete Scenario 2: Password Change
- [ ] Complete Scenario 3: Promote User
- [ ] Complete Scenario 4: Demote Admin
- [ ] Complete Scenario 5: Protection Tests
- [ ] Update README.md
- [ ] Update CLAUDE.md
- [ ] Update API.md
- [ ] Update ADMIN_GUIDE.md
- [ ] Update .env.example
- [ ] Create migration guide document

---

## Testing Strategy

### Unit Tests

**Files to create:**

1. `internal/database/seed_test.go` - Test seed data generation
2. `internal/services/super_admin_service_test.go` - Test password file management
3. `internal/repository/user_repository_test.go` - Add tests for promote/demote
4. `internal/handlers/user_handler_test.go` - Add tests for new endpoints

**Key test cases:**

```go
// Example: internal/database/seed_test.go

func TestSeedDatabase_EmptyDatabase(t *testing.T) {
    // Test seed runs on empty database
}

func TestSeedDatabase_ExistingData(t *testing.T) {
    // Test seed skips when data exists
}

func TestGenerateSecurePassword(t *testing.T) {
    // Test password contains required characters
}

// Example: internal/services/super_admin_service_test.go

func TestCheckAndUpdatePassword_PasswordChanged(t *testing.T) {
    // Test password update when changed
}

func TestCheckAndUpdatePassword_PasswordUnchanged(t *testing.T) {
    // Test no update when password same
}

func TestParseCredentialsFile(t *testing.T) {
    // Test file parsing
}
```

### Integration Tests

**Files to create:**

1. `internal/handlers/admin_integration_test.go` - Test promote/demote workflows

**Key test cases:**

```go
func TestPromoteToAdmin_Integration(t *testing.T) {
    // 1. Create test database
    // 2. Create super admin
    // 3. Create regular user
    // 4. Call promote endpoint
    // 5. Verify database updated
    // 6. Verify JWT includes admin flag
}

func TestAdminProtection_Integration(t *testing.T) {
    // 1. Create admin user
    // 2. Run auto-deactivation cron
    // 3. Verify admin still active
}
```

### End-to-End Tests

**Manual test script:**

1. Delete database
2. Start server
3. Login with Super Admin credentials
4. Promote user to admin
5. Logout and login as new admin
6. Verify admin access
7. Logout and login as Super Admin
8. Demote admin
9. Logout and login as demoted user
10. Verify no admin access

---

## Rollout Plan

### Pre-Deployment Checklist

- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] Manual testing completed
- [ ] Documentation updated
- [ ] Migration guide created
- [ ] Backup procedures documented
- [ ] Rollback plan prepared

### Deployment Steps (New Installation)

1. **Prepare environment:**
   - Set `SUPER_ADMIN_EMAIL` in `.env`
   - Remove `ADMIN_EMAILS` from `.env`

2. **Deploy code:**
   - `git pull origin master`
   - `go build -o gassigeher ./cmd/server`

3. **Start server:**
   - `./gassigeher`
   - Wait for seed data generation
   - Save Super Admin credentials from console or file

4. **Verify installation:**
   - Login as Super Admin
   - Access admin pages
   - Test promote/demote functionality

### Deployment Steps (Existing Installation)

1. **Backup database:**
   ```bash
   cp gassigeher.db gassigeher.db.backup  # SQLite
   mysqldump -u user -p gassigeher > backup.sql  # MySQL
   pg_dump gassigeher > backup.sql  # PostgreSQL
   ```

2. **Update code:**
   ```bash
   git pull origin master
   go build -o gassigeher ./cmd/server
   ```

3. **Update configuration:**
   - Add `SUPER_ADMIN_EMAIL` to `.env`
   - Remove `ADMIN_EMAILS` from `.env`

4. **Set Super Admin manually:**
   ```sql
   UPDATE users SET is_admin = 1, is_super_admin = 1 WHERE email = 'your-admin@shelter.com';
   ```

5. **Create credentials file:**
   ```bash
   # Create file with current password
   nano SUPER_ADMIN_CREDENTIALS.txt
   # Add content from InstallationAndSelfServices.md template
   chmod 600 SUPER_ADMIN_CREDENTIALS.txt
   ```

6. **Restart server:**
   ```bash
   systemctl restart gassigeher  # Linux
   # OR
   ./gassigeher  # Direct execution
   ```

7. **Verify migration:**
   - Login as Super Admin
   - Check for "Super Admin" badge
   - Test promote/demote buttons
   - Promote other admins (if needed)

### Rollback Plan

If issues occur during deployment:

1. **Stop server:**
   ```bash
   systemctl stop gassigeher
   ```

2. **Restore database:**
   ```bash
   cp gassigeher.db.backup gassigeher.db  # SQLite
   mysql -u user -p gassigeher < backup.sql  # MySQL
   psql gassigeher < backup.sql  # PostgreSQL
   ```

3. **Revert code:**
   ```bash
   git checkout [previous-commit-hash]
   go build -o gassigeher ./cmd/server
   ```

4. **Restore configuration:**
   - Restore `ADMIN_EMAILS` in `.env`
   - Remove `SUPER_ADMIN_EMAIL`

5. **Start server:**
   ```bash
   systemctl start gassigeher
   ```

---

## Summary

This implementation plan provides a comprehensive, phase-by-phase approach to implementing the Installation and Self-Services features:

**Phase 1:** Database schema changes and model updates
**Phase 2:** Seed data generation and Super Admin password file management
**Phase 3:** Authentication system updates (JWT, middleware)
**Phase 4:** Backend API endpoints for promote/demote
**Phase 5:** Frontend UI updates for admin management
**Phase 6:** Protection and security measures
**Phase 7:** Testing and documentation

**Estimated Effort:**
- Phase 1: 2-3 hours
- Phase 2: 4-5 hours
- Phase 3: 2-3 hours
- Phase 4: 2-3 hours
- Phase 5: 3-4 hours
- Phase 6: 1-2 hours
- Phase 7: 4-6 hours

**Total: 18-26 hours**

**Dependencies:**
- Phases must be completed in order
- Each phase builds on previous phases
- Testing should be continuous throughout

**Success Criteria:**
- ✅ Fresh installation automatically creates Super Admin
- ✅ Super Admin can promote/demote other admins via UI
- ✅ File-based password management works
- ✅ Admins protected from auto-deactivation
- ✅ All tests passing
- ✅ Documentation complete

---

**Document Version:** 1.0
**Last Updated:** 2025-01-23
**Status:** Ready for Implementation
