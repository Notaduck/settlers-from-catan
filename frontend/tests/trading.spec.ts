import { test, expect } from '@playwright/test';

// Trading E2E test spec as per specs/trading.md

test.describe('Trading (Bank and Player)', () => {
  test('Bank trade 4:1 works', async ({ page }) => {
    // Setup/join game, advance to TRADE phase
    // Open trade phase, open bank trade modal, offer 4 of a resource
    // Select a resource to receive, confirm trade
    // Assert updated resources
  });

  test('Cannot bank trade without 4 resources', async ({ page }) => {
    // Setup game, open bank trade modal, try to trade without 4 of any resource
    // Should be disabled or show error
  });

  test('Player can propose trade to another', async ({ page }) => {
    // Setup with 2+ players, advance to TRADE phase
    // Open propose trade modal, pick offers/requests, send
    // Assert recipient sees incoming trade modal
  });

  test('Trade recipient sees offer modal', async ({ page }) => {
    // As above, verify correct modal and buttons appear for recipient
  });

  test('Accepting trade transfers resources', async ({ page }) => {
    // Complete trade as recipient, resources update on both sides
  });

  test('Declining trade notifies proposer', async ({ page }) => {
    // Decline offer, proposer sees notification/trade disappears
  });

  test('Cannot trade outside trade phase', async ({ page }) => {
    // End trade phase, verify trade/bank buttons and modals are unavailable
  });
});
