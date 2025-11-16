package cron

import (
	"database/sql"
	"log"
	"strconv"
	"time"

	"github.com/tranm/gassigeher/internal/repository"
)

// CronService handles scheduled tasks
type CronService struct {
	db           *sql.DB
	bookingRepo  *repository.BookingRepository
	userRepo     *repository.UserRepository
	settingsRepo *repository.SettingsRepository
	stopChan     chan bool
}

// NewCronService creates a new cron service
func NewCronService(db *sql.DB) *CronService {
	return &CronService{
		db:           db,
		bookingRepo:  repository.NewBookingRepository(db),
		userRepo:     repository.NewUserRepository(db),
		settingsRepo: repository.NewSettingsRepository(db),
		stopChan:     make(chan bool),
	}
}

// Start starts all cron jobs
func (s *CronService) Start() {
	log.Println("Starting cron service...")

	// Run auto-complete job every hour
	go s.runPeriodically("Auto-complete bookings", 1*time.Hour, s.autoCompleteBookings)

	// Run auto-deactivation job daily at 3am
	go s.runDaily("Auto-deactivate inactive users", 3, 0, s.autoDeactivateInactiveUsers)

	// Run booking reminder job every 15 minutes
	// Note: This is a placeholder for future implementation
	// go s.runPeriodically("Send booking reminders", 15*time.Minute, s.sendBookingReminders)
}

// Stop stops all cron jobs
func (s *CronService) Stop() {
	log.Println("Stopping cron service...")
	close(s.stopChan)
}

// runPeriodically runs a function periodically
func (s *CronService) runPeriodically(name string, interval time.Duration, fn func()) {
	// Run immediately on start
	fn()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Printf("Running cron job: %s", name)
			fn()
		case <-s.stopChan:
			log.Printf("Stopped cron job: %s", name)
			return
		}
	}
}

// autoCompleteBookings marks past scheduled bookings as completed
func (s *CronService) autoCompleteBookings() {
	count, err := s.bookingRepo.AutoComplete()
	if err != nil {
		log.Printf("Error auto-completing bookings: %v", err)
		return
	}

	if count > 0 {
		log.Printf("Auto-completed %d booking(s)", count)
	}
}

// sendBookingReminders sends reminders for upcoming bookings
// This is a placeholder for future implementation
func (s *CronService) sendBookingReminders() {
	// Get bookings that need reminders (1 hour before)
	bookings, err := s.bookingRepo.GetForReminders()
	if err != nil {
		log.Printf("Error getting bookings for reminders: %v", err)
		return
	}

	if len(bookings) == 0 {
		return
	}

	log.Printf("Found %d booking(s) that need reminders", len(bookings))

	// TODO: Send reminder emails
	// This requires getting user and dog details, then sending email
	// Will be implemented when email reminder system is fully set up
}

// runDaily runs a function daily at a specific time
func (s *CronService) runDaily(name string, hour, minute int, fn func()) {
	for {
		now := time.Now()
		next := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		
		// If we've passed today's scheduled time, schedule for tomorrow
		if now.After(next) {
			next = next.Add(24 * time.Hour)
		}

		duration := next.Sub(now)
		log.Printf("Scheduling daily job '%s' to run in %v (at %s)", name, duration, next.Format("2006-01-02 15:04:05"))

		select {
		case <-time.After(duration):
			log.Printf("Running daily job: %s", name)
			fn()
		case <-s.stopChan:
			log.Printf("Stopped daily job: %s", name)
			return
		}
	}
}

// autoDeactivateInactiveUsers deactivates users who haven't been active for the configured period
func (s *CronService) autoDeactivateInactiveUsers() {
	// Get deactivation period from settings
	setting, err := s.settingsRepo.Get("auto_deactivation_days")
	if err != nil {
		log.Printf("Error getting auto_deactivation_days setting: %v", err)
		return
	}

	days := 365 // default 1 year
	if setting != nil {
		if d, err := strconv.Atoi(setting.Value); err == nil {
			days = d
		}
	}

	// Find inactive users
	users, err := s.userRepo.FindInactiveUsers(days)
	if err != nil {
		log.Printf("Error finding inactive users: %v", err)
		return
	}

	if len(users) == 0 {
		log.Println("No inactive users to deactivate")
		return
	}

	log.Printf("Found %d inactive user(s) to deactivate", len(users))

	// Deactivate each user
	for _, user := range users {
		if err := s.userRepo.Deactivate(user.ID, "auto_inactivity"); err != nil {
			log.Printf("Error deactivating user %d: %v", user.ID, err)
			continue
		}

		log.Printf("Auto-deactivated user %d (inactive for %d days)", user.ID, days)
	}
}
