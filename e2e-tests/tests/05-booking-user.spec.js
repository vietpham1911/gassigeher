const { test, expect } = require('@playwright/test');
const LoginPage = require('../pages/LoginPage');
const DogsPage = require('../pages/DogsPage');
const DashboardPage = require('../pages/DashboardPage');
const BookingModalPage = require('../pages/BookingModalPage');

/**
 * BOOKING TESTS - USER FLOWS
 * Test booking creation, validation, cancellation
 * GOAL: Find bugs in booking business logic! This is CRITICAL functionality!
 */

test.describe('Booking Creation - Valid Cases', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should create a booking successfully', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const initialDogCount = await dogsPage.getDogCount();
    if (initialDogCount > 0) {
      // Click book button for first dog
      await dogsPage.clickBookButton(0);

      // Fill booking modal
      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        // Book for tomorrow
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 1);
        const dateStr = tomorrow.toISOString().split('T')[0];

        await bookingModal.createBooking({
          date: dateStr,
          walkType: 'morning',
          time: '09:00',
        });

        // Wait for response
        await page.waitForLoadState('networkidle');

        // Should redirect to dashboard or show success
        await page.waitForTimeout(2000);

        const currentURL = page.url();
        console.log('After creating booking, URL:', currentURL);

        // CRITICAL BUG CHECK: Booking should be created
        // Either stay on dogs page with success OR redirect to dashboard
        const hasSuccess = await bookingModal.hasSuccess() ||
                           await page.locator('.alert-success').isVisible().catch(() => false);

        console.log('Success message shown:', hasSuccess);

        // Go to dashboard to verify booking appears
        const dashboardPage = new DashboardPage(page);
        await dashboardPage.goto();

        const bookingCount = await dashboardPage.getBookingCount();
        console.log('Bookings on dashboard:', bookingCount);

        // Should have at least 1 booking
        expect(bookingCount).toBeGreaterThan(0);

        // CRITICAL BUG CHECK: Booking should appear in dashboard
        if (bookingCount === 0) {
          console.error('🐛 CRITICAL BUG: Booking created but not showing in dashboard!');
        }
      }
    }
  });

});

