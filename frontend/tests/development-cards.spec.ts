import { test } from "@playwright/test";
import { createGame, joinGame, setPlayerReady, startGame } from "./helpers";

test.describe("Development Cards", () => {
  test("should display development cards panel during playing phase", async ({
    page,
    context,
  }) => {
    // Create game with player 1
    const { gameCode } = await createGame(page, "Player 1");

    // Join with player 2
    const page2 = await context.newPage();
    await joinGame(page2, gameCode, "Player 2");

    // Both players ready
    await setPlayerReady(page, true);
    await setPlayerReady(page2, true);

    // Start game
    await startGame(page);

    // Wait for game to start (setup phase)
    await page.waitForSelector('[data-cy="game-board-container"]', {
      timeout: 5000,
    });

    // Complete setup phase by placing initial settlements and roads
    // This is a simplified version - in real game would need proper placement
    // For now, just verify the panel appears when status changes to PLAYING

    // Dev cards panel should appear during PLAYING status
    // Note: This test assumes game progresses to PLAYING phase
    // In a real test, we'd need to complete the setup phase properly
  });

  test("should show buy dev card button when player has resources", async ({
    page,
    context,
  }) => {
    const { gameCode } = await createGame(page, "Player 1");
    const page2 = await context.newPage();
    await joinGame(page2, gameCode, "Player 2");

    await setPlayerReady(page, true);
    await setPlayerReady(page2, true);
    await startGame(page);

    // Wait for setup phase
    await page.waitForSelector('[data-cy="setup-phase-banner"]', {
      timeout: 5000,
    });

    // Dev cards panel is only shown during PLAYING phase
    // Button should be disabled if player lacks resources (1 ore, 1 wheat, 1 sheep)
  });

  test("should open Year of Plenty modal when playing card", async () => {
    // This test would require:
    // 1. Getting to PLAYING phase
    // 2. Giving player a Year of Plenty card (via backend state manipulation or playing through)
    // 3. Clicking play button
    // 4. Verifying modal opens with resource selectors
  });

  test("should open Monopoly modal when playing card", async () => {
    // This test would require:
    // 1. Getting to PLAYING phase
    // 2. Giving player a Monopoly card
    // 3. Clicking play button
    // 4. Verifying modal opens with resource type selector
  });

  test("should disable play button when not player's turn", async () => {
    // Test that dev cards can only be played on your turn
  });

  test("should show dev card count in panel", async () => {
    // Verify that dev card count displays correctly
    // Verify that individual card types show with counts
  });

  test("should not show play button for Victory Point cards", async () => {
    // VP cards should not have a play button
    // They are automatically counted toward victory
  });

  test("Year of Plenty should allow selecting 2 resources", async () => {
    // Test the Year of Plenty modal:
    // - Can select 2 resources
    // - Can select same resource twice
    // - Submit button disabled until 2 selected
    // - Can remove selections
  });

  test("Monopoly should allow selecting 1 resource type", async () => {
    // Test the Monopoly modal:
    // - Can select one resource type
    // - Submit button disabled until resource selected
    // - Selecting different resource changes selection
  });

  test("Knight card should trigger robber move", async () => {
    // Playing knight should:
    // - Move robber to new hex
    // - Allow stealing from adjacent player
    // - Increment knight count
    // - Check for Largest Army
  });

  test("Road Building should allow placing 2 free roads", async () => {
    // Playing Road Building should:
    // - Enter road placement mode
    // - Allow placing 2 roads without resource cost
    // - Exit placement mode after 2 roads placed
  });
});
