package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/tranm/gassigeher/internal/config"
	"github.com/tranm/gassigeher/internal/cron"
	"github.com/tranm/gassigeher/internal/database"
	"github.com/tranm/gassigeher/internal/handlers"
	"github.com/tranm/gassigeher/internal/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.CORSMiddleware)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	userHandler := handlers.NewUserHandler(db, cfg)
	dogHandler := handlers.NewDogHandler(db, cfg)
	bookingHandler := handlers.NewBookingHandler(db, cfg)
	blockedDateHandler := handlers.NewBlockedDateHandler(db, cfg)
	settingsHandler := handlers.NewSettingsHandler(db, cfg)
	experienceHandler := handlers.NewExperienceRequestHandler(db, cfg)
	reactivationHandler := handlers.NewReactivationRequestHandler(db, cfg)

	// Start cron service for auto-completion and reminders
	cronService := cron.NewCronService(db)
	cronService.Start()
	defer cronService.Stop()

	// Public routes
	router.HandleFunc("/api/auth/register", authHandler.Register).Methods("POST")
	router.HandleFunc("/api/auth/verify-email", authHandler.VerifyEmail).Methods("POST")
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods("POST")
	router.HandleFunc("/api/auth/forgot-password", authHandler.ForgotPassword).Methods("POST")
	router.HandleFunc("/api/auth/reset-password", authHandler.ResetPassword).Methods("POST")

	// Reactivation request (public - for deactivated users)
	router.HandleFunc("/api/reactivation-requests", reactivationHandler.CreateRequest).Methods("POST")

	// Protected routes (authenticated users)
	protected := router.PathPrefix("/api").Subrouter()
	protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))

	// Auth
	protected.HandleFunc("/auth/change-password", authHandler.ChangePassword).Methods("PUT")

	// Users
	protected.HandleFunc("/users/me", userHandler.GetMe).Methods("GET")
	protected.HandleFunc("/users/me", userHandler.UpdateMe).Methods("PUT")
	protected.HandleFunc("/users/me/photo", userHandler.UploadPhoto).Methods("POST")
	protected.HandleFunc("/users/me", userHandler.DeleteAccount).Methods("DELETE")

	// Dogs (read-only for authenticated users)
	protected.HandleFunc("/dogs", dogHandler.ListDogs).Methods("GET")
	protected.HandleFunc("/dogs/breeds", dogHandler.GetBreeds).Methods("GET")
	protected.HandleFunc("/dogs/{id}", dogHandler.GetDog).Methods("GET")

	// Bookings (authenticated users)
	protected.HandleFunc("/bookings", bookingHandler.ListBookings).Methods("GET")
	protected.HandleFunc("/bookings", bookingHandler.CreateBooking).Methods("POST")
	protected.HandleFunc("/bookings/{id}", bookingHandler.GetBooking).Methods("GET")
	protected.HandleFunc("/bookings/{id}/cancel", bookingHandler.CancelBooking).Methods("PUT")
	protected.HandleFunc("/bookings/{id}/notes", bookingHandler.AddNotes).Methods("PUT")
	protected.HandleFunc("/bookings/calendar/{year}/{month}", bookingHandler.GetCalendarData).Methods("GET")

	// Blocked dates (read-only for authenticated users)
	protected.HandleFunc("/blocked-dates", blockedDateHandler.ListBlockedDates).Methods("GET")

	// Experience requests (authenticated users)
	protected.HandleFunc("/experience-requests", experienceHandler.CreateRequest).Methods("POST")
	protected.HandleFunc("/experience-requests", experienceHandler.ListRequests).Methods("GET")

	// Admin-only routes
	admin := protected.PathPrefix("").Subrouter()
	admin.Use(middleware.RequireAdmin)

	// Dog management (admin only)
	admin.HandleFunc("/dogs", dogHandler.CreateDog).Methods("POST")
	admin.HandleFunc("/dogs/{id}", dogHandler.UpdateDog).Methods("PUT")
	admin.HandleFunc("/dogs/{id}", dogHandler.DeleteDog).Methods("DELETE")
	admin.HandleFunc("/dogs/{id}/photo", dogHandler.UploadDogPhoto).Methods("POST")
	admin.HandleFunc("/dogs/{id}/availability", dogHandler.ToggleAvailability).Methods("PUT")

	// Blocked dates management (admin only)
	admin.HandleFunc("/blocked-dates", blockedDateHandler.CreateBlockedDate).Methods("POST")
	admin.HandleFunc("/blocked-dates/{id}", blockedDateHandler.DeleteBlockedDate).Methods("DELETE")

	// Booking management (admin only)
	admin.HandleFunc("/bookings/{id}/move", bookingHandler.MoveBooking).Methods("PUT")

	// System settings (admin only)
	admin.HandleFunc("/settings", settingsHandler.GetAllSettings).Methods("GET")
	admin.HandleFunc("/settings/{key}", settingsHandler.UpdateSetting).Methods("PUT")

	// Experience requests management (admin only)
	admin.HandleFunc("/experience-requests/{id}/approve", experienceHandler.ApproveRequest).Methods("PUT")
	admin.HandleFunc("/experience-requests/{id}/deny", experienceHandler.DenyRequest).Methods("PUT")

	// User management (admin only)
	admin.HandleFunc("/users", userHandler.ListUsers).Methods("GET")
	admin.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	admin.HandleFunc("/users/{id}/activate", userHandler.ActivateUser).Methods("PUT")
	admin.HandleFunc("/users/{id}/deactivate", userHandler.DeactivateUser).Methods("PUT")

	// Reactivation requests management (admin only)
	admin.HandleFunc("/reactivation-requests", reactivationHandler.ListRequests).Methods("GET")
	admin.HandleFunc("/reactivation-requests/{id}/approve", reactivationHandler.ApproveRequest).Methods("PUT")
	admin.HandleFunc("/reactivation-requests/{id}/deny", reactivationHandler.DenyRequest).Methods("PUT")

	// Static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./frontend")))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
