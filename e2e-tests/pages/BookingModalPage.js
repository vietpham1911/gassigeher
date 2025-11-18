const BasePage = require('./BasePage');

/**
 * Booking Modal Page Object
 * For creating bookings via the modal dialog
 */
class BookingModalPage extends BasePage {
  constructor(page) {
    super(page);

    // Modal selectors
    this.modal = '#booking-modal, [role="dialog"]';
    this.modalTitle = '.modal-title, h2, h3';
    this.closeButton = '.modal-close, button:has-text("Schließen"), .close';

    // Form fields
    this.dateInput = '#booking-date, input[type="date"]';
    this.walkTypeSelect = '#booking-walk-type, #walk-type, select[name="walk_type"]';
    this.timeSelect = '#booking-time, #time, select[name="time"]';
    this.submitButton = '#booking-form button[type="submit"], button:has-text("Buchen"), button:has-text("Bestätigen")';

    // Validation
    this.errorMessage = '.error, .alert-error, .form-error';
    this.successMessage = '.success, .alert-success';
  }

  /**
   * Check if modal is visible
   */
  async isVisible() {
    return await this.page.locator(this.modal).isVisible().catch(() => false);
  }

  /**
   * Wait for modal to appear
   */
  async waitForModal(timeout = 5000) {
    await this.page.waitForSelector(this.modal, { state: 'visible', timeout });
  }

  /**
   * Get modal title
   */
  async getTitle() {
    await this.waitForModal();
    const titleElement = this.page.locator(this.modalTitle).first();
    return await titleElement.textContent();
  }

  /**
   * Fill booking form
   */
  async fillBookingForm({ date, walkType, time }) {
    await this.waitForModal();

    if (date) {
      const dateField = this.page.locator(this.dateInput).first();
      await dateField.fill(date);
    }

    if (walkType) {
      const walkTypeField = this.page.locator(this.walkTypeSelect).first();
      await walkTypeField.selectOption(walkType);
    }

    if (time) {
      const timeField = this.page.locator(this.timeSelect).first();
      await timeField.selectOption(time);
    }
  }

  /**
   * Create booking (fill form and submit)
   */
  async createBooking({ date, walkType, time }) {
    await this.fillBookingForm({ date, walkType, time });
    await this.submit();
  }

  /**
   * Submit booking form
   */
  async submit() {
    const submitBtn = this.page.locator(this.submitButton).first();
    await submitBtn.click();
    // Wait for modal to close or error to appear
    await this.page.waitForTimeout(1000);
  }

  /**
   * Close modal
   */
  async close() {
    const closeBtn = this.page.locator(this.closeButton).first();
    const btnExists = await closeBtn.count() > 0;
    if (btnExists) {
      await closeBtn.click();
      await this.page.waitForTimeout(500);
    } else {
      // Try pressing Escape key
      await this.page.keyboard.press('Escape');
    }
  }

  /**
   * Check if error message is shown
   */
  async hasError() {
    return await this.page.locator(this.errorMessage).isVisible().catch(() => false);
  }

  /**
   * Get error message text
   */
  async getErrorMessage() {
    await this.page.waitForSelector(this.errorMessage, { timeout: 3000 });
    return await this.page.locator(this.errorMessage).first().textContent();
  }

  /**
   * Check if success message is shown
   */
  async hasSuccess() {
    return await this.page.locator(this.successMessage).isVisible().catch(() => false);
  }
}

module.exports = BookingModalPage;

// DONE: Booking modal page object for creating bookings