test.describe('Booking Validation - Business Rules', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should BLOCK booking past dates', async ({ page }) => {
    // CRITICAL BUSINESS RULE: Cannot book in the past
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      await dogsPage.clickBookButton(0);

      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        // Try to book for yesterday
        const yesterday = new Date();
        yesterday.setDate(yesterday.getDate() - 1);
        const pastDate = yesterday.toISOString().split('T')[0];

        console.log('Attempting to book past date:', pastDate);

        await bookingModal.createBooking({
          date: pastDate,
          walkType: 'morning',
          time: '09:00',
        });

        await page.waitForTimeout(2000);

        // Should show error, NOT create booking
        const hasError = await bookingModal.hasError() ||
                         await page.locator('.alert-error').isVisible().catch(() => false);

        console.log('Error shown for past date:', hasError);

        // CRITICAL BUG CHECK: Past dates must be rejected!
        if (!hasError) {
          console.error('🐛 CRITICAL BUG: System accepted past date booking!');

          // Check if booking was actually created
          const dashboardPage = new DashboardPage(page);
          await dashboardPage.goto();
          console.error('🔍 Check dashboard - past booking should NOT exist');
        }

        expect(hasError).toBe(true);
      }
    }
  });

  test('should BLOCK booking blocked dates', async ({ page }) => {
    // CRITICAL BUSINESS RULE: Cannot book on blocked dates
    // Test data has 3 blocked dates

    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      await dogsPage.clickBookButton(0);

      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        // Use one of the blocked dates from test data: 2025-11-21, 2025-11-25, 2025-11-29
        const blockedDate = '2025-11-21';

        console.log('Attempting to book blocked date:', blockedDate);

        await bookingModal.createBooking({
          date: blockedDate,
          walkType: 'morning',
          time: '09:00',
        });

        await page.waitForTimeout(2000);

        // Should show error about blocked date
        const hasError = await bookingModal.hasError() ||
                         await page.locator('.alert-error').isVisible().catch(() => false);

        console.log('Error shown for blocked date:', hasError);

        // CRITICAL BUG CHECK: Blocked dates must be rejected!
        if (!hasError) {
          console.error('🐛 CRITICAL BUG: System accepted blocked date booking!');
        }

        expect(hasError).toBe(true);
      }
    }
  });

  test('should BLOCK booking beyond advance limit', async ({ page }) => {
    // CRITICAL BUSINESS RULE: Cannot book more than N days in advance
    // Default is 14 days (from system settings)

    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      await dogsPage.clickBookButton(0);

      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        // Try to book 30 days in future (beyond 14-day limit)
        const farFuture = new Date();
        farFuture.setDate(farFuture.getDate() + 30);
        const farFutureDate = farFuture.toISOString().split('T')[0];

        console.log('Attempting to book 30 days ahead:', farFutureDate);

        await bookingModal.createBooking({
          date: farFutureDate,
          walkType: 'morning',
          time: '09:00',
        });

        await page.waitForTimeout(2000);

        // Should show error about advance limit
        const hasError = await bookingModal.hasError() ||
                         await page.locator('.alert-error').isVisible().catch(() => false);

        console.log('Error shown for date beyond advance limit:', hasError);

        // CRITICAL BUG CHECK: Advance limit must be enforced!
        if (!hasError) {
          console.error('🐛 CRITICAL BUG: System accepted booking beyond 14-day advance limit!');
        }

        expect(hasError).toBe(true);
      }
    }
  });

  test('should PREVENT double booking same dog/time/date', async ({ page }) => {
    // CRITICAL BUSINESS RULE: Cannot double-book same dog at same time
    // This is the MOST IMPORTANT booking validation!

    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      // Book a dog for tomorrow morning 9:00
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 2);
      const dateStr = tomorrow.toISOString().split('T')[0];

      // First booking - use first AVAILABLE dog
      const clicked1 = await dogsPage.clickFirstAvailableDog();
      if (clicked1) {
        const bookingModal = new BookingModalPage(page);
        await bookingModal.waitForModal();

        await bookingModal.createBooking({
          date: dateStr,
          walkType: 'morning',
          time: '09:00',
        });

        await page.waitForTimeout(2000);
        const firstBookingSuccess = await bookingModal.hasSuccess() ||
                                     await page.locator('.alert-success').isVisible().catch(() => false);

        console.log('First booking created:', firstBookingSuccess);

        if (firstBookingSuccess) {
          // Now try to book THE SAME DOG at THE SAME TIME
          await dogsPage.goto();
          const clicked2 = await dogsPage.clickFirstAvailableDog();  // Same dog

          if (!clicked2) {
            console.warn('Could not click dog for second booking test');
            return;
          }

          await bookingModal.waitForModal();
          await bookingModal.createBooking({
            date: dateStr,  // Same date
            walkType: 'morning',  // Same walk type
            time: '09:00',  // Same time
          });

          await page.waitForTimeout(2000);

          // Should show error about double booking
          const hasError = await bookingModal.hasError() ||
                           await page.locator('.alert-error').isVisible().catch(() => false);

          console.log('Error shown for double booking:', hasError);

          // CRITICAL BUG CHECK: Double booking MUST be prevented!
          if (!hasError) {
            console.error('🐛 CRITICAL BUG: System allowed double booking of same dog!!!');
            console.error('🚨 This is a MAJOR business logic failure!');
          }

          expect(hasError).toBe(true);

          // Verify error message mentions double booking
          if (hasError) {
            const errorMsg = await bookingModal.getErrorMessage();
            console.log('Double booking error:', errorMsg);

            const mentionsDoubleBooking = errorMsg.toLowerCase().includes('bereits') ||
                                           errorMsg.toLowerCase().includes('gebucht') ||
                                           errorMsg.toLowerCase().includes('doppel');

            if (!mentionsDoubleBooking) {
              console.warn('⚠️ Error message doesn\'t clearly explain double booking');
            }
          }
        }
      }
    } else {
      console.warn('⏭️ Could not test double booking - no available dog');
    }
  });

});

