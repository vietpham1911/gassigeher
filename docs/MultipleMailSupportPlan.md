# Multi-Provider Email Support Plan - Gmail, SMTP (Strato, etc.)

**Created:** 2025-01-22
**Status:** Planning Document
**Priority:** Medium
**Complexity:** Medium
**Estimated Duration:** 3-4 days

---

## Executive Summary

This document outlines the comprehensive plan for adding SMTP email provider support to the Gassigeher application, alongside the existing Gmail API implementation. The goal is to provide flexible email sending options while maintaining Gmail as the default for easy setup.

**Key Principle:** Gmail remains the default email provider for development and small deployments. SMTP (for Strato, Office365, custom servers, etc.) is an optional alternative for deployments that cannot or prefer not to use Gmail API.

**IMPORTANT CLARIFICATION:**
- **SMTP** is for **SENDING** emails (what we need)
- **IMAP/POP3** are for **RECEIVING** emails (not needed for this application)
- Strato uses `smtp.strato.de` for sending (not imap/pop3)

---

## Table of Contents

1. [Current State Analysis](#1-current-state-analysis)
2. [Requirements](#2-requirements)
3. [Email Provider Comparison](#3-email-provider-comparison)
4. [Architecture Design](#4-architecture-design)
5. [Implementation Phases](#5-implementation-phases)
6. [Configuration](#6-configuration)
7. [Testing Strategy](#7-testing-strategy)
8. [Security Considerations](#8-security-considerations)
9. [Migration Guide](#9-migration-guide)
10. [Deployment Guide](#10-deployment-guide)

---

## 1. Current State Analysis

### 1.1 Current Architecture ✅ **Good Foundation**

**Email Layer:**
- `internal/services/email_service.go` - Gmail API implementation
- `internal/services/email_account.go` - Account lifecycle emails
- 17 email types (verification, bookings, admin, account lifecycle)
- HTML email templates with inline CSS

**What's Good:**
- ✅ Clean EmailService struct with clear interface
- ✅ All email methods use SendEmail() base method
- ✅ HTML templates with inline CSS (works everywhere)
- ✅ German language emails
- ✅ Goroutine-based async sending
- ✅ Graceful failure handling

**What Needs Adaptation:**
- ❌ Hardcoded Gmail API dependency
- ❌ OAuth2-specific initialization
- ❌ No abstraction for different providers
- ❌ Single provider approach

### 1.2 Gmail API Usage

**Current Implementation:**
```go
type EmailService struct {
    service   *gmail.Service  // Gmail-specific
    fromEmail string
}

func NewEmailService(clientID, clientSecret, refreshToken, fromEmail string) (*EmailService, error) {
    // Gmail OAuth2 setup
    config := &oauth2.Config{
        ClientID:     clientID,
        ClientSecret: clientSecret,
        Endpoint:     google.Endpoint,
        Scopes:       []string{gmail.GmailSendScope},
    }
    // ... Gmail service creation
}
```

**Gmail API Pros:**
- ✅ Free (within quotas)
- ✅ Reliable delivery
- ✅ No spam issues (using own domain)
- ✅ OAuth2 security

**Gmail API Cons:**
- ❌ Requires Google Cloud setup
- ❌ OAuth2 refresh token complexity
- ❌ Quota limits (100 emails/day for free)
- ❌ Cannot use with non-Gmail addresses easily

### 1.3 Email Methods Inventory

**17 Email Types:**
1. `SendEmail(to, subject, body)` - Base method
2. `SendVerificationEmail()` - Email verification
3. `SendWelcomeEmail()` - Welcome after verification
4. `SendPasswordResetEmail()` - Password reset
5. `SendBookingConfirmation()` - Booking created
6. `SendBookingCancellation()` - User cancelled
7. `SendAdminCancellation()` - Admin cancelled
8. `SendBookingReminder()` - 1 hour before walk
9. `SendBookingMoved()` - Admin moved booking
10. `SendExperienceLevelApproved()` - Level promotion approved
11. `SendExperienceLevelDenied()` - Level promotion denied
12. `SendAccountDeactivated()` - Account deactivated
13. `SendAccountReactivated()` - Account reactivated
14. `SendReactivationDenied()` - Reactivation denied
15. `SendAccountDeletionConfirmation()` - GDPR deletion

**All methods call `SendEmail()` internally - perfect for abstraction!**

### 1.4 Current Configuration

```bash
# .env (current)
GMAIL_CLIENT_ID=your-client-id
GMAIL_CLIENT_SECRET=your-client-secret
GMAIL_REFRESH_TOKEN=your-refresh-token
GMAIL_FROM_EMAIL=noreply@gassigeher.com
```

---

## 2. Requirements

### 2.1 Functional Requirements

**FR1: Support Multiple Email Providers**
- Gmail API (current) - for easy setup
- SMTP (new) - for Strato, Office365, custom servers, etc.

**FR2: Provider Selection**
- Configure via environment variable `EMAIL_PROVIDER`
- Default to Gmail if not specified
- Connection string or individual parameters

**FR3: Feature Parity**
- All 17 email types work identically across providers
- Same HTML templates for all providers
- No feature degradation

**FR4: Backward Compatibility**
- Existing Gmail API setup continues to work
- No breaking changes to existing deployments
- Smooth migration path from Gmail to SMTP

**FR5: SMTP Provider Support**
- **Strato**: smtp.strato.de (port 465 SSL or 587 TLS)
- **Office365**: smtp.office365.com (port 587 TLS)
- **Gmail SMTP**: smtp.gmail.com (port 587 TLS)
- **Custom**: Any SMTP server

**FR6: Admin Email Copy (BCC)**
- All sent emails automatically BCC'd to admin mailbox for record-keeping
- Configure via `EMAIL_BCC_ADMIN` environment variable
- Optional feature (empty = disabled)
- Transparent to recipients (blind carbon copy)
- Works with all providers (Gmail API, SMTP)
- Enables email audit trail and compliance
- Admin can review all system communications

### 2.2 Non-Functional Requirements

**NFR1: Security**
- Secure credential storage (environment variables)
- TLS/SSL encryption for SMTP
- No plaintext passwords in logs

**NFR2: Reliability**
- Retry logic for failed sends
- Graceful degradation (log but don't crash)
- Connection pooling for SMTP

**NFR3: Testing**
- Unit tests for each provider
- Integration tests with real SMTP servers (Mailtrap, etc.)
- Mock tests for all 17 email types

**NFR4: Documentation**
- Provider selection guide
- Configuration examples for each provider
- Troubleshooting guide

---

## 3. Email Provider Comparison

### 3.1 Use Cases

| Feature | Gmail API | SMTP (Strato) | SMTP (Generic) |
|---------|-----------|---------------|----------------|
| **Setup Complexity** | ⭐⭐⭐ High (OAuth2) | ⭐ Easy (username/password) | ⭐ Easy |
| **Free Tier** | 100 emails/day | Depends on plan | Depends on provider |
| **Reliability** | ⭐⭐⭐⭐⭐ Excellent | ⭐⭐⭐⭐ Good | ⭐⭐⭐ Varies |
| **Spam Risk** | ⭐⭐⭐⭐⭐ Very Low | ⭐⭐⭐ Medium | ⭐⭐ High (if misconfigured) |
| **Use Own Domain** | ✅ Yes (via Gmail) | ✅ Yes | ✅ Yes |
| **Authentication** | OAuth2 | Username/Password | Username/Password |
| **Port Options** | API only | 465 (SSL), 587 (TLS) | Varies |

### 3.2 When to Use Each Provider

**Use Gmail API if:**
- Small to medium deployment (<100 emails/day)
- Can setup Google Cloud project
- Want maximum deliverability
- Don't mind OAuth2 complexity

**Use SMTP (Strato) if:**
- Already have Strato email hosting
- Need more than 100 emails/day
- Prefer username/password over OAuth2
- Want to use existing email infrastructure

**Use SMTP (Generic) if:**
- Have Office365, custom SMTP server, etc.
- Corporate email requirement
- Need full control over email infrastructure

---

## 4. Architecture Design

### 4.1 Proposed Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│                    (Handlers, Services)                      │
└────────────────────────┬────────────────────────────────────┘
                         │
┌────────────────────────┴────────────────────────────────────┐
│                   Email Service Layer                        │
│              (Sends all 17 email types)                      │
│  ┌──────────────────────────────────────────────────┐      │
│  │  SendVerificationEmail()                         │      │
│  │  SendBookingConfirmation()                       │      │
│  │  SendAccountDeactivated()                        │      │
│  │  ... (14 more methods)                           │      │
│  └──────────────────┬───────────────────────────────┘      │
│                     │ All call SendEmail()                   │
│                     ▼                                         │
│  ┌──────────────────────────────────────────────────┐      │
│  │  SendEmail(to, subject, body string) error       │      │
│  └──────────────────┬───────────────────────────────┘      │
└─────────────────────┼────────────────────────────────────────┘
                      │
┌─────────────────────┴────────────────────────────────────────┐
│              Email Provider Abstraction (NEW)                 │
│  ┌────────────────────────────────────────────────────┐    │
│  │  EmailProvider Interface                           │    │
│  │  • SendEmail(to, subject, body) error              │    │
│  │  • ValidateConfig() error                          │    │
│  │  • Close() error                                   │    │
│  └────────────────────────────────────────────────────┘    │
│                                                              │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐        │
│  │   Gmail     │  │    SMTP     │  │   Future    │        │
│  │  Provider   │  │  Provider   │  │  Providers  │        │
│  └─────────────┘  └─────────────┘  └─────────────┘        │
└──────────────────────────────────────────────────────────────┘
```

### 4.2 EmailProvider Interface

**New File:** `internal/services/email_provider.go`

```go
package services

// EmailProvider defines the interface for email sending
type EmailProvider interface {
    // SendEmail sends an email with HTML body
    // Automatically includes BCC if configured
    SendEmail(to, subject, body string) error

    // ValidateConfig validates the provider configuration
    ValidateConfig() error

    // Close closes any open connections
    Close() error

    // GetFromEmail returns the from email address
    GetFromEmail() string
}

// EmailConfig holds configuration for email providers
type EmailConfig struct {
    Provider string // "gmail" or "smtp"

    // Gmail API settings
    GmailClientID     string
    GmailClientSecret string
    GmailRefreshToken string
    GmailFromEmail    string

    // SMTP settings
    SMTPHost      string
    SMTPPort      int
    SMTPUsername  string
    SMTPPassword  string
    SMTPFromEmail string
    SMTPUseTLS    bool  // Use STARTTLS (port 587)
    SMTPUseSSL    bool  // Use SSL/TLS (port 465)

    // BCC settings (applies to all providers)
    BCCAdmin string  // Optional: BCC all emails to this address for audit trail
}
```

### 4.3 Provider Implementations

**Files to Create:**
- `internal/services/email_provider_gmail.go` - Gmail API implementation
- `internal/services/email_provider_smtp.go` - SMTP implementation
- `internal/services/email_provider_factory.go` - Factory to create providers

**Gmail Provider Example:**
```go
type GmailProvider struct {
    service   *gmail.Service
    fromEmail string
    bccAdmin  string  // BCC address for admin copy
}

func NewGmailProvider(config *EmailConfig) (EmailProvider, error) {
    // Current Gmail implementation moved here
    // Store bccAdmin from config
    // Returns GmailProvider that implements EmailProvider interface
}

func (p *GmailProvider) SendEmail(to, subject, body string) error {
    // Current SendEmail logic
    // Add BCC header if bccAdmin is set
    // Format: "To: {to}\r\nBcc: {bccAdmin}\r\n..."
}
```

**SMTP Provider Example:**
```go
type SMTPProvider struct {
    host      string
    port      int
    username  string
    password  string
    fromEmail string
    bccAdmin  string  // BCC address for admin copy
    useTLS    bool
    useSSL    bool
    auth      smtp.Auth
}

func NewSMTPProvider(config *EmailConfig) (EmailProvider, error) {
    // Create SMTP provider
    // Setup authentication
    // Store bccAdmin from config
    // Return SMTPProvider that implements EmailProvider interface
}

func (p *SMTPProvider) SendEmail(to, subject, body string) error {
    // Standard SMTP sending with net/smtp
    // Include BCC in recipients list if bccAdmin is set
    // Add BCC header to message
    // Support TLS and SSL
    // Handle authentication
}
```

### 4.4 Updated EmailService

**File:** `internal/services/email_service.go`

```go
type EmailService struct {
    provider EmailProvider  // Interface instead of *gmail.Service
}

func NewEmailService(config *EmailConfig) (*EmailService, error) {
    // Use factory to create provider based on config.Provider
    provider, err := NewEmailProvider(config)
    if err != nil {
        return nil, err
    }

    return &EmailService{
        provider: provider,
    }, nil
}

func (s *EmailService) SendEmail(to, subject, body string) error {
    // Delegate to provider
    return s.provider.SendEmail(to, subject, body)
}

// All 17 email methods remain unchanged!
// They all call s.SendEmail() which now uses the provider interface
```

---

## 5. Implementation Phases

### Phase 1: Abstraction Layer (Day 1)

**Goal:** Create email provider abstraction without breaking existing code

**Tasks:**
1. **Create EmailProvider Interface** (`internal/services/email_provider.go`)
   - Define interface methods
   - Create EmailConfig struct
   - Document interface

2. **Create Provider Factory** (`internal/services/email_provider_factory.go`)
   - Factory pattern to create providers
   - Validation and error handling

3. **Refactor Gmail to Provider** (`internal/services/email_provider_gmail.go`)
   - Move current Gmail logic to GmailProvider
   - Implement EmailProvider interface
   - Add BCC support to SendEmail method
   - Keep all functionality identical

4. **Update EmailService** (modify `internal/services/email_service.go`)
   - Change from `*gmail.Service` to `EmailProvider` interface
   - Update NewEmailService to use factory
   - All 17 email methods remain unchanged
   - BCC automatically applied to all email types

**Acceptance Criteria:**
- ✅ EmailProvider interface defined with BCC support
- ✅ GmailProvider implements interface
- ✅ BCC functionality works (when EMAIL_BCC_ADMIN set)
- ✅ BCC disabled when EMAIL_BCC_ADMIN empty
- ✅ All existing tests pass
- ✅ No breaking changes
- ✅ Gmail still works identically

**Files Created:**
- `internal/services/email_provider.go` (interface)
- `internal/services/email_provider_factory.go` (factory)
- `internal/services/email_provider_gmail.go` (Gmail implementation)

**Files Modified:**
- `internal/services/email_service.go` (use interface)

---

### Phase 2: SMTP Provider Implementation (Day 2)

**Goal:** Implement SMTP email provider

**Tasks:**
1. **Create SMTP Provider** (`internal/services/email_provider_smtp.go`)
   - Implement EmailProvider interface
   - Use Go's `net/smtp` package
   - Support TLS (STARTTLS, port 587)
   - Support SSL (port 465)
   - Handle authentication (PLAIN, LOGIN, CRAM-MD5)

2. **MIME Email Formatting**
   - Proper MIME headers
   - HTML email support
   - UTF-8 encoding
   - From/To/Subject/BCC headers

3. **BCC Implementation**
   - Add BCC header if BCCAdmin is set in config
   - Include BCC address in recipient list (SMTP RCPT TO)
   - Ensure BCC is invisible to primary recipient

4. **Connection Management**
   - Connection pooling (optional)
   - Timeout handling
   - Retry logic

5. **Error Handling**
   - Specific SMTP errors
   - Connection failures
   - Authentication failures

**Acceptance Criteria:**
- ✅ SMTPProvider implements EmailProvider
- ✅ Sends emails via standard SMTP
- ✅ Supports TLS and SSL
- ✅ Handles authentication
- ✅ HTML emails formatted correctly
- ✅ German umlauts work (UTF-8)
- ✅ BCC works correctly (admin receives copy)
- ✅ BCC disabled when not configured

**Dependencies:**
```go
import (
    "crypto/tls"
    "net/smtp"
    "net/mail"
    "fmt"
    "strings"
)
```

---

### Phase 3: Configuration Updates (Day 2)

**Goal:** Add SMTP configuration support

**Tasks:**
1. **Update Config Struct** (`internal/config/config.go`)
   ```go
   type Config struct {
       // Email Provider Selection
       EmailProvider string // "gmail" or "smtp"

       // Gmail settings (existing)
       GmailClientID     string
       GmailClientSecret string
       GmailRefreshToken string
       GmailFromEmail    string

       // SMTP settings (new)
       SMTPHost      string
       SMTPPort      int
       SMTPUsername  string
       SMTPPassword  string
       SMTPFromEmail string
       SMTPUseTLS    bool
       SMTPUseSSL    bool

       // ... existing fields
   }
   ```

2. **Environment Variable Loading**
   - Load EMAIL_PROVIDER (default: "gmail")
   - Load Gmail settings (existing)
   - Load SMTP settings (new)
   - Validation

3. **Create EmailConfig Helper**
   ```go
   func (c *Config) GetEmailConfig() *services.EmailConfig {
       return &services.EmailConfig{
           Provider: c.EmailProvider,
           // Map config fields to EmailConfig
       }
   }
   ```

**Acceptance Criteria:**
- ✅ EMAIL_PROVIDER env var supported
- ✅ Gmail config still works (backward compatible)
- ✅ SMTP config fully supported
- ✅ Validation errors for missing fields
- ✅ Sensible defaults

---

### Phase 4: Application Integration (Day 3)

**Goal:** Integrate new provider system into application

**Tasks:**
1. **Update Handler Initialization**
   - Change all `NewEmailService()` calls
   - Pass `EmailConfig` instead of individual params
   - Handle initialization errors gracefully

2. **Update All Handlers** (12 handlers that use email)
   ```go
   // Old way
   emailService, err := services.NewEmailService(
       cfg.GmailClientID,
       cfg.GmailClientSecret,
       cfg.GmailRefreshToken,
       cfg.GmailFromEmail,
   )

   // New way
   emailService, err := services.NewEmailService(cfg.GetEmailConfig())
   ```

3. **Logging**
   - Log which provider is being used
   - Log connection successes/failures
   - Don't log passwords!

**Acceptance Criteria:**
- ✅ Application starts with Gmail provider (default)
- ✅ Application starts with SMTP provider (if configured)
- ✅ All 17 email types work with both providers
- ✅ Graceful failure if email config invalid
- ✅ No breaking changes for existing deployments

---

### Phase 5: Testing (Day 3)

**Goal:** Comprehensive testing of both providers

**Tasks:**
1. **Unit Tests**
   - Test EmailProvider interface
   - Test GmailProvider
   - Test SMTPProvider
   - Test factory

2. **Integration Tests**
   - Use Mailtrap.io for SMTP testing
   - Test all 17 email types
   - Test HTML rendering
   - Test German umlauts (UTF-8)

3. **Mock Tests**
   - Mock EmailProvider interface
   - Test handler logic without real emails
   - Test error handling

4. **Manual Testing**
   - Test with Strato SMTP (smtp.strato.de)
   - Test with Gmail SMTP (smtp.gmail.com)
   - Test with Mailtrap (sandbox)

**Test Configuration (Mailtrap):**
```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=587
SMTP_USERNAME=your-mailtrap-username
SMTP_PASSWORD=your-mailtrap-password
SMTP_FROM_EMAIL=test@gassigeher.com
SMTP_USE_TLS=true
```

**Acceptance Criteria:**
- ✅ All unit tests pass for both providers
- ✅ Integration tests with Mailtrap pass
- ✅ All 17 email types work identically
- ✅ HTML emails render correctly
- ✅ German umlauts display properly

---

### Phase 6: Documentation (Day 4)

**Goal:** Complete documentation for email providers

**Tasks:**
1. **Update Main Documentation**
   - README.md - Email provider options
   - DEPLOYMENT.md - SMTP setup instructions
   - CLAUDE.md - Email provider patterns

2. **Create Provider Selection Guide**
   - When to use Gmail API
   - When to use SMTP
   - Provider comparison table

3. **Create Setup Guides**
   - **Strato SMTP Setup Guide**
   - **Office365 SMTP Setup Guide**
   - **Gmail SMTP Setup Guide** (alternative to API)
   - **Generic SMTP Setup Guide**

4. **Create Migration Guide**
   - Migrate from Gmail API to SMTP
   - Configuration changes needed
   - Testing checklist

**Acceptance Criteria:**
- ✅ All documentation updated
- ✅ Setup guide for each provider
- ✅ Migration guide complete
- ✅ Configuration examples provided

---

## 6. Configuration

### 6.1 Environment Variables

**New Variables:**

```bash
# ==================================================
# Email Provider Selection (default: gmail)
# ==================================================
EMAIL_PROVIDER=gmail  # or "smtp"

# ==================================================
# BCC Admin Copy (optional - works with all providers)
# ==================================================
# All sent emails will be BCC'd to this address for audit trail
# Leave empty to disable
EMAIL_BCC_ADMIN=admin@yourdomain.com

# ==================================================
# Gmail API Configuration (existing - unchanged)
# ==================================================
GMAIL_CLIENT_ID=your-client-id.apps.googleusercontent.com
GMAIL_CLIENT_SECRET=your-client-secret
GMAIL_REFRESH_TOKEN=your-refresh-token
GMAIL_FROM_EMAIL=noreply@gassigeher.com

# ==================================================
# SMTP Configuration (new)
# ==================================================

# Strato Example
SMTP_HOST=smtp.strato.de
SMTP_PORT=465
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-email-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_SSL=true
SMTP_USE_TLS=false

# Office365 Example
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-email-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false

# Gmail SMTP Example (alternative to Gmail API)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password  # Not regular password!
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false
```

### 6.2 Configuration Examples

**Option A: Gmail API (Default - Existing)**

```bash
EMAIL_PROVIDER=gmail
GMAIL_CLIENT_ID=123456-abc.apps.googleusercontent.com
GMAIL_CLIENT_SECRET=GOCSPX-abc123
GMAIL_REFRESH_TOKEN=1//abc123...
GMAIL_FROM_EMAIL=noreply@gassigeher.com

# Optional: BCC all emails to admin for audit trail
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

**Option B: Strato SMTP**

```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.strato.de
SMTP_PORT=465
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=yourpassword
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_SSL=true
SMTP_USE_TLS=false

# Optional: BCC all emails to admin for audit trail
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

**Option C: Office365 SMTP**

```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=yourpassword
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false

# Optional: BCC all emails to admin for audit trail
EMAIL_BCC_ADMIN=admin@yourdomain.com
```

### 6.3 BCC Admin Copy Feature

**Purpose**: Keep a copy of all sent emails in an admin mailbox for:
- **Audit trail**: Track all system communications
- **Compliance**: GDPR/legal requirements for data processing
- **Support**: Help users by reviewing sent emails
- **Debugging**: Verify email content and delivery
- **Record-keeping**: Archive of all notifications

**How it works:**
```
User Registration
      ↓
SendVerificationEmail(user@example.com, ...)
      ↓
Email sent to:
- TO: user@example.com (recipient sees this)
- BCC: admin@yourdomain.com (recipient CANNOT see this)
      ↓
Admin mailbox receives copy for records
```

**Configuration:**
```bash
# Enable BCC to admin
EMAIL_BCC_ADMIN=admin@yourdomain.com

# Disable BCC (leave empty)
EMAIL_BCC_ADMIN=
```

**Technical Implementation:**
- BCC header added automatically by EmailProvider
- Recipient cannot see BCC address (blind copy)
- Works identically with Gmail API and SMTP
- No changes to existing email methods needed
- All 17 email types automatically BCC'd

**Use Cases:**
1. **Compliance Officer** receives all account deletion confirmations
2. **Support Team** can review booking confirmations sent to users
3. **Administrator** monitors password reset requests
4. **Legal Requirements** maintain communication records

**Storage Recommendations:**
- Use dedicated mailbox (e.g., archive@yourdomain.com)
- Set up email retention rules (auto-delete after X months)
- Consider mailbox size limits (may grow large)
- Use email filtering to organize by type

**Privacy Note:** Inform users in privacy policy that copies of system emails may be retained for administrative purposes.

### 6.4 Port Selection

| Port | Protocol | Security | Use Case |
|------|----------|----------|----------|
| **25** | SMTP | None (plaintext) | ❌ Never use (blocked by ISPs) |
| **465** | SMTPS | SSL/TLS from start | ✅ Use with SMTP_USE_SSL=true |
| **587** | SMTP | STARTTLS | ✅ Use with SMTP_USE_TLS=true (recommended) |
| **2525** | SMTP | STARTTLS | ⚠️ Alternative to 587 (if blocked) |

---

## 7. Testing Strategy

### 7.1 Test Providers

**Mailtrap (Recommended for Development)**
```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=sandbox.smtp.mailtrap.io
SMTP_PORT=587
SMTP_USERNAME=your-mailtrap-username
SMTP_PASSWORD=your-mailtrap-password
SMTP_FROM_EMAIL=test@gassigeher.com
SMTP_USE_TLS=true
```

**MailHog (Local SMTP Server)**
```bash
docker run -d -p 1025:1025 -p 8025:8025 mailhog/mailhog

EMAIL_PROVIDER=smtp
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_FROM_EMAIL=test@localhost
SMTP_USE_TLS=false
```

### 7.2 Test Matrix

| Test Case | Gmail API | SMTP (Strato) | SMTP (Office365) | SMTP (Generic) |
|-----------|-----------|---------------|------------------|----------------|
| **Send verification email** | ✅ | ✅ | ✅ | ✅ |
| **Send booking confirmation** | ✅ | ✅ | ✅ | ✅ |
| **Send with German umlauts** | ✅ | ✅ | ✅ | ✅ |
| **HTML email rendering** | ✅ | ✅ | ✅ | ✅ |
| **BCC admin copy (when enabled)** | ✅ | ✅ | ✅ | ✅ |
| **BCC disabled (when empty)** | ✅ | ✅ | ✅ | ✅ |
| **All 17 email types** | ✅ | ✅ | ✅ | ✅ |
| **Error handling** | ✅ | ✅ | ✅ | ✅ |
| **Connection retry** | ✅ | ✅ | ✅ | ✅ |

### 7.3 Unit Tests

**File:** `internal/services/email_provider_test.go`

```go
func TestEmailProviders(t *testing.T) {
    providers := []struct {
        name     string
        provider EmailProvider
    }{
        {"Gmail", newMockGmailProvider()},
        {"SMTP", newMockSMTPProvider()},
    }

    for _, p := range providers {
        t.Run(p.name, func(t *testing.T) {
            // Test SendEmail
            err := p.provider.SendEmail("test@example.com", "Test", "<p>Test</p>")
            assert.NoError(t, err)

            // Test ValidateConfig
            err = p.provider.ValidateConfig()
            assert.NoError(t, err)

            // Test GetFromEmail
            from := p.provider.GetFromEmail()
            assert.NotEmpty(t, from)
        })
    }
}
```

### 7.4 Integration Tests

**File:** `internal/services/email_integration_test.go`

```go
func TestEmailService_AllProviders(t *testing.T) {
    // Skip if not running integration tests
    if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
        t.Skip("Skipping integration tests")
    }

    providers := []string{"gmail", "smtp"}

    for _, provider := range providers {
        t.Run(provider, func(t *testing.T) {
            config := getTestConfig(provider)
            emailService, err := NewEmailService(config)
            assert.NoError(t, err)

            // Test all 17 email types
            testAllEmailTypes(t, emailService)
        })
    }
}
```

---

## 8. Security Considerations

### 8.1 Credential Storage

**DO:**
- ✅ Store credentials in environment variables
- ✅ Use app-specific passwords (Gmail SMTP)
- ✅ Encrypt credentials at rest (if storing in DB)
- ✅ Use TLS/SSL for all SMTP connections
- ✅ Validate all configuration on startup

**DON'T:**
- ❌ Store passwords in code
- ❌ Store passwords in version control
- ❌ Log passwords (even debug mode)
- ❌ Send emails over unencrypted connections
- ❌ Use port 25 (insecure, blocked by ISPs)

### 8.2 SMTP Authentication

**Supported Methods:**
1. **PLAIN** - Username/password (over TLS)
2. **LOGIN** - Similar to PLAIN (legacy)
3. **CRAM-MD5** - Challenge-response (more secure)

**Gmail SMTP Requirement:**
- Must use "App Password" (not regular password)
- Enable 2FA on Google account first
- Generate app password at: https://myaccount.google.com/apppasswords

### 8.3 Email Spoofing Prevention

**SPF Record** (DNS):
```
v=spf1 include:_spf.strato.de ~all
```

**DKIM** (Configured by email provider)

**DMARC Record** (DNS):
```
v=DMARC1; p=none; rua=mailto:dmarc@yourdomain.com
```

---

## 9. Migration Guide

### 9.1 Gmail API → SMTP (Strato)

**Step 1: Get Strato SMTP Credentials**
- Login to Strato admin panel
- Go to Email settings
- Note: smtp.strato.de, port 465 (SSL) or 587 (TLS)
- Username: your-email@yourdomain.com
- Password: your email password

**Step 2: Update .env File**
```bash
# Change from Gmail API
# EMAIL_PROVIDER=gmail
# GMAIL_CLIENT_ID=...
# GMAIL_CLIENT_SECRET=...
# GMAIL_REFRESH_TOKEN=...
# GMAIL_FROM_EMAIL=noreply@gassigeher.com

# To SMTP (Strato)
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.strato.de
SMTP_PORT=465
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-email-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_SSL=true
SMTP_USE_TLS=false
```

**Step 3: Restart Application**
```bash
sudo systemctl restart gassigeher
```

**Step 4: Test Email Sending**
```bash
# Register a new user and check email delivery
# Check logs for any errors
sudo journalctl -u gassigeher -n 50 | grep -i email
```

**Step 5: Monitor for Issues**
- Check email delivery rates
- Watch for bounce messages
- Monitor spam complaints

---

## 10. Deployment Guide

### 10.1 Strato SMTP Setup

**Prerequisites:**
- Strato email package
- Email address created (e.g., noreply@yourdomain.com)
- Domain verified

**Configuration:**
```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.strato.de
SMTP_PORT=465          # SSL (recommended)
# OR
SMTP_PORT=587          # TLS (alternative)
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-strato-email-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_SSL=true      # If using port 465
SMTP_USE_TLS=false     # If using port 465
# OR
SMTP_USE_SSL=false     # If using port 587
SMTP_USE_TLS=true      # If using port 587
```

**Strato-Specific Notes:**
- Port 465 (SSL) recommended for Strato
- Port 587 (TLS) also works
- May have sending limits (check your plan)
- SPF/DKIM automatically configured by Strato

### 10.2 Office365 SMTP Setup

**Prerequisites:**
- Office365 account
- Email address (e.g., noreply@yourdomain.com)
- Basic authentication enabled (or use app password)

**Configuration:**
```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.office365.com
SMTP_PORT=587
SMTP_USERNAME=noreply@yourdomain.com
SMTP_PASSWORD=your-office365-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false
```

**Office365-Specific Notes:**
- Must use port 587 with TLS
- Sending limit: 10,000 emails/day
- May require app password if 2FA enabled

### 10.3 Gmail SMTP Setup (Alternative to Gmail API)

**Prerequisites:**
- Gmail account
- 2FA enabled
- App password generated

**Configuration:**
```bash
EMAIL_PROVIDER=smtp
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password  # NOT regular password!
SMTP_FROM_EMAIL=your-email@gmail.com
SMTP_USE_TLS=true
SMTP_USE_SSL=false
```

**Gmail SMTP Notes:**
- Limit: 500 emails/day (2,000 for Google Workspace)
- Must use app password: https://myaccount.google.com/apppasswords
- Less setup than Gmail API (no OAuth2)

---

## 11. Troubleshooting

### 11.1 Common SMTP Errors

**Error: "Connection refused"**
- Check SMTP_HOST is correct
- Check SMTP_PORT is correct
- Check firewall rules
- Verify port is not blocked by ISP

**Error: "Authentication failed"**
- Check SMTP_USERNAME is correct (usually full email)
- Check SMTP_PASSWORD is correct
- For Gmail: Must use app password, not regular password
- For Office365: May need app password if 2FA enabled

**Error: "TLS handshake failed"**
- Check SMTP_USE_TLS or SMTP_USE_SSL is correct
- Port 587 → use SMTP_USE_TLS=true
- Port 465 → use SMTP_USE_SSL=true
- Check server supports TLS/SSL

**Error: "Emails not arriving"**
- Check spam folder
- Check SPF/DKIM/DMARC records
- Check email provider logs
- Use Mailtrap for testing first

### 11.2 Testing Checklist

- [ ] Application starts without errors
- [ ] Email provider logged on startup
- [ ] Test email sends successfully
- [ ] HTML emails render correctly
- [ ] German umlauts display properly (ä, ö, ü, ß)
- [ ] All 17 email types work
- [ ] Emails arrive in inbox (not spam)
- [ ] From address displays correctly
- [ ] Error handling works (wrong password, etc.)

---

## 12. Future Enhancements

### Short-term

1. **Retry Logic**
   - Automatic retry on failed sends
   - Exponential backoff
   - Queue failed emails

2. **Email Queuing**
   - Queue emails in database
   - Background worker to send
   - Retry failed sends

3. **Email Templates in DB**
   - Store templates in database
   - Admin UI to edit templates
   - A/B testing for emails

### Long-term

1. **Additional Providers**
   - SendGrid API
   - Mailgun API
   - Amazon SES
   - Postmark

2. **Email Analytics**
   - Track email opens
   - Track link clicks
   - Delivery reports

3. **Email Preferences**
   - User opt-out of certain emails
   - Email frequency settings
   - Digest emails

---

## 13. Summary

### Estimated Effort

| Phase | Duration | Complexity |
|-------|----------|------------|
| Phase 1: Abstraction Layer | 1 day | Medium |
| Phase 2: SMTP Implementation | 1 day | Medium |
| Phase 3: Configuration | 0.5 day | Low |
| Phase 4: Integration | 0.5 day | Low |
| Phase 5: Testing | 1 day | Medium |
| Phase 6: Documentation | 0.5 day | Low |
| **Total** | **4-5 days** | **Medium** |

### Benefits

✅ **Flexibility**: Choose email provider based on needs
✅ **No Gmail API Dependency**: Use simple SMTP if preferred
✅ **Strato Support**: Direct support for Strato hosting
✅ **Cost Effective**: Use existing email infrastructure
✅ **Backward Compatible**: Gmail API still works
✅ **Standard Protocol**: SMTP works everywhere

### Risks

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Gmail users break | High | Low | Comprehensive testing, backward compatibility |
| SMTP delivery issues | Medium | Medium | Test with Mailtrap, monitor bounce rates |
| Authentication complexity | Medium | Low | Clear documentation, examples |
| Port blocking | Low | Medium | Support multiple ports (465, 587, 2525) |

---

**Document Version:** 1.0
**Last Updated:** 2025-01-22
**Author:** Claude Code
**Review Status:** Ready for Implementation
**Approval Required:** Yes (impacts email infrastructure)
