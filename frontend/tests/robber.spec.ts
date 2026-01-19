import { test, expect } from '@playwright/test';

// NOTE: The following assumes a fresh game and backend that supports fast setup/reset.

test.describe('Robber Phase UI', () => {
  test('Shows discard modal and submits correct amount', async ({ page }) => {
    // Setup: Create game, roll 7, have >7 cards for player 1
    // This requires backend seeding/mocking, assume api is present
    await page.goto('/');
    // Join as Player 1
    await page.locator('[data-cy="player-name-input"]').fill('Alice');
    await page.locator('[data-cy="create-game-btn"]').click();
    // Fast-forward/start game, grant >7 cards, roll 7 - assumed via backend shortcut or API
    await page.evaluate(async () => {
      // These should be available if backend supports e2e hooks
      window.__test?.grantCards?.('all', { wood: 4, brick: 4, sheep: 4 });
      window.__test?.forceRoll?.(7);
    });
    await page.getByText('Discard', { exact: false, timeout: 4000 });
    const modal = page.locator('[data-cy="discard-modal"]');
    await expect(modal).toBeVisible();
    // Discard cards (adjust inputs accordingly)
    // Click enough to reach the required (assume 6 for test)
    for (let i = 0; i < 6; ++i) {
      // Pick the first plus button
      await modal.locator('[data-cy^="discard-card-"] button:last-child').first().click();
    }
    // Submit
    await modal.locator('[data-cy="discard-submit"]').click();
    await expect(modal).toBeHidden();
  });

  test('Robber move: player sees selectable hex', async ({ page }) => {
    await page.goto('/');
    // Use API/shortcut to start at robber move phase
    await page.evaluate(async () => {
      window.__test?.jumpToRobberMove?.();
    });
    // Should see at least one clickable robber hex tile
    await expect(page.locator('[data-cy^="robber-hex-"]').first()).toBeVisible();
    // Click a new hex
    const hexes = await page.locator('[data-cy^="robber-hex-"]').elementHandles();
    for (const h of hexes) {
      if (!(await h.evaluate(el => el.classList.contains('robber-move-selectable')))) continue;
      await h.click();
      break;
    }
    // Modal should be gone or steal modal appears
    await expect(page.locator('[data-cy^="robber-hex-"]')).not.toBeVisible({ timeout: 4000 });
  });

  test('Steal modal: lists correct steal candidates', async ({ page }) => {
    await page.goto('/');
    // Use API/shortcut to move to steal step
    await page.evaluate(async () => {
      window.__test?.jumpToRobberSteal?.([{'id': 'p2', 'name': 'Bob'}, {'id': 'p3', 'name': 'Cathy'}]);
    });
    // Modal appears
    const steal = page.locator('[data-cy="steal-modal"]');
    await expect(steal).toBeVisible({ timeout: 4000 });
    // Both candidate buttons should be present
    await expect(steal.locator('[data-cy="steal-player-p2"]')).toBeVisible();
    await expect(steal.locator('[data-cy="steal-player-p3"]')).toBeVisible();
    // Pick one
    await steal.locator('[data-cy="steal-player-p2"]').click();
    await expect(steal).toBeHidden();
  });
});