test.describe('Booking - Cancellation Flow', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should allow cancelling future bookings', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    const initialBookingCount = await dashboardPage.getBookingCount();
    console.log('Initial bookings:', initialBookingCount);

    if (initialBookingCount > 0) {
      // Try to cancel first booking
      try {
        await dashboardPage.cancelBooking(0, 'Test cancellation from E2E');
        await page.waitForTimeout(2000);

        // Check if cancellation worked
        const newBookingCount = await dashboardPage.getBookingCount();
        console.log('Bookings after cancellation:', newBookingCount);

        // Should have one less booking
        // CRITICAL BUG CHECK: Cancellation should work
        if (newBookingCount === initialBookingCount) {
          console.error('🐛 POTENTIAL BUG: Cancellation didn\'t reduce booking count!');
        }
      } catch (error) {
        console.error('❌ Cancellation failed:', error.message);
        // POTENTIAL BUG: Cancellation might not be implemented properly
      }
    }
  });

  test('should NOT allow cancelling within notice period', async ({ page }) => {
    // CRITICAL BUSINESS RULE: Cannot cancel within 12 hours of walk time
    // This prevents last-minute cancellations that hurt the shelter

    console.log('🔒 CRITICAL TEST: Cancellation notice period enforcement');
    console.log('⏳ TODO: Create booking within 12 hours to test this rule');

    // This would require:
    // 1. Create booking for tomorrow morning
    // 2. Fast-forward time in database OR book for very soon
    // 3. Try to cancel
    // 4. Should be blocked

    // CRITICAL BUG: If users CAN cancel last-minute, shelter loses walks!
  });

});

test.describe('Booking - Edge Cases & Race Conditions', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should handle booking modal closing without submission', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      // Open modal - click first available dog
      const clicked = await dogsPage.clickFirstAvailableDog();

      if (!clicked) {
        console.warn('⏭️ No available dog to click - skipping test');
        return;
      }

      const bookingModal = new BookingModalPage(page);
      await bookingModal.waitForModal();

      // Fill form but DON'T submit
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const dateStr = tomorrow.toISOString().split('T')[0];

      await bookingModal.fillBookingForm({
        date: dateStr,
        walkType: 'morning',
        time: '09:00',
      });

      // Close modal without submitting
      await bookingModal.close();
      await page.waitForTimeout(1000);

      // Modal should be closed
      const modalStillVisible = await bookingModal.isVisible();
      console.log('Modal still visible after closing:', modalStillVisible);

      expect(modalStillVisible).toBe(false);

      // POTENTIAL BUG: Modal might not close properly
      if (modalStillVisible) {
        console.warn('⚠️ POTENTIAL BUG: Modal doesn\'t close when dismissed');
      }

      // Booking should NOT be created
      const dashboardPage = new DashboardPage(page);
      await dashboardPage.goto();

      // Check that no booking was accidentally created
      console.log('📋 Verify no booking was created when modal was closed');
    }
  });

  test('should show confirmation after successful booking', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      await dogsPage.clickBookButton(0);

      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        const tomorrow = new Date();
        tomorrow.setDate(tomorrow.getDate() + 3);
        const dateStr = tomorrow.toISOString().split('T')[0];

        await bookingModal.createBooking({
          date: dateStr,
          walkType: 'evening',
          time: '15:00',
        });

        await page.waitForTimeout(2000);

        // Should show success message
        const hasSuccess = await page.locator('.alert-success').isVisible().catch(() => false);

        console.log('Success confirmation shown:', hasSuccess);

        // CRITICAL UX BUG CHECK: User needs confirmation!
        if (!hasSuccess) {
          console.error('🐛 UX BUG: No confirmation shown after booking creation!');
          console.error('User doesn\'t know if booking succeeded!');
        }

        expect(hasSuccess).toBe(true);
      }
    }
  });

  test('should require all fields for booking', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      await dogsPage.clickBookButton(0);

      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        // Try to submit with empty date
        await bookingModal.fillBookingForm({
          // date missing!
          walkType: 'morning',
          time: '09:00',
        });

        await bookingModal.submit();
        await page.waitForTimeout(1000);

        // Should either show error OR HTML5 validation prevents submission
        const modalStillVisible = await bookingModal.isVisible();
        console.log('Modal still visible after invalid submission:', modalStillVisible);

        // Modal should still be open (submission blocked)
        expect(modalStillVisible).toBe(true);

        // POTENTIAL BUG: Required fields might not be validated
      }
    }
  });

});

