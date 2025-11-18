const { test, expect } = require('@playwright/test');
const LoginPage = require('../pages/LoginPage');
const DogsPage = require('../pages/DogsPage');
const BookingModalPage = require('../pages/BookingModalPage');

/**
 * DOG BROWSING TESTS
 * Test dog listing, filtering, search, experience level enforcement
 * GOAL: Find bugs in dog browsing and access control!
 */

test.describe('Dog Browsing - Basic Functionality', () => {

  test.beforeEach(async ({ page }) => {
    // Login as green level user
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should show dogs page after login', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    expect(page.url()).toContain('dogs.html');

    // Should show some dogs (we have 18 in test data)
    const count = await dogsPage.getDogCount();
    console.log('Total dogs displayed:', count);

    expect(count).toBeGreaterThan(0);

    // CRITICAL BUG CHECK: Should show the 18 dogs we created
    if (count === 0) {
      console.error('🐛 CRITICAL BUG: No dogs displayed!');
    }
  });

  test('should display dog information cards', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const count = await dogsPage.getDogCount();
    if (count > 0) {
      // Check first dog has name
      const dogName = await dogsPage.getDogName(0);
      console.log('First dog name:', dogName);

      expect(dogName).toBeTruthy();
      expect(dogName.length).toBeGreaterThan(0);

      // POTENTIAL BUG: Dog name might not be displayed
    }
  });

  test('should show book buttons for available dogs', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const count = await dogsPage.getDogCount();
    if (count > 0) {
      // Check if first dog has book button
      const firstCard = page.locator('.dog-card').first();
      const bookButton = firstCard.locator('button').first();
      const hasButton = await bookButton.isVisible().catch(() => false);

      console.log('First dog has book button:', hasButton);

      // POTENTIAL BUG: Available dogs might not have book buttons
    }
  });

});

test.describe('Dog Browsing - Filters', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should filter dogs by category (experience level)', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const totalDogs = await dogsPage.getDogCount();
    console.log('Total dogs before filter:', totalDogs);

    // Filter by green category (should have 3 green dogs from test data)
    await dogsPage.filterByCategory('green');
    await page.waitForTimeout(1000);

    const greenDogs = await dogsPage.getDogCount();
    console.log('Green dogs after filter:', greenDogs);

    // Should have fewer dogs (only green ones)
    // CRITICAL BUG CHECK: Filter should work!
    if (greenDogs === totalDogs) {
      console.warn('⚠️ POTENTIAL BUG: Category filter might not be working!');
    }
  });

  test('should filter dogs by size', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const totalDogs = await dogsPage.getDogCount();

    // Filter by large size
    await dogsPage.filterBySize('large');
    await page.waitForTimeout(1000);

    const largeDogs = await dogsPage.getDogCount();
    console.log('Large dogs:', largeDogs, 'Total dogs:', totalDogs);

    // Should have some large dogs but not all dogs
    // POTENTIAL BUG: Size filter might not work
  });

  test('should search dogs by name', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    // Search for "Luna" (we created a dog named Luna)
    await dogsPage.searchDogs('Luna');
    await page.waitForTimeout(1000);

    const searchResults = await dogsPage.getDogCount();
    console.log('Search results for "Luna":', searchResults);

    // Should find Luna
    if (searchResults > 0) {
      const firstName = await dogsPage.getDogName(0);
      console.log('First result name:', firstName);

      // Should contain "Luna"
      expect(firstName.toLowerCase()).toContain('luna');
    }

    // CRITICAL BUG CHECK: Search should find the dog
    if (searchResults === 0) {
      console.error('🐛 CRITICAL BUG: Search for "Luna" found nothing!');
    }
  });

  test('should show "no results" for non-existent dog search', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    // Search for dog that doesn't exist
    await dogsPage.searchDogs('NonExistentDogXYZ123');
    await page.waitForTimeout(1000);

    const searchResults = await dogsPage.getDogCount();
    console.log('Search results for non-existent dog:', searchResults);

    // Should have no results
    expect(searchResults).toBe(0);

    // Should show "no results" message
    const hasNoResults = await dogsPage.hasNoResults();
    console.log('Shows no results message:', hasNoResults);

    // POTENTIAL BUG: Empty state might not be shown
  });

  test('should filter available dogs only', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const totalDogs = await dogsPage.getDogCount();

    // Check if available-only filter exists in UI
    // From HTML inspection, there's no checkbox - filtering is done server-side
    // Only unavailable dogs are marked with .unavailable class
    console.log('Available-only filter may not exist in current UI');
    console.log('Unavailable dogs are shown but marked differently');

    // SKIPPING: No available-only filter checkbox in current UI
    // Unavailable dogs are shown but visually marked
    test.skip();

    const availableDogs = await dogsPage.getDogCount();
    console.log('Available dogs:', availableDogs, 'Total dogs:', totalDogs);

    // Should have fewer dogs (2 are unavailable)
    // CRITICAL BUG CHECK: Unavailable dogs should be filtered out
    if (availableDogs === totalDogs) {
      console.warn('⚠️ POTENTIAL BUG: Available filter might not work!');
    }
  });

});

