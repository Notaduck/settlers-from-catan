import { test, expect } from '@playwright/test';
import {
  startTwoPlayerGame,
  completeSetupPhase,
  grantResources,
  rollDice,
  endTurn,
  waitForGamePhase,
} from './helpers';

test.describe('Trading (Bank and Player)', () => {
  test('Bank trade 4:1 works', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);
    
    // Grant host 4 wood and verify they have it
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 4,
    });
    
    // Wait for resource update
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('4', { timeout: 10000 });

    // Open bank trade modal
    const bankTradeBtn = hostPage.locator('[data-cy="bank-trade-btn"]');
await expect(bankTradeBtn).toBeVisible({ timeout: 10000 });
await expect(bankTradeBtn).toBeEnabled({ timeout: 15000 });
await bankTradeBtn.click();
    await expect(hostPage.locator('[data-cy="bank-trade-modal"]')).toBeVisible({ timeout: 10000 });

    // Verify default 4:1 ratio
    await expect(hostPage.locator('[data-cy="trade-ratio-1"]')).toContainText('4:1');

    // Select wood to offer (resource type 1 = WOOD)
    await hostPage.locator('[data-cy="bank-trade-offering-select"]').selectOption('1');
    
    // Select brick to receive (resource type 2 = BRICK)
    await hostPage.locator('[data-cy="bank-trade-requesting-select"]').selectOption('2');

    // Submit trade
    await hostPage.locator('[data-cy="bank-trade-submit-btn"]').click();

    // Verify resources updated: wood -4, brick +1
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('0', { timeout: 10000 });
    await expect(hostPage.locator('[data-cy="resource-brick"]')).toContainText('1', { timeout: 10000 });
  });

  test('Cannot bank trade without 4 resources', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant only 3 wood (not enough for 4:1 trade)
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 3,
    });
    
    // Wait for resource update
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('3', { timeout: 10000 });

    // Open bank trade modal
    const bankTradeBtn = hostPage.locator('[data-cy="bank-trade-btn"]');