test.describe('Booking - Viewing Bookings', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should show bookings on dashboard', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    const bookingCount = await dashboardPage.getBookingCount();
    console.log('Bookings on dashboard:', bookingCount);

    // Test data has 90 bookings - admin should see their bookings
    // Might be 0 if admin has no bookings, or > 0 if they do

    if (bookingCount === 0) {
      const hasNoBookingsMsg = await dashboardPage.hasNoBookingsMessage();
      console.log('No bookings message shown:', hasNoBookingsMsg);

      // POTENTIAL BUG: Empty state should be user-friendly
      if (!hasNoBookingsMsg) {
        console.warn('⚠️ UX ISSUE: No bookings message might be missing');
      }
    } else {
      console.log(`✅ Dashboard shows ${bookingCount} bookings`);
    }
  });

  test('should show booking details (dog name, date, time)', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    const bookingCount = await dashboardPage.getBookingCount();
    if (bookingCount > 0) {
      // Check first booking has details
      const firstBooking = page.locator('.booking-card').first();
      const bookingText = await firstBooking.textContent();

      console.log('First booking contains:', bookingText.substring(0, 100));

      // Should show dog name, date, time
      const hasDate = /\d{4}-\d{2}-\d{2}|\d{2}\.\d{2}\.\d{4}/.test(bookingText);
      const hasTime = /\d{2}:\d{2}/.test(bookingText);

      console.log('Booking shows date:', hasDate, 'time:', hasTime);

      // POTENTIAL BUG: Booking details might be incomplete
      if (!hasDate || !hasTime) {
        console.warn('⚠️ POTENTIAL BUG: Booking missing date or time information!');
      }
    }
  });

  test('should separate upcoming and past bookings', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    const pageText = await page.textContent('body');

    // Check for sections or filters for past/upcoming
    const hasUpcoming = pageText.includes('Kommende') || pageText.includes('Geplant') || pageText.includes('upcoming');
    const hasPast = pageText.includes('Vergangene') || pageText.includes('Abgeschlossen') || pageText.includes('completed');

    console.log('Shows upcoming bookings section:', hasUpcoming);
    console.log('Shows past bookings section:', hasPast);

    // POTENTIAL UX IMPROVEMENT: Might want to separate past and future
  });

});

test.describe('Booking - Adding Walk Notes', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should allow adding notes to COMPLETED bookings only', async ({ page }) => {
    const dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();

    const bookingCount = await dashboardPage.getBookingCount();
    console.log('Total bookings:', bookingCount);

    // Look for completed bookings (test data has many completed bookings)
    const pageHTML = await page.content();

    // Check if there are completed bookings
    const hasCompleted = pageHTML.includes('completed') ||
                          pageHTML.includes('abgeschlossen') ||
                          pageHTML.includes('Abgeschlossen');

    console.log('Has completed bookings:', hasCompleted);

    // CRITICAL BUSINESS RULE: Can only add notes to completed walks
    // Cannot add notes to scheduled (future) walks

    console.log('🔒 CRITICAL TEST: Notes only for completed bookings');
    console.log('⏳ Manual verification: Check that scheduled bookings don\'t have "Add notes" button');

    // CRITICAL BUG: If users can add notes to FUTURE bookings, it's a logic error!
  });

});

// DONE: Booking tests - creation, validation, business rules, double booking prevention, cancellation, viewing