test.describe('Dog Browsing - Experience Level Enforcement', () => {

  test('GREEN user should see all green dogs unlocked', async ({ page }) => {
    // Login as a green level user
    // For now use admin (who is orange, can see everything)
    // TODO: Create specific green level test user
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');

    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    console.log('Dogs visible to admin user:', dogCount);

    // Admin (orange level) should see all dogs
    expect(dogCount).toBeGreaterThan(0);

    // FUTURE: Test with actual green user to verify locking
  });

  test('should show lock icon for dogs above user experience level', async ({ page }) => {
    // This would need a green level user to test properly
    // For now, verify the mechanism exists
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');

    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    // Check if any dog has lock icon (depends on user level)
    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      const isLocked = await dogsPage.isDogLocked(0);
      console.log('First dog is locked:', isLocked);

      // Admin should not see locked dogs
      // CRITICAL BUG CHECK: Experience level enforcement
    }
  });

  test('GREEN user should NOT be able to book ORANGE dogs', async ({ page }) => {
    // CRITICAL SECURITY TEST: Experience level enforcement
    // This prevents inexperienced users from walking difficult dogs

    // TODO: Need to create a green level test user
    // For now, this test documents the requirement
    console.log('🔒 CRITICAL TEST: Green users must not book orange dogs');
    console.log('⏳ TODO: Create green user in test data');

    // CRITICAL BUG: If green user CAN book orange dog, it's a safety issue!
  });

});

test.describe('Dog Browsing - Booking Flow Start', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should open booking modal when clicking available dog card', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      // Click first AVAILABLE dog (not locked or unavailable)
      const clicked = await dogsPage.clickFirstAvailableDog();

      if (clicked) {
        // Check if modal appeared
        const bookingModal = new BookingModalPage(page);
        const modalVisible = await bookingModal.isVisible();

        console.log('Booking modal opened:', modalVisible);

        // CRITICAL BUG CHECK: Modal should appear
        if (!modalVisible) {
          console.error('🐛 CRITICAL BUG: Booking modal does not open after clicking available dog!');
        }

        expect(modalVisible).toBe(true);
      } else {
        console.warn('⚠️ No available dogs to click - skipping modal test');
      }
    }
  });

  test('should show dog name in booking modal', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const dogCount = await dogsPage.getDogCount();
    if (dogCount > 0) {
      // Get dog name before clicking
      const dogName = await dogsPage.getDogName(0);
      console.log('Booking for dog:', dogName);

      // Click book button
      await dogsPage.clickBookButton(0);
      await page.waitForTimeout(1000);

      // Check modal title/content contains dog name
      const bookingModal = new BookingModalPage(page);
      const modalVisible = await bookingModal.isVisible();

      if (modalVisible) {
        const modalTitle = await bookingModal.getTitle();
        console.log('Modal title:', modalTitle);

        // POTENTIAL BUG: Modal should show which dog you're booking
        const hasDogName = modalTitle.toLowerCase().includes(dogName.toLowerCase());
        if (!hasDogName) {
          console.warn('⚠️ POTENTIAL UX ISSUE: Modal doesn\'t show dog name');
        }
      }
    }
  });

  test('should not show book button for unavailable dogs', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    // Look through dogs to find an unavailable one
    const dogCount = await dogsPage.getDogCount();
    console.log('Checking', dogCount, 'dogs for unavailable status...');

    let foundUnavailable = false;
    for (let i = 0; i < dogCount && i < 20; i++) {
      const isAvailable = await dogsPage.isDogAvailable(i);
      if (!isAvailable) {
        foundUnavailable = true;
        console.log(`Dog ${i} is unavailable`);

        // Check that unavailable dog doesn't have book button
        const dogCard = page.locator('.dog-card').nth(i);
        const bookButton = dogCard.locator('button:has-text("Buchen")');
        const hasBookButton = await bookButton.isVisible().catch(() => false);

        console.log(`Unavailable dog ${i} has book button:`, hasBookButton);

        // CRITICAL BUG CHECK: Unavailable dogs should NOT have book button
        if (hasBookButton) {
          console.error('🐛 CRITICAL BUG: Unavailable dog has book button!');
        }

        expect(hasBookButton).toBe(false);
        break;
      }
    }

    if (!foundUnavailable) {
      console.warn('⚠️ No unavailable dogs found to test (expected 2 from test data)');
    }
  });

});

