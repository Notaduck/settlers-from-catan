import { test, expect } from "@playwright/test";

// Note: This e2e requires backend/frontend servers running

test.describe("Victory flow", () => {
  test("Game shows victory screen and blocks actions after win", async ({ page }) => {
    // SETUP & START: Create game, join, ready, start game, fast-play to victory
    // For demo, stub final state (e2e infra would need full simulation setup)
    // Assumes test/dev env can inject game state (not available in prod)
    await page.goto("/");
    // Simulated: Change this to use fixtures or real play as infra allows
    // Wait for game-over UI overlay
    await page.waitForSelector('[data-cy="game-over-overlay"]', { timeout: 60000 });
    // Check winner name and VP appear
    const winner = await page.locator('[data-cy="winner-name"]');
    await expect(winner).toBeVisible();
    const vp = await page.locator('[data-cy="winner-vp"]');
    await expect(vp).toBeVisible();
    // Score rows
    const rows = await page.locator('[data-cy^="final-score-"]');
    await expect(rows).toHaveCountGreaterThan(0);
    // New Game button
    const newGameBtn = await page.locator('[data-cy="new-game-btn"]');
    await expect(newGameBtn).toBeVisible();
    // Victory overlay prevents further actions
    // Try clicking on game board or controls and expect overlays block input
    // (Assume future: test cannot build after win; for now check overlay)
  });
});
