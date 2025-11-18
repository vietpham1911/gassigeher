const BasePage = require('./BasePage');

/**
 * Dogs Page Object
 * For browsing and filtering dogs
 */
class DogsPage extends BasePage {
  constructor(page) {
    super(page);

    // Selectors (from actual rendered HTML)
    this.dogCards = '.dog-card';
    this.dogName = '.dog-card-title';  // Corrected: Actual class name
    this.dogCardBody = '.dog-card-body';
    this.lockedBanner = '.dog-locked-banner';
    this.unavailableBanner = '.dog-unavailable-banner';
    this.categoryBadge = '.dog-category-badge';

    // Filters (from actual HTML)
    this.breedFilter = '#filter-breed';
    this.categoryFilter = '#filter-category';
    this.sizeFilter = '#filter-size';
    this.searchInput = '#filter-search';  // Corrected ID
    this.applyFiltersButton = 'button:has-text("Anwenden")';  // MUST click this!
    this.resetFiltersButton = 'button:has-text("Zurücksetzen")';

    // No results
    this.noResultsMessage = '.no-results, .empty-state';
  }

  /**
   * Navigate to dogs page
   */
  async goto() {
    await super.goto('/dogs.html');
  }

  /**
   * Get number of dog cards displayed
   */
  async getDogCount() {
    await this.page.waitForLoadState('networkidle');
    const count = await this.page.locator(this.dogCards).count();
    return count;
  }

  /**
   * Check if "no results" message is shown
   */
  async hasNoResults() {
    return await this.page.locator(this.noResultsMessage).isVisible().catch(() => false);
  }

  /**
   * Filter by breed
   */
  async filterByBreed(breed) {
    await this.page.selectOption(this.breedFilter, breed);
    await this.page.click(this.applyFiltersButton);  // MUST click Apply!
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Filter by category (experience level)
   */
  async filterByCategory(category) {
    await this.page.selectOption(this.categoryFilter, category);
    await this.page.click(this.applyFiltersButton);  // MUST click Apply!
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Filter by size
   */
  async filterBySize(size) {
    await this.page.selectOption(this.sizeFilter, size);
    await this.page.click(this.applyFiltersButton);  // MUST click Apply!
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Search dogs by name
   */
  async searchDogs(query) {
    await this.page.fill(this.searchInput, query);
    await this.page.click(this.applyFiltersButton);  // MUST click Apply!
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Clear all filters
   */
  async resetFilters() {
    await this.page.click(this.resetFiltersButton);
    await this.page.waitForLoadState('networkidle');
  }

  /**
   * Click dog card to open booking modal
   * NOTE: Whole card is clickable, no separate book button
   * Only works for accessible, available dogs!
   */
  async clickDogCard(index = 0) {
    const dogCards = this.page.locator(this.dogCards);
    const card = dogCards.nth(index);

    // Check if dog is clickable (not locked or unavailable)
    const classes = await card.getAttribute('class');
    const isLocked = classes.includes('locked');
    const isUnavailable = classes.includes('unavailable');

    if (isLocked || isUnavailable) {
      console.warn(`⚠️ Dog at index ${index} is locked or unavailable, cannot click`);
      return false;
    }

    // Click the dog card (whole card is clickable)
    await card.click();
    // Wait for booking modal to appear
    await this.page.waitForTimeout(1000);
    return true;
  }

  /**
   * Find and click first AVAILABLE dog
   */
  async clickFirstAvailableDog() {
    const dogCards = this.page.locator(this.dogCards);
    const count = await dogCards.count();

    for (let i = 0; i < count; i++) {
      const card = dogCards.nth(i);
      const classes = await card.getAttribute('class');
      const isAvailable = !classes.includes('locked') && !classes.includes('unavailable');

      if (isAvailable) {
        console.log(`Found available dog at index ${i}`);
        await card.click();
        await this.page.waitForTimeout(1000);
        return true;
      }
    }

    console.error('❌ No available dogs found to click!');
    return false;
  }

  /**
   * Alias for clickDogCard (for backwards compatibility)
   */
  async clickBookButton(index = 0) {
    await this.clickDogCard(index);
  }

  /**
   * Check if dog is locked (experience level too high)
   */
  async isDogLocked(index = 0) {
    const dogCards = this.page.locator(this.dogCards);
    const card = dogCards.nth(index);
    // Check for locked class or locked banner
    const hasLockedClass = await card.getAttribute('class').then(c => c.includes('locked')).catch(() => false);
    const hasLockedBanner = await card.locator(this.lockedBanner).isVisible().catch(() => false);
    return hasLockedClass || hasLockedBanner;
  }

  /**
   * Get dog name by index
   */
  async getDogName(index = 0) {
    const dogCards = this.page.locator(this.dogCards);
    const card = dogCards.nth(index);
    const nameElement = card.locator(this.dogName).first();
    return await nameElement.textContent();
  }

  /**
   * Check if dog is available (not unavailable)
   */
  async isDogAvailable(index = 0) {
    const dogCards = this.page.locator(this.dogCards);
    const card = dogCards.nth(index);
    const cardText = await card.textContent();
    // Look for "nicht verfügbar" or unavailable indicators
    return !cardText.toLowerCase().includes('nicht verfügbar');
  }
}

module.exports = DogsPage;

// DONE: Dogs page object for browsing and filtering dogs