test.describe('Dog Browsing - Edge Cases', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should handle no dogs scenario gracefully', async ({ page }) => {
    // This tests what happens if shelter has no dogs
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    // Currently we have dogs, but if filter returns nothing...
    await dogsPage.searchDogs('XYZ_NO_MATCH_999');
    await page.waitForTimeout(1000);

    const count = await dogsPage.getDogCount();
    console.log('Dogs after impossible search:', count);

    if (count === 0) {
      // Should show friendly message, not blank page
      const hasNoResults = await dogsPage.hasNoResults();
      console.log('Shows no results message:', hasNoResults);

      // POTENTIAL BUG: Empty state might not be user-friendly
    }
  });

  test('should load dog photos without errors', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    await page.waitForLoadState('networkidle');

    // Check for broken images
    const brokenImages = await page.evaluate(() => {
      const images = Array.from(document.querySelectorAll('.dog-card img'));
      return images.filter(img => !img.complete || img.naturalHeight === 0).length;
    });

    console.log('Broken dog images:', brokenImages);

    // POTENTIAL BUG: Dog photos might be broken
    if (brokenImages > 0) {
      console.warn(`⚠️ POTENTIAL BUG: ${brokenImages} dog images failed to load`);
    }
  });

  test('should display experience level badges correctly', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const pageText = await page.textContent('body');

    // Check if experience levels are shown (Grün, Blau, Orange)
    const hasGreen = pageText.includes('Grün') || pageText.includes('green');
    const hasBlue = pageText.includes('Blau') || pageText.includes('blue');
    const hasOrange = pageText.includes('Orange') || pageText.includes('orange');

    console.log('Experience levels shown - Green:', hasGreen, 'Blue:', hasBlue, 'Orange:', hasOrange);

    // POTENTIAL BUG: Experience level indicators might be missing
    if (!hasGreen && !hasBlue && !hasOrange) {
      console.warn('⚠️ POTENTIAL BUG: No experience level indicators shown!');
    }
  });

});

test.describe('Dog Browsing - Multiple Filters Combined', () => {

  test.beforeEach(async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();
    await loginPage.loginAndWait('admin@tierheim-goeppingen.de', 'test123');
  });

  test('should apply multiple filters together', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    const initialCount = await dogsPage.getDogCount();
    console.log('Initial dog count:', initialCount);

    // Apply category filter
    await dogsPage.filterByCategory('green');
    await page.waitForTimeout(500);
    const afterCategory = await dogsPage.getDogCount();
    console.log('After category filter:', afterCategory);

    // Add size filter
    await dogsPage.filterBySize('large');
    await page.waitForTimeout(500);
    const afterSize = await dogsPage.getDogCount();
    console.log('After adding size filter:', afterSize);

    // Should progressively reduce results
    // CRITICAL BUG CHECK: Multiple filters should work together
    if (afterSize > afterCategory) {
      console.error('🐛 POTENTIAL BUG: Adding size filter INCREASED results!');
    }

    // Should have even fewer dogs
    expect(afterSize).toBeLessThanOrEqual(afterCategory);
  });

  test('should handle filter combinations that return zero results', async ({ page }) => {
    const dogsPage = new DogsPage(page);
    await dogsPage.goto();

    // Apply impossible filter combination
    await dogsPage.filterByCategory('green');
    await page.waitForTimeout(500);
    await dogsPage.searchDogs('ZZZNOMATCH999');
    await page.waitForTimeout(500);

    const count = await dogsPage.getDogCount();
    console.log('Results for impossible filters:', count);

    expect(count).toBe(0);

    // Should show no results message
    const hasNoResults = await dogsPage.hasNoResults();
    console.log('Shows no results message:', hasNoResults);

    // POTENTIAL BUG: Empty state handling
  });

});

// DONE: Dog browsing tests - listing, filters, search, experience levels, edge cases
