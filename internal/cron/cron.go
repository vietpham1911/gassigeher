package cron

import (
	"database/sql"
	"log"
	"time"

	"github.com/tranm/gassigeher/internal/repository"
)

// CronService handles scheduled tasks
type CronService struct {
	db          *sql.DB
	bookingRepo *repository.BookingRepository
	stopChan    chan bool
}

// NewCronService creates a new cron service
func NewCronService(db *sql.DB) *CronService {
	return &CronService{
		db:          db,
		bookingRepo: repository.NewBookingRepository(db),
		stopChan:    make(chan bool),
	}
}

// Start starts all cron jobs
func (s *CronService) Start() {
	log.Println("Starting cron service...")

	// Run auto-complete job every hour
	go s.runPeriodically("Auto-complete bookings", 1*time.Hour, s.autoCompleteBookings)

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