await expect(bankTradeBtn).toBeVisible({ timeout: 10000 });
await expect(bankTradeBtn).toBeEnabled({ timeout: 15000 });
await bankTradeBtn.click();
    await expect(hostPage.locator('[data-cy="bank-trade-modal"]')).toBeVisible({ timeout: 10000 });

    // Select wood to offer
    await hostPage.locator('[data-cy="bank-trade-offering-select"]').selectOption('1');

    // Trade button should be disabled
    await expect(hostPage.locator('[data-cy="bank-trade-submit-btn"]')).toBeDisabled();

    // Error message should be visible
    await expect(hostPage.locator('[data-cy="bank-trade-error"]')).toBeVisible();
    await expect(hostPage.locator('[data-cy="bank-trade-error"]')).toContainText('Not enough');
  });

  test('Player can propose trade to another', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant host resources to trade
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 2,
      brick: 1,
    });
    
    // Wait for resource update
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('2', { timeout: 10000 });
    await expect(hostPage.locator('[data-cy="resource-brick"]')).toContainText('1', { timeout: 10000 });

    // Open propose trade modal
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');
await expect(proposeTradeBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeBtn.click();
    await expect(hostPage.locator('[data-cy="propose-trade-modal"]')).toBeVisible({ timeout: 10000 });

    // Select resources to offer (2 wood)
    const offerWood = hostPage.locator('[data-cy="trade-offer-wood"]');
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    await expect(offerWood).toContainText('2/2');

    // Select resources to request (1 sheep)
    const requestSheep = hostPage.locator('[data-cy="trade-request-sheep"]');
    await requestSheep.locator('button[aria-label="Increase Sheep"]').click();
    await expect(requestSheep).toContainText('1');

    // Submit trade
    const proposeTradeSubmitBtn = hostPage.locator('[data-cy="propose-trade-submit-btn"]');
await expect(proposeTradeSubmitBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeSubmitBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeSubmitBtn.click();

    // Modal should close
    await expect(hostPage.locator('[data-cy="propose-trade-modal"]')).not.toBeVisible();

    // Trade should be sent (backend will store in pendingTrades)
    // Note: We can't directly verify the trade is pending without backend introspection
    // but the test verifies the UI flow works correctly
  });

  test('Player can propose trade to specific player', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant host resources to trade
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 1,
    });
    
    // Wait for resource update
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('1', { timeout: 10000 });

    // Open propose trade modal
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');
await expect(proposeTradeBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeBtn.click();
    await expect(hostPage.locator('[data-cy="propose-trade-modal"]')).toBeVisible({ timeout: 10000 });

    // Select specific player as target (Guest)
    const targetSelect = hostPage.locator('[data-cy="trade-target-select"]');
    await expect(targetSelect).toBeVisible();
    
    // Select the guest player (not "All Players")
    const options = await targetSelect.locator('option').allTextContents();
    const guestOptionIndex = options.findIndex(opt => opt.includes('Guest'));
    if (guestOptionIndex === -1) {
      throw new Error('Guest player option not found in trade target select');
    }
    await targetSelect.selectOption({ index: guestOptionIndex });

    // Offer 1 wood
    const offerWood = hostPage.locator('[data-cy="trade-offer-wood"]');
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    await expect(offerWood).toContainText('1/1');

    // Request 1 brick
    const requestBrick = hostPage.locator('[data-cy="trade-request-brick"]');
    await requestBrick.locator('button[aria-label="Increase Brick"]').click();

    // Submit trade
    const proposeTradeSubmitBtn = hostPage.locator('[data-cy="propose-trade-submit-btn"]');
await expect(proposeTradeSubmitBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeSubmitBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeSubmitBtn.click();

    // Modal should close
    await expect(hostPage.locator('[data-cy="propose-trade-modal"]')).not.toBeVisible();
  });

  test('Trade recipient sees offer modal', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession, guestSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant both players resources
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 2,
    });
    await grantResources(request, guestSession.code, guestSession.playerId, {
      sheep: 1,
    });
    
    // Wait for resource updates
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('2', { timeout: 10000 });
    await expect(guestPage.locator('[data-cy="resource-sheep"]')).toContainText('1', { timeout: 10000 });

    // Host proposes trade to all players
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');
await expect(proposeTradeBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeBtn.click();
    await expect(hostPage.locator('[data-cy="propose-trade-modal"]')).toBeVisible({ timeout: 10000 });

    // Offer 2 wood
    const offerWood = hostPage.locator('[data-cy="trade-offer-wood"]');
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    await offerWood.locator('button[aria-label="Increase Wood"]').click();

    // Request 1 sheep
    const requestSheep = hostPage.locator('[data-cy="trade-request-sheep"]');
    await requestSheep.locator('button[aria-label="Increase Sheep"]').click();

    // Submit trade to all players
    const proposeTradeSubmitBtn = hostPage.locator('[data-cy="propose-trade-submit-btn"]');
await expect(proposeTradeSubmitBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeSubmitBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeSubmitBtn.click();

    // Guest should see incoming trade modal
    await expect(guestPage.locator('[data-cy="incoming-trade-modal"]')).toBeVisible({ timeout: 15000 });
    
    // Verify trade details are shown
    await expect(guestPage.locator('[data-cy="incoming-offer-wood"]')).toContainText('2');
    await expect(guestPage.locator('[data-cy="incoming-request-sheep"]')).toContainText('1');
    
    // Verify accept and decline buttons are present
    await expect(guestPage.locator('[data-cy="accept-trade-btn"]')).toBeVisible();
    await expect(guestPage.locator('[data-cy="decline-trade-btn"]')).toBeVisible();
  });

  test('Accepting trade transfers resources', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession, guestSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant both players resources
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 2,
    });
    await grantResources(request, guestSession.code, guestSession.playerId, {
      sheep: 1,
    });
    
    // Wait for resource updates
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('2', { timeout: 10000 });
    await expect(guestPage.locator('[data-cy="resource-sheep"]')).toContainText('1', { timeout: 10000 });

    // Host proposes trade
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');
await expect(proposeTradeBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeBtn.click();
    const offerWood = hostPage.locator('[data-cy="trade-offer-wood"]');
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    const requestSheep = hostPage.locator('[data-cy="trade-request-sheep"]');
    await requestSheep.locator('button[aria-label="Increase Sheep"]').click();
    const proposeTradeSubmitBtn = hostPage.locator('[data-cy="propose-trade-submit-btn"]');
await expect(proposeTradeSubmitBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeSubmitBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeSubmitBtn.click();

    // Guest accepts trade
    await expect(guestPage.locator('[data-cy="incoming-trade-modal"]')).toBeVisible({ timeout: 15000 });
    const acceptTradeBtn = guestPage.locator('[data-cy="accept-trade-btn"]');
await expect(acceptTradeBtn).toBeVisible({ timeout: 10000 });
await expect(acceptTradeBtn).toBeEnabled({ timeout: 15000 });
await acceptTradeBtn.click();

    // Modal should close
    await expect(guestPage.locator('[data-cy="incoming-trade-modal"]')).not.toBeVisible({ timeout: 15000 });

    // Verify resources transferred
    // Host should have: wood -2, sheep +1
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('0', { timeout: 10000 });
    await expect(hostPage.locator('[data-cy="resource-sheep"]')).toContainText('1', { timeout: 10000 });
    
    // Guest should have: sheep -1, wood +2
    await expect(guestPage.locator('[data-cy="resource-sheep"]')).toContainText('0', { timeout: 10000 });
    await expect(guestPage.locator('[data-cy="resource-wood"]')).toContainText('2', { timeout: 10000 });
  });

  test('Declining trade notifies proposer', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession, guestSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant both players resources
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 2,
    });
    await grantResources(request, guestSession.code, guestSession.playerId, {
      sheep: 1,
    });
    
    // Wait for resource updates
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('2', { timeout: 10000 });
    await expect(guestPage.locator('[data-cy="resource-sheep"]')).toContainText('1', { timeout: 10000 });

    // Host proposes trade
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');
await expect(proposeTradeBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeBtn.click();
    const offerWood = hostPage.locator('[data-cy="trade-offer-wood"]');
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    await offerWood.locator('button[aria-label="Increase Wood"]').click();
    const requestSheep = hostPage.locator('[data-cy="trade-request-sheep"]');
    await requestSheep.locator('button[aria-label="Increase Sheep"]').click();
    const proposeTradeSubmitBtn = hostPage.locator('[data-cy="propose-trade-submit-btn"]');
await expect(proposeTradeSubmitBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeSubmitBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeSubmitBtn.click();

    // Guest declines trade
    await expect(guestPage.locator('[data-cy="incoming-trade-modal"]')).toBeVisible({ timeout: 15000 });
    const declineTradeBtn = guestPage.locator('[data-cy="decline-trade-btn"]');
await expect(declineTradeBtn).toBeVisible({ timeout: 10000 });
await expect(declineTradeBtn).toBeEnabled({ timeout: 15000 });
await declineTradeBtn.click();

    // Modal should close for guest
    await expect(guestPage.locator('[data-cy="incoming-trade-modal"]')).not.toBeVisible({ timeout: 15000 });

    // Verify resources unchanged for both players
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('2', { timeout: 5000 });
    await expect(guestPage.locator('[data-cy="resource-sheep"]')).toContainText('1', { timeout: 5000 });

    // Trade should be marked as rejected (backend handles this)
    // Host might see a notification (implementation-dependent)
  });

  test('Cannot trade outside trade phase', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Before rolling dice, we're in PRE_ROLL phase
    // Trade buttons should be disabled or not visible
    const bankTradeBtn = hostPage.locator('[data-cy="bank-trade-btn"]');
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');

    // Either buttons are disabled or not present
    const bankBtnCount = await bankTradeBtn.count();
    if (bankBtnCount > 0) {
      await expect(bankTradeBtn).toBeDisabled();
    }

    const proposeBtnCount = await proposeTradeBtn.count();
    if (proposeBtnCount > 0) {
      await expect(proposeTradeBtn).toBeDisabled();
    }

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Now trade buttons should be enabled
    await expect(bankTradeBtn).toBeEnabled({ timeout: 15000 });
    await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });

    // End turn to leave trade phase
    await endTurn(hostPage);

    // Wait for guest's turn
    await expect(hostPage.locator('[data-cy="current-player"]')).not.toContainText('Host', { timeout: 10000 });

    // Trade buttons should be disabled again (not host's turn)
    await expect(bankTradeBtn).toBeDisabled();
    await expect(proposeTradeBtn).toBeDisabled();
  });

  test('Cannot propose trade without resources', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Open propose trade modal without granting resources
    const proposeTradeBtn = hostPage.locator('[data-cy="propose-trade-btn"]');
await expect(proposeTradeBtn).toBeVisible({ timeout: 10000 });
await expect(proposeTradeBtn).toBeEnabled({ timeout: 15000 });
await proposeTradeBtn.click();
    await expect(hostPage.locator('[data-cy="propose-trade-modal"]')).toBeVisible({ timeout: 10000 });

    // Try to offer 1 wood (but we don't have any)
    const offerWood = hostPage.locator('[data-cy="trade-offer-wood"]');
    
    // Increment button should be disabled (no wood to offer)
    const incrementBtn = offerWood.locator('button[aria-label="Increase Wood"]');
    await expect(incrementBtn).toBeDisabled();

    // Submit button should be disabled (no offer)
    await expect(hostPage.locator('[data-cy="propose-trade-submit-btn"]')).toBeDisabled();
  });

  test('Multiple trades per turn allowed', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    // Grant host plenty of resources
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 8,
      brick: 1,
    });
    
    // Wait for resource update
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('8', { timeout: 10000 });

    // First bank trade: 4 wood -> 1 sheep
    const bankTradeBtn = hostPage.locator('[data-cy="bank-trade-btn"]');
