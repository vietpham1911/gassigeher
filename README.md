# Gassigeher - Dog Walking Booking System

A complete web-based dog walking booking system built with Go and Vanilla JavaScript.

## Features

- User registration with email verification
- JWT-based authentication
- Password reset flow
- GDPR-compliant account deletion
- Automatic user deactivation after inactivity
- German UI with i18n support
- Mobile-first responsive design
- Gmail API integration for email notifications

## Tech Stack

**Backend:**
- Go 1.24+
- SQLite database
- gorilla/mux router
- JWT authentication
- bcrypt password hashing
- Gmail API for emails

**Frontend:**
- Vanilla JavaScript (ES6+)
- HTML5 & CSS3
- Custom i18n system
- No external dependencies

## Project Structure

```
gassigeher/
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── config/               # Configuration management
│   ├── database/             # Database setup and migrations
│   ├── handlers/             # HTTP request handlers
│   ├── middleware/           # Auth, logging, CORS middleware
│   ├── models/               # Data models
│   ├── repository/           # Database operations
│   └── services/             # Business logic (auth, email)
├── frontend/
│   ├── assets/
│   │   └── css/              # Stylesheets
│   ├── i18n/                 # Translation files
│   ├── js/                   # JavaScript modules
│   ├── index.html            # Landing page
│   ├── login.html            # Login page
│   ├── register.html         # Registration page
│   ├── verify.html           # Email verification
│   ├── forgot-password.html  # Password reset request
│   ├── reset-password.html   # Password reset
│   └── terms.html            # Terms & Conditions
├── migrations/               # Database migrations
├── uploads/                  # User and dog photos
├── .env                      # Environment variables
├── .env.example              # Environment template
├── go.mod                    # Go dependencies
├── go.sum                    # Go dependencies checksums
├── ImplementationPlan.md     # Complete implementation plan
└── README.md                 # This file
```

## Setup

### 1. Prerequisites

- Go 1.24 or higher
- SQLite3
- Gmail account (for email notifications)

### 2. Clone and Install

```bash
cd gassigeher
go mod download
```

### 3. Configure Environment

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Edit `.env` and set your configuration, especially:
- `JWT_SECRET`: Generate a secure random string
- `ADMIN_EMAILS`: Your admin email addresses
- Gmail API credentials (see below)

### 4. Gmail API Setup

To enable email notifications:

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Gmail API
4. Create OAuth 2.0 credentials
5. Download credentials and get:
   - Client ID
   - Client Secret
   - Refresh Token (use OAuth Playground or your app to generate)
6. Add these to your `.env` file

**Note:** For development, you can skip Gmail setup. The app will run but emails won't be sent.

### 5. Build and Test

**Windows:**
```cmd
bat.bat
```

**Linux/Mac:**
```bash
chmod +x bat.sh
./bat.sh
```

This will:
- Check Go installation
- Download dependencies
- Build the application
- Run all tests

### 6. Run the Application

**Development mode:**
```bash
go run cmd/server/main.go
```

**Using compiled binary:**

Windows:
```cmd
gassigeher.exe
```

Linux/Mac:
```bash
./gassigeher
```

The server will start on `http://localhost:8080`

### 7. Custom Port

```bash
# Windows
set PORT=3000 && gassigeher.exe

# Linux/Mac
PORT=3000 ./gassigeher
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/verify-email` - Verify email with token
- `POST /api/auth/login` - Login
- `POST /api/auth/forgot-password` - Request password reset
- `POST /api/auth/reset-password` - Reset password with token
- `PUT /api/auth/change-password` - Change password (authenticated)

### Users
- `GET /api/users/me` - Get current user profile
- `PUT /api/users/me` - Update profile
- `POST /api/users/me/photo` - Upload profile photo

## Database

The application uses SQLite with automatic migrations. The database file is created automatically on first run at the path specified in `DATABASE_PATH` (default: `./gassigeher.db`).

### Tables Created
- `users` - User accounts and profiles
- `dogs` - Dog information
- `bookings` - Walk bookings
- `blocked_dates` - Admin-blocked dates
- `experience_requests` - User level promotion requests
- `reactivation_requests` - Account reactivation requests
- `system_settings` - Configurable system settings

## Phase 1 Status: DONE ✅

Phase 1 (Foundation) is complete with:
- ✅ Go backend with all auth endpoints
- ✅ SQLite database with migrations
- ✅ JWT authentication
- ✅ Email verification flow
- ✅ Password reset flow
- ✅ Gmail API integration
- ✅ Frontend with German i18n
- ✅ All authentication pages (register, login, verify, reset)
- ✅ Responsive design with Tierheim Göppingen colors

## Development Notes

### Color Scheme (Tierheim Göppingen)
- Primary Green: `#82b965`
- Dark Background: `#26272b`
- Dark Gray: `#33363b`
- Border Radius: `6px`

### Admin Access
Admins are defined in the `ADMIN_EMAILS` environment variable. Users with these emails get admin privileges automatically upon login.

### Testing Accounts

For development, you can:
1. Register a new account
2. Verify via email link (check console logs if Gmail not configured)
3. Login with your credentials

## Next Phases

- **Phase 2**: Dog Management (CRUD, photos, categories)
- **Phase 3**: Booking System (calendar, availability)
- **Phase 4**: Blocked Dates & Admin Actions
- **Phase 5**: Experience Levels (Green/Blue/Orange)
- **Phase 6**: User Profiles & Photos
- **Phase 7**: Account Management & GDPR
- **Phase 8**: Admin Dashboard & Reports
- **Phase 9**: Polish & Testing (90% coverage)
- **Phase 10**: Deployment

See `ImplementationPlan.md` for complete details.

## Contributing

This is a complete application following the implementation plan. Each phase builds upon the previous one with comprehensive testing and documentation.

## License

© 2025 Gassigeher. All rights reserved.
