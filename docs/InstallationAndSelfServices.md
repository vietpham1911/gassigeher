# Installation and Self-Service Guide

## Overview

This document describes the automatic installation system, Super Admin management, and self-service operations that shelter staff can perform without developer assistance.

**Design Philosophy**: Keep it simple. All operations via Web UI. Super Admin password management via file-based system (no complex recovery flows).

---

## Table of Contents

1. [First-Time Installation](#first-time-installation)
2. [Super Admin System](#super-admin-system)
3. [Admin Management](#admin-management)
4. [Database Schema Changes](#database-schema-changes)
5. [Implementation Details](#implementation-details)
6. [Migration Guide (Existing Installations)](#migration-guide-existing-installations)
7. [Troubleshooting](#troubleshooting)

---

## First-Time Installation

### Automatic Seed Data Generation

When the application starts and detects an empty `users` table, it **automatically** generates seed data:

**Seed Data Created:**

1. **1 Super Admin** (ID: 1)
   - Email from `.env` (`SUPER_ADMIN_EMAIL`)
   - Auto-generated secure password (20 characters, alphanumeric + symbols)
   - Flags: `is_super_admin = TRUE`, `is_admin = TRUE`, `is_active = TRUE`
   - Experience level: Orange (full access)
   - Email verified: TRUE (bypass verification)

2. **3 Normal Users** (for testing)
   - User 1: Green level walker
   - User 2: Blue level walker
   - User 3: Orange level walker
   - All with auto-generated passwords
   - All active and verified

3. **5 Dogs** (mix of categories)
   - 2 Green category dogs (easy)
   - 2 Blue category dogs (medium)
   - 1 Orange category dog (challenging)
   - All available (not disabled)

4. **3 Bookings** (sample data)
   - 1 past booking (completed)
   - 1 present/today booking (scheduled)
   - 1 future booking (scheduled)
   - Distributed across different users and dogs

5. **Default System Settings**
   - `booking_advance_days = 14`
   - `cancellation_notice_hours = 12`
   - `auto_deactivation_days = 365`

**Credentials Output:**

All credentials are written to **two locations**:

1. **Console Output** (visible during server startup):
```
=============================================================
  GASSIGEHER - INSTALLATION COMPLETE
=============================================================

SUPER ADMIN CREDENTIALS (SAVE THESE!):
  Email: superadmin@tierheim-goeppingen.de
  Password: Xy9$mK2#pL5@qR8*tN3!

TEST USER CREDENTIALS:
  1. green-walker@test.com / aB3$cD7*eF2
  2. blue-walker@test.com / gH9#iJ4@kL6
  3. orange-walker@test.com / mN2$oP8*qR5

IMPORTANT:
- Super Admin password saved to: SUPER_ADMIN_CREDENTIALS.txt
- Change Super Admin password: Edit file and restart server
- Test users can be deleted after setup

=============================================================
```

2. **File: `SUPER_ADMIN_CREDENTIALS.txt`** (in application root):
```
=============================================================
GASSIGEHER - SUPER ADMIN CREDENTIALS
=============================================================

EMAIL: superadmin@tierheim-goeppingen.de
PASSWORD: Xy9$mK2#pL5@qR8*tN3!

CREATED: 2025-01-15 14:32:18
LAST UPDATED: 2025-01-15 14:32:18

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
```

### Environment Variable Required

Add to `.env` file **before first startup**:

```bash
# Super Admin Configuration (Required)
SUPER_ADMIN_EMAIL=superadmin@tierheim-goeppingen.de

# Existing variables...
JWT_SECRET=your-secret-key
DATABASE_PATH=./gassigeher.db
# ... etc
```

**Notes:**
- `SUPER_ADMIN_EMAIL` is **required** (server exits with error if missing)
- Email is **fixed** and cannot be changed via UI (prevents account takeover)
- Only changeable by editing `.env` and restarting (requires server access)

---

## Super Admin System

### What is Super Admin?

The **Super Admin** is the ultimate administrator with special privileges:

1. **Always exists** (created automatically on first startup)
2. **Cannot be deleted** (ID 1 is protected)
3. **Cannot be deactivated** (immune to auto-deactivation cron job)
4. **Never expires** (exempt from inactivity rules)
5. **Can manage other admins** (promote/demote privileges)
6. **Full system access** (all admin features + admin management)

### Super Admin Password Management

**Password Change Process (File-Based):**

1. **Locate the file**: `SUPER_ADMIN_CREDENTIALS.txt` in application root
2. **Open and edit**: Change the `PASSWORD:` line to your new password
   ```
   PASSWORD: MyNewSecurePassword123!
   ```
3. **Save the file**
4. **Restart the server**: `systemctl restart gassigeher` (Linux) or restart service
5. **Verify**: File will be updated with timestamp confirming password change

**What Happens on Server Startup:**

```go
// Pseudo-code flow
1. Read SUPER_ADMIN_CREDENTIALS.txt
2. Parse EMAIL and PASSWORD fields
3. Check if PASSWORD has changed (compare hash)
4. If changed:
   - Hash new password (bcrypt)
   - Update database: UPDATE users SET password_hash = ? WHERE id = 1
   - Rewrite file with confirmation and new timestamp
5. If not changed:
   - No action needed
```

**File Format After Password Change:**

```
=============================================================
GASSIGEHER - SUPER ADMIN CREDENTIALS
=============================================================

EMAIL: superadmin@tierheim-goeppingen.de
PASSWORD: MyNewSecurePassword123!

CREATED: 2025-01-15 14:32:18
LAST UPDATED: 2025-01-20 09:15:42  ← Updated!

PASSWORD CHANGE CONFIRMED: ✓

=============================================================
```

**Password Recovery:**

If Super Admin forgets password:
1. Open `SUPER_ADMIN_CREDENTIALS.txt`
2. Read current password from `PASSWORD:` line
3. If file lost: Restore from backup or contact developer

**Security Notes:**

- File has restricted permissions: `chmod 600 SUPER_ADMIN_CREDENTIALS.txt` (Linux)
- File is in `.gitignore` (never committed to version control)
- Plain text password in file is acceptable (file system security assumed)
- Alternative complex solutions (CLI recovery, emergency tokens) rejected for simplicity

---

## Admin Management

### Admin vs Super Admin

**Super Admin Can:**
- ✅ Promote normal users to admin
- ✅ Revoke admin privileges from admins
- ✅ View all users in admin panel
- ✅ Cannot be deactivated or deleted
- ✅ All regular admin functions

**Regular Admin Can:**
- ✅ Manage dogs (create, edit, disable)
- ✅ View all bookings
- ✅ Manage blocked dates
- ✅ Review experience requests
- ✅ View dashboard statistics
- ✅ Manage system settings
- ❌ Cannot promote/demote other admins
- ❌ Can be demoted by Super Admin
- ⚠️ Cannot be auto-deactivated (but Super Admin can manually deactivate)

**Normal User:**
- Regular booking and profile management
- No access to admin pages

### Admin Management UI

**Location:** `admin-users.html` (existing user management page)

**New Features (Only Visible to Super Admin):**

In the user table, each row will have additional buttons:

```html
<!-- Existing buttons -->
<button class="btn btn-view">View</button>
<button class="btn btn-deactivate">Deactivate</button>

<!-- NEW: Only shown to Super Admin, not on super admin's own row -->
<button class="btn btn-promote" data-user-id="123">
  Promote to Admin
</button>
<!-- OR -->
<button class="btn btn-demote" data-user-id="123">
  Revoke Admin
</button>
```

**Button Logic:**

```javascript
// Show "Promote" if:
- Current user is Super Admin (is_super_admin = true)
- Target user is NOT admin (is_admin = false)
- Target user is NOT the Super Admin (id !== 1)

// Show "Demote" if:
- Current user is Super Admin (is_super_admin = true)
- Target user IS admin (is_admin = true)
- Target user is NOT Super Admin (is_super_admin = false)

// Super Admin's own row:
- Shows badge: "Super Admin"
- No promote/demote buttons (cannot modify self)
```

**User Table Display:**

| Name | Email | Level | Status | Role | Actions |
|------|-------|-------|--------|------|---------|
| Max Mustermann | super@shelter.com | Orange | Active | **Super Admin** | View |
| Anna Schmidt | anna@shelter.com | Blue | Active | **Admin** | View, **Revoke Admin** |
| Peter Klein | peter@shelter.com | Green | Active | User | View, Deactivate, **Promote to Admin** |

### New API Endpoints

**1. Promote User to Admin**

```http
POST /api/admin/users/:id/promote
Authorization: Bearer <super_admin_token>

Response 200:
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

Response 403:
{
  "error": "Only Super Admin can promote users"
}

Response 400:
{
  "error": "User is already an admin"
}
```

**2. Revoke Admin Privileges**

```http
POST /api/admin/users/:id/demote
Authorization: Bearer <super_admin_token>

Response 200:
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

Response 403:
{
  "error": "Only Super Admin can demote admins"
}

Response 400:
{
  "error": "Cannot demote Super Admin"
}
```

**Middleware Protection:**

```go
// New middleware: internal/middleware/middleware.go
func RequireSuperAdmin(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        isSuperAdmin, ok := r.Context().Value(IsSuperAdminKey).(bool)
        if !ok || !isSuperAdmin {
            http.Error(w, "Forbidden: Super Admin required", http.StatusForbidden)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

**Route Registration:**

```go
// cmd/server/main.go
superAdmin := admin.With(middleware.RequireSuperAdmin)
superAdmin.HandleFunc("/api/admin/users/{id}/promote", userHandler.PromoteToAdmin).Methods("POST")
superAdmin.HandleFunc("/api/admin/users/{id}/demote", userHandler.DemoteAdmin).Methods("POST")
```

---

## Database Schema Changes

### Users Table Modifications

**Add Two New Columns:**

```sql
-- Migration: Add admin flags
ALTER TABLE users ADD COLUMN is_admin BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN is_super_admin BOOLEAN DEFAULT FALSE;

-- Create index for faster admin queries
CREATE INDEX idx_users_admin ON users(is_admin);
CREATE INDEX idx_users_super_admin ON users(is_super_admin);

-- Ensure only one Super Admin exists (database constraint)
CREATE UNIQUE INDEX idx_one_super_admin ON users(is_super_admin) WHERE is_super_admin = TRUE;
```

**Updated Users Table Schema:**

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    password_hash TEXT,
    phone TEXT,
    experience_level TEXT DEFAULT 'green',
    profile_photo TEXT,

    -- Admin flags (NEW)
    is_admin BOOLEAN DEFAULT FALSE,
    is_super_admin BOOLEAN DEFAULT FALSE,

    -- Account status
    is_active BOOLEAN DEFAULT TRUE,
    is_verified BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,

    -- Tracking
    last_activity_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deactivated_at TIMESTAMP,
    deactivation_reason TEXT,
    anonymous_id TEXT,

    -- Verification
    verification_token TEXT,
    verification_token_expiry TIMESTAMP,
    reset_token TEXT,
    reset_token_expiry TIMESTAMP
);
```

**Database Constraints:**

1. **Unique Super Admin**: Only one user can have `is_super_admin = TRUE`
2. **Super Admin is Always Admin**: If `is_super_admin = TRUE`, then `is_admin` MUST be `TRUE`
3. **Super Admin ID**: Super Admin always has `id = 1` (by convention, not enforced)

### Migration Strategy

**New Migration File:** `internal/database/migrations.go`

```go
func RunMigrations(db *sql.DB) error {
    // ... existing migrations ...

    // Migration: Add admin flags
    _, err := db.Exec(`
        ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;
    `)
    if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
        return err
    }

    _, err = db.Exec(`
        ALTER TABLE users ADD COLUMN IF NOT EXISTS is_super_admin BOOLEAN DEFAULT FALSE;
    `)
    if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
        return err
    }

    // Create indexes
    db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_admin ON users(is_admin)`)
    db.Exec(`CREATE INDEX IF NOT EXISTS idx_users_super_admin ON users(is_super_admin)`)

    return nil
}
```

---

## Implementation Details

### Seed Data Generation

**Location:** `internal/database/seed.go` (new file)

**Function Signature:**

```go
func SeedDatabase(db *sql.DB, cfg *config.Config) error {
    // 1. Check if users table is empty
    var count int
    err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    if err != nil {
        return err
    }

    if count > 0 {
        // Database already seeded, skip
        return nil
    }

    // 2. Generate Super Admin
    superAdminEmail := cfg.SuperAdminEmail
    if superAdminEmail == "" {
        return errors.New("SUPER_ADMIN_EMAIL not set in .env")
    }

    superAdminPassword := generateSecurePassword(20) // Random 20 chars
    hashedPassword := hashPassword(superAdminPassword)

    _, err = db.Exec(`
        INSERT INTO users (
            id, name, email, password_hash, experience_level,
            is_admin, is_super_admin, is_active, is_verified,
            last_activity_at, created_at
        ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `, 1, "Super Admin", superAdminEmail, hashedPassword, "orange",
       true, true, true, true, time.Now(), time.Now())

    if err != nil {
        return err
    }

    // 3. Write credentials to file
    writeCredentialsFile(superAdminEmail, superAdminPassword)

    // 4. Generate test users (3 users)
    testUsers := generateTestUsers(db)

    // 5. Generate dogs (5 dogs)
    generateDogs(db)

    // 6. Generate bookings (3 bookings)
    generateBookings(db)

    // 7. Print credentials to console
    printSetupComplete(superAdminEmail, superAdminPassword, testUsers)

    return nil
}
```

**Helper Functions:**

```go
func generateSecurePassword(length int) string {
    // Chars: a-z, A-Z, 0-9, special chars
    chars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
    password := make([]byte, length)
    for i := range password {
        password[i] = chars[rand.Intn(len(chars))]
    }
    return string(password)
}

func writeCredentialsFile(email, password string) error {
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
`, email, password, time.Now().Format("2006-01-02 15:04:05"),
   time.Now().Format("2006-01-02 15:04:05"))

    return os.WriteFile("SUPER_ADMIN_CREDENTIALS.txt", []byte(content), 0600)
}

func generateTestUsers(db *sql.DB) []TestUser {
    users := []TestUser{
        {Name: "Test Walker (Green)", Email: "green-walker@test.com", Level: "green"},
        {Name: "Test Walker (Blue)", Email: "blue-walker@test.com", Level: "blue"},
        {Name: "Test Walker (Orange)", Email: "orange-walker@test.com", Level: "orange"},
    }

    for i := range users {
        users[i].Password = generateSecurePassword(12)
        hashedPassword := hashPassword(users[i].Password)

        db.Exec(`
            INSERT INTO users (name, email, password_hash, experience_level,
                             is_active, is_verified, last_activity_at, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?)
        `, users[i].Name, users[i].Email, hashedPassword, users[i].Level,
           true, true, time.Now(), time.Now())
    }

    return users
}

func generateDogs(db *sql.DB) {
    dogs := []struct {
        Name     string
        Category string
        Breed    string
    }{
        {"Bella", "green", "Labrador Retriever"},
        {"Max", "green", "Golden Retriever"},
        {"Luna", "blue", "German Shepherd"},
        {"Charlie", "blue", "Border Collie"},
        {"Rocky", "orange", "Belgian Malinois"},
    }

    for _, dog := range dogs {
        db.Exec(`
            INSERT INTO dogs (name, category, breed, age, gender,
                            description, special_needs, is_available, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
        `, dog.Name, dog.Category, dog.Breed, 3, "male",
           fmt.Sprintf("Friendly %s looking for walks!", dog.Name),
           "None", true, time.Now())
    }
}

func generateBookings(db *sql.DB) {
    today := time.Now()
    yesterday := today.AddDate(0, 0, -1)
    tomorrow := today.AddDate(0, 0, 1)

    bookings := []struct {
        UserID int
        DogID  int
        Date   time.Time
        Time   string
        Status string
    }{
        {2, 1, yesterday, "09:00", "completed"},  // Past booking
        {3, 2, today, "14:00", "scheduled"},       // Today booking
        {4, 3, tomorrow, "10:30", "scheduled"},    // Future booking
    }

    for _, booking := range bookings {
        db.Exec(`
            INSERT INTO bookings (user_id, dog_id, date, scheduled_time,
                                walk_type, status, created_at)
            VALUES (?, ?, ?, ?, ?, ?, ?)
        `, booking.UserID, booking.DogID,
           booking.Date.Format("2006-01-02"), booking.Time,
           "short", booking.Status, time.Now())
    }
}
```

**Calling Seed Function:**

```go
// cmd/server/main.go
func main() {
    // ... config and database setup ...

    db, err := database.Connect(cfg)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Run migrations
    err = database.RunMigrations(db)
    if err != nil {
        log.Fatal("Failed to run migrations:", err)
    }

    // Run seed data (NEW)
    err = database.SeedDatabase(db, cfg)
    if err != nil {
        log.Fatal("Failed to seed database:", err)
    }

    // ... rest of server setup ...
}
```

### Super Admin Password File Management

**Location:** `internal/services/super_admin_service.go` (new file)

```go
package services

type SuperAdminService struct {
    db  *sql.DB
    cfg *config.Config
}

func NewSuperAdminService(db *sql.DB, cfg *config.Config) *SuperAdminService {
    return &SuperAdminService{db: db, cfg: cfg}
}

// CheckAndUpdatePassword reads credentials file and updates password if changed
func (s *SuperAdminService) CheckAndUpdatePassword() error {
    filePath := "SUPER_ADMIN_CREDENTIALS.txt"

    // Check if file exists
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        // File doesn't exist, this is okay (might be first run)
        return nil
    }

    // Read file
    content, err := os.ReadFile(filePath)
    if err != nil {
        return fmt.Errorf("failed to read credentials file: %w", err)
    }

    // Parse EMAIL and PASSWORD lines
    email, password, err := parseCredentialsFile(string(content))
    if err != nil {
        return err
    }

    // Verify email matches config
    if email != s.cfg.SuperAdminEmail {
        return fmt.Errorf("email in credentials file doesn't match SUPER_ADMIN_EMAIL in .env")
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
        return nil
    }

    // Password changed! Hash new password and update database
    newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    _, err = s.db.Exec("UPDATE users SET password_hash = ? WHERE id = 1", newHash)
    if err != nil {
        return err
    }

    // Update file with confirmation and new timestamp
    err = s.writeUpdatedCredentialsFile(email, password, true)
    if err != nil {
        return err
    }

    log.Println("✓ Super Admin password updated successfully")

    return nil
}

func parseCredentialsFile(content string) (email, password string, err error) {
    lines := strings.Split(content, "\n")
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "EMAIL:") {
            email = strings.TrimSpace(strings.TrimPrefix(line, "EMAIL:"))
        }
        if strings.HasPrefix(line, "PASSWORD:") {
            password = strings.TrimSpace(strings.TrimPrefix(line, "PASSWORD:"))
        }
    }

    if email == "" || password == "" {
        return "", "", errors.New("invalid credentials file format")
    }

    return email, password, nil
}

func (s *SuperAdminService) writeUpdatedCredentialsFile(email, password string, changed bool) error {
    changeConfirmation := ""
    if changed {
        changeConfirmation = "\nPASSWORD CHANGE CONFIRMED: ✓\n"
    }

    content := fmt.Sprintf(`=============================================================
GASSIGEHER - SUPER ADMIN CREDENTIALS
=============================================================

EMAIL: %s
PASSWORD: %s

CREATED: [Original timestamp preserved]
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
`, email, password, time.Now().Format("2006-01-02 15:04:05"), changeConfirmation)

    return os.WriteFile("SUPER_ADMIN_CREDENTIALS.txt", []byte(content), 0600)
}
```

**Integration in main.go:**

```go
// cmd/server/main.go
func main() {
    // ... after database setup and migrations ...

    // Check and update Super Admin password (NEW)
    superAdminService := services.NewSuperAdminService(db, cfg)
    err = superAdminService.CheckAndUpdatePassword()
    if err != nil {
        log.Printf("Warning: Failed to check Super Admin password: %v", err)
        // Don't exit - allow server to start
    }

    // ... rest of server setup ...
}
```

### Authentication Changes

**JWT Claims Update:**

```go
// internal/services/auth_service.go

type Claims struct {
    UserID       int    `json:"user_id"`
    Email        string `json:"email"`
    IsAdmin      bool   `json:"is_admin"`       // NEW: From database
    IsSuperAdmin bool   `json:"is_super_admin"` // NEW: From database
    jwt.RegisteredClaims
}

func (s *AuthService) GenerateJWT(user *models.User) (string, error) {
    claims := &Claims{
        UserID:       user.ID,
        Email:        user.Email,
        IsAdmin:      user.IsAdmin,      // NEW
        IsSuperAdmin: user.IsSuperAdmin, // NEW
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(s.jwtSecret))
}
```

**Middleware Update:**

```go
// internal/middleware/middleware.go

const (
    UserIDKey       contextKey = "userID"
    EmailKey        contextKey = "email"
    IsAdminKey      contextKey = "isAdmin"
    IsSuperAdminKey contextKey = "isSuperAdmin" // NEW
)

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ... token extraction and validation ...

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

**Remove Config-Based Admin Check:**

```go
// REMOVE THIS FUNCTION from internal/config/config.go:
// func (c *Config) IsAdmin(email string) bool { ... }

// REMOVE ADMIN_EMAILS from .env.example
// REMOVE references to config.IsAdmin() in all handlers
```

### Deactivation Protection

**Cron Job Update:**

```go
// internal/cron/cron.go

func (s *CronService) AutoDeactivateUsers() {
    // ... existing logic ...

    // Build query to exclude admins and super admin
    query := `
        SELECT id, name, email, last_activity_at
        FROM users
        WHERE is_active = ?
          AND is_deleted = ?
          AND is_admin = ?       -- Exclude all admins
          AND is_super_admin = ? -- Exclude super admin
          AND last_activity_at < ?
    `

    cutoffDate := time.Now().AddDate(0, 0, -settings.AutoDeactivationDays)

    rows, err := s.db.Query(query, true, false, false, false, cutoffDate)
    // ... rest of logic ...
}
```

**UI Protection:**

```javascript
// frontend/admin-users.html

function renderUserRow(user) {
    // Don't show deactivate button for admins or super admin
    const canDeactivate = !user.is_admin && !user.is_super_admin && user.is_active;

    return `
        <tr>
            <td>${user.name}</td>
            <td>${user.email}</td>
            <td>${user.experience_level}</td>
            <td>
                ${user.is_super_admin ? '<span class="badge badge-super-admin">Super Admin</span>' : ''}
                ${user.is_admin ? '<span class="badge badge-admin">Admin</span>' : ''}
                ${user.is_active ? 'Active' : 'Inactive'}
            </td>
            <td>
                <button class="btn btn-view" onclick="viewUser(${user.id})">View</button>
                ${canDeactivate ? `<button class="btn btn-deactivate" onclick="deactivateUser(${user.id})">Deactivate</button>` : ''}
                ${renderAdminButtons(user)}
            </td>
        </tr>
    `;
}

function renderAdminButtons(user) {
    // Only show to Super Admin
    if (!window.currentUser.is_super_admin) {
        return '';
    }

    // Don't show on Super Admin's own row
    if (user.is_super_admin) {
        return '';
    }

    if (user.is_admin) {
        return `<button class="btn btn-demote" onclick="demoteAdmin(${user.id})">Revoke Admin</button>`;
    } else {
        return `<button class="btn btn-promote" onclick="promoteToAdmin(${user.id})">Promote to Admin</button>`;
    }
}
```

---

## Migration Guide (Existing Installations)

### For Existing Gassigeher Deployments

If you have an existing Gassigeher installation with users and data, follow these steps:

**Step 1: Backup Database**

```bash
# SQLite
cp gassigeher.db gassigeher.db.backup

# MySQL
mysqldump -u user -p gassigeher > gassigeher_backup.sql

# PostgreSQL
pg_dump gassigeher > gassigeher_backup.sql
```

**Step 2: Update Code**

```bash
git pull origin master
go build -o gassigeher ./cmd/server
```

**Step 3: Update .env Configuration**

```bash
# Add new required variable
SUPER_ADMIN_EMAIL=your-admin@shelter.com

# REMOVE old variable (no longer used)
# ADMIN_EMAILS=admin1@shelter.com,admin2@shelter.com
```

**Step 4: Run Database Migration**

The migration will run automatically on server startup. It will:
- Add `is_admin` column to users table
- Add `is_super_admin` column to users table
- Create indexes

**Step 5: Set Super Admin Manually**

Since you already have users, you need to manually designate one as Super Admin:

```bash
# SQLite
sqlite3 gassigeher.db
UPDATE users SET is_admin = 1, is_super_admin = 1 WHERE email = 'your-admin@shelter.com';
.quit

# MySQL
mysql -u user -p gassigeher
UPDATE users SET is_admin = TRUE, is_super_admin = TRUE WHERE email = 'your-admin@shelter.com';
exit

# PostgreSQL
psql gassigeher
UPDATE users SET is_admin = TRUE, is_super_admin = TRUE WHERE email = 'your-admin@shelter.com';
\q
```

**Step 6: Create Super Admin Credentials File**

Since existing Super Admin already has a password, create file manually:

```bash
cat > SUPER_ADMIN_CREDENTIALS.txt << 'EOF'
=============================================================
GASSIGEHER - SUPER ADMIN CREDENTIALS
=============================================================

EMAIL: your-admin@shelter.com
PASSWORD: [Use your existing password]

CREATED: [Migration date]
LAST UPDATED: [Migration date]

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
EOF

chmod 600 SUPER_ADMIN_CREDENTIALS.txt
```

**Step 7: Promote Existing Admins (Optional)**

If you had multiple admins in `ADMIN_EMAILS`, promote them manually:

```sql
-- Promote other admins (but not super admin)
UPDATE users SET is_admin = 1, is_super_admin = 0
WHERE email IN ('admin2@shelter.com', 'admin3@shelter.com');
```

OR use the new UI after starting the server (Super Admin can promote via admin-users.html).

**Step 8: Restart Server**

```bash
systemctl restart gassigeher  # Linux with systemd
# OR
./gassigeher  # Direct execution
```

**Step 9: Verify Migration**

1. Login as Super Admin
2. Go to `admin-users.html`
3. Verify you see "Super Admin" badge
4. Verify "Promote to Admin" / "Revoke Admin" buttons appear on other users
5. Test promoting a user to admin
6. Test revoking admin privileges

---

## Troubleshooting

### Issue: Server Won't Start - "SUPER_ADMIN_EMAIL not set"

**Cause:** Missing required environment variable.

**Solution:**
```bash
# Add to .env file
SUPER_ADMIN_EMAIL=your-admin@shelter.com

# Restart server
systemctl restart gassigeher
```

### Issue: Can't Login as Super Admin - "Invalid credentials"

**Cause:** Password changed in file but server not restarted.

**Solution:**
```bash
# Restart server to apply password change
systemctl restart gassigeher

# Verify file was updated with confirmation message
cat SUPER_ADMIN_CREDENTIALS.txt
# Look for: PASSWORD CHANGE CONFIRMED: ✓
```

### Issue: Lost Super Admin Password

**Cause:** `SUPER_ADMIN_CREDENTIALS.txt` deleted or lost.

**Solution:**
```bash
# Method 1: Restore from backup
cp SUPER_ADMIN_CREDENTIALS.txt.backup SUPER_ADMIN_CREDENTIALS.txt

# Method 2: Manually reset password
# Create new credentials file with new password
echo "EMAIL: your-admin@shelter.com" > SUPER_ADMIN_CREDENTIALS.txt
echo "PASSWORD: NewSecurePassword123!" >> SUPER_ADMIN_CREDENTIALS.txt
chmod 600 SUPER_ADMIN_CREDENTIALS.txt

# Restart server - it will hash and save new password
systemctl restart gassigeher
```

### Issue: "Promote to Admin" Button Not Visible

**Cause:** Not logged in as Super Admin, or browser cache.

**Solution:**
```bash
# Verify you are Super Admin in database
sqlite3 gassigeher.db "SELECT id, email, is_admin, is_super_admin FROM users WHERE id = 1;"

# Should show:
# 1|your-admin@shelter.com|1|1

# Clear browser cache and re-login
# Hard refresh: Ctrl + Shift + R (Windows/Linux) or Cmd + Shift + R (Mac)
```

### Issue: Cannot Demote Super Admin

**Cause:** By design - Super Admin cannot be demoted.

**Solution:**
This is intentional. To change Super Admin:
1. Manually update database to designate new Super Admin
2. Update `.env` with new email
3. Demote old Super Admin manually via SQL

```sql
-- Promote new Super Admin (must be ID 1)
UPDATE users SET is_super_admin = 1, is_admin = 1 WHERE email = 'new-admin@shelter.com';

-- Demote old Super Admin
UPDATE users SET is_super_admin = 0 WHERE email = 'old-admin@shelter.com';
-- Note: Keep is_admin = 1 if they should remain admin
```

### Issue: Auto-Deactivation Still Affecting Admins

**Cause:** Migration didn't update cron job, or code not updated.

**Solution:**
```bash
# Verify code is updated
git pull origin master
go build -o gassigeher ./cmd/server

# Check cron job excludes admins
grep "is_admin" internal/cron/cron.go
# Should see: WHERE is_admin = ? AND is_super_admin = ?

# Restart server
systemctl restart gassigeher
```

### Issue: Seed Data Not Generated on Fresh Install

**Cause:** Users table not empty, or seed function not running.

**Solution:**
```bash
# Check if users exist
sqlite3 gassigeher.db "SELECT COUNT(*) FROM users;"

# If count > 0, seed won't run (expected)
# To force re-seed (WARNING: DELETES ALL DATA):
sqlite3 gassigeher.db "DELETE FROM bookings; DELETE FROM dogs; DELETE FROM users;"

# Restart server - seed will run
./gassigeher
```

### Issue: Multiple Super Admins in Database

**Cause:** Database constraint not enforced, or manual SQL error.

**Solution:**
```sql
-- Find all super admins
SELECT id, email, is_super_admin FROM users WHERE is_super_admin = 1;

-- Keep only ID 1 as Super Admin
UPDATE users SET is_super_admin = 0 WHERE id != 1 AND is_super_admin = 1;

-- Ensure ID 1 is Super Admin
UPDATE users SET is_super_admin = 1, is_admin = 1 WHERE id = 1;
```

### Issue: "Forbidden: Super Admin required" When Promoting Users

**Cause:** JWT token doesn't have `is_super_admin` claim (old token).

**Solution:**
```bash
# Logout and login again to get new JWT with updated claims
# New token will include is_super_admin: true
```

---

## Security Considerations

### Super Admin Protection

1. **Cannot be deleted**: ID 1 is protected in all delete operations
2. **Cannot be deactivated**: Immune to auto-deactivation and manual deactivation
3. **Cannot be demoted**: No UI or API allows removing Super Admin flag
4. **Email fixed**: Must change `.env` and restart server (requires server access)

### Password File Security

1. **File permissions**: Set to `600` (owner read/write only)
2. **Gitignore**: File never committed to version control
3. **Plaintext acceptable**: Assumes server filesystem is secure
4. **No remote access**: File cannot be read via API or UI

### Admin Privilege Escalation Prevention

1. **Only Super Admin can promote**: Regular admins cannot promote others
2. **Cannot promote self**: No user can make themselves admin
3. **Audit trail**: All promotions/demotions logged (future enhancement)
4. **Database flags**: Source of truth (not config file)

### JWT Token Security

1. **Claims immutable**: Admin status in token (no repeated DB checks)
2. **Re-login required**: After promotion/demotion, user must re-login for new claims
3. **Token expiry**: 7-day expiry forces periodic re-validation

---

## Future Enhancements (Not Implemented)

These features are intentionally not included to keep the system simple:

1. ❌ **Audit Log**: Track all admin privilege changes
2. ❌ **Multi-Factor Authentication**: For Super Admin login
3. ❌ **Role-Based Access Control**: Different admin permission levels
4. ❌ **Password Complexity Rules**: For Super Admin password file
5. ❌ **CLI Management Tool**: Command-line admin management
6. ❌ **Email Notifications**: When user promoted/demoted
7. ❌ **Super Admin Transfer Wizard**: UI for changing Super Admin
8. ❌ **Factory Reset Button**: Too dangerous, manual only

**Reasoning**: Shelters need simple, reliable tools. Complex enterprise features add maintenance burden without clear benefit for this use case.

---

## Summary Checklist

### For New Installations

- [ ] Add `SUPER_ADMIN_EMAIL=your@email.com` to `.env`
- [ ] Remove `ADMIN_EMAILS` from `.env` (if exists)
- [ ] Start server for first time
- [ ] Copy Super Admin credentials from console or `SUPER_ADMIN_CREDENTIALS.txt`
- [ ] Login as Super Admin
- [ ] Change password via credentials file (optional)
- [ ] Delete test users (optional)
- [ ] Create real shelter users

### For Existing Installations

- [ ] Backup database
- [ ] Update code (`git pull`)
- [ ] Add `SUPER_ADMIN_EMAIL` to `.env`
- [ ] Remove `ADMIN_EMAILS` from `.env`
- [ ] Update database: Set Super Admin manually (SQL)
- [ ] Create `SUPER_ADMIN_CREDENTIALS.txt` with existing password
- [ ] Restart server
- [ ] Verify Super Admin login
- [ ] Promote other admins via UI (if needed)

### For Daily Operations

- [ ] Manage admins via `admin-users.html` (Super Admin only)
- [ ] Change Super Admin password via `SUPER_ADMIN_CREDENTIALS.txt`
- [ ] Admins cannot be auto-deactivated (protected)
- [ ] Super Admin can promote/demote without developer help

---

## Questions & Support

**Common Questions:**

**Q: Can I have multiple Super Admins?**
A: No, only one Super Admin (ID 1). Other users can be regular admins with most privileges.

**Q: What if I want to change Super Admin to a different person?**
A: Requires manual database update and `.env` change. Contact developer for assistance.

**Q: Can regular admins see who is Super Admin?**
A: Yes, Super Admin badge is visible to all admins in user management page.

**Q: What happens if `SUPER_ADMIN_CREDENTIALS.txt` is deleted?**
A: Create new file with new password, restart server. Old password cannot be recovered (need database backup).

**Q: Can I disable the file-based password system?**
A: No, this is the only password management method for simplicity. Alternative would require complex recovery flows.

---

**Document Version:** 1.0
**Last Updated:** 2025-01-23
**Applies to:** Gassigeher v2.0.0+