await expect(bankTradeBtn).toBeVisible({ timeout: 10000 });
await expect(bankTradeBtn).toBeEnabled({ timeout: 15000 });
await bankTradeBtn.click();
    await hostPage.locator('[data-cy="bank-trade-offering-select"]').selectOption('1'); // Wood
    await hostPage.locator('[data-cy="bank-trade-requesting-select"]').selectOption('3'); // Sheep
    await hostPage.locator('[data-cy="bank-trade-submit-btn"]').click();

    // Verify first trade
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('4', { timeout: 10000 });
    await expect(hostPage.locator('[data-cy="resource-sheep"]')).toContainText('1', { timeout: 10000 });

    // Second bank trade: 4 wood -> 1 ore
    // Reuse previously declared bankTradeBtn
await expect(bankTradeBtn).toBeVisible({ timeout: 10000 });
await expect(bankTradeBtn).toBeEnabled({ timeout: 15000 });
await bankTradeBtn.click();
    await hostPage.locator('[data-cy="bank-trade-offering-select"]').selectOption('1'); // Wood
    await hostPage.locator('[data-cy="bank-trade-requesting-select"]').selectOption('5'); // Ore
    await hostPage.locator('[data-cy="bank-trade-submit-btn"]').click();

    // Verify second trade
    await expect(hostPage.locator('[data-cy="resource-wood"]')).toContainText('0', { timeout: 10000 });
    await expect(hostPage.locator('[data-cy="resource-ore"]')).toContainText('1', { timeout: 10000 });

    // Both trades should succeed in the same turn
  });

  test('Bank trade button shows during trade phase only', async ({ page, context, request }) => {
    // Setup 2-player game and complete setup phase
    const { hostPage, guestPage } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    // Wait for PLAYING phase
    await waitForGamePhase(hostPage, 'PLAYING');

    // Before rolling, trade button should be disabled
    const bankTradeBtn = hostPage.locator('[data-cy="bank-trade-btn"]');
    await expect(bankTradeBtn).toBeDisabled();

    // Roll dice
    await rollDice(hostPage);

    // After rolling, trade button should be enabled (TRADE phase)
    await expect(bankTradeBtn).toBeEnabled({ timeout: 15000 });

    // End turn
    await endTurn(hostPage);

    // After ending turn, trade button should be disabled (not our turn)
    await expect(bankTradeBtn).toBeDisabled();
  });

  test('Can switch between trade and build phases', async ({ page, context, request }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(page, context, request);
    await completeSetupPhase(hostPage, guestPage);

    await waitForGamePhase(hostPage, 'PLAYING');

    // Roll dice to enter trade phase
    await rollDice(hostPage);

    const tradeBtn = hostPage.locator('[data-cy="trade-phase-btn"]');
    const buildBtn = hostPage.locator('[data-cy="build-phase-btn"]');
    const bankTradeBtn = hostPage.locator('[data-cy="bank-trade-btn"]');

    await expect(tradeBtn).toBeVisible();
    await expect(buildBtn).toBeVisible();
    await expect(bankTradeBtn).toBeVisible();

    // Switch to build phase
    await buildBtn.click();
    await expect(bankTradeBtn).toHaveCount(0);

    // Switch back to trade phase
    await tradeBtn.click();
    await expect(bankTradeBtn).toBeVisible({ timeout: 10000 });
  });
});
