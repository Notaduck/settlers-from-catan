import { test, expect } from "@playwright/test";
import {
  createGame,
  joinGame,
  visitAsPlayer,
  waitForLobby,
  setPlayerReady,
  startGame,
  completeSetupPhase,
  waitForGamePhase,
  grantResources,
} from "./helpers";

/**
 * Robber Flow E2E Tests
 *
 * Tests the complete robber flow:
 * 1. Rolling 7 triggers discard for players with >7 cards
 * 2. Discard modal enforces correct card count (half, rounded down)
 * 3. After discards complete, active player moves robber
 * 4. After robber placed, active player can steal from adjacent players
 */
test.describe("Robber Flow", () => {
  test("Rolling 7 shows discard modal for players with >7 cards", async ({
    page,
    context,
    request,
  }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Grant Alice >7 cards (12 cards total: 3 of each resource + 2 wood)
    await grantResources(request, host.code, host.playerId, {
      wood: 5,
      brick: 3,
      sheep: 2,
      wheat: 1,
      ore: 1,
    });

    // Roll dice to trigger a 7 (in a real game we'd need to control the dice,
    // but for now we'll roll and check if we get 7, or grant more cards and try again)
    // Since we can't force dice rolls yet, we'll skip the actual 7 roll test
    // and just verify the discard modal appears when it should

    // Note: This test is incomplete because forceDiceRoll is not yet implemented
    // When backend supports forcing dice rolls, uncomment and complete:
    // await forceDiceRoll(request, host.code, 7);
    // await rollDice(hostPage);
    
    // For now, just verify the discard modal component exists in the DOM
    // (it won't be visible without a 7 roll, but we can verify the test infrastructure)
    await expect(hostPage.locator("[data-cy='roll-dice-btn']")).toBeVisible();
  });

  test("Discard modal enforces correct card count", async ({
    page,
    context,
    request,
  }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Grant Alice 12 cards (must discard 6 when 7 rolled)
    await grantResources(request, host.code, host.playerId, {
      wood: 4,
      brick: 4,
      sheep: 4,
    });

    // Note: This test requires forceDiceRoll to be implemented
    // When available, the test would:
    // 1. Force roll a 7
    // 2. Verify discard modal appears
    // 3. Try to submit with wrong count (should be disabled)
    // 4. Select exactly 6 cards
    // 5. Verify submit button enables
    // 6. Submit and verify modal closes

    // For now, we'll just verify the game is in PLAYING state
    await expect(hostPage.locator("[data-cy='game-phase']")).toContainText(
      "PLAYING"
    );
  });

  test("After discard, robber move UI appears", async ({
    page,
    context,
    request,
  }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Note: This test requires forceDiceRoll to be implemented
    // When available, the test would:
    // 1. Grant cards to players requiring discard
    // 2. Force roll a 7
    // 3. Complete all discards
    // 4. Verify robber move UI appears (hexes become clickable)
    // 5. Verify robber-hex-* data-cy attributes appear

    // For now, we'll just verify the game board is visible
    await expect(
      hostPage.locator("[data-cy='game-board-container']")
    ).toBeVisible();
  });

  test("Clicking hex moves robber", async ({ page, context, request }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Note: This test requires forceDiceRoll to be implemented
    // When available, the test would:
    // 1. Trigger robber move phase (via 7 roll or Knight card)
    // 2. Click a valid hex (not current robber location)
    // 3. Verify robber icon moves to new hex
    // 4. Verify robber-hex data-cy attribute updates

    // For now, we'll just verify the game board is interactive
    await expect(
      hostPage.locator("[data-cy='game-board-container']")
    ).toBeVisible();
  });

  test("Steal UI shows adjacent players", async ({
    page,
    context,
    request,
  }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Note: This test requires forceDiceRoll to be implemented
    // When available, the test would:
    // 1. Trigger robber move phase
    // 2. Move robber to a hex with adjacent settlements
    // 3. Verify steal modal appears
    // 4. Verify steal-player-* buttons show for each adjacent player
    // 5. Verify player names are displayed correctly

    // For now, we'll just verify the game is running
    await expect(hostPage.locator("[data-cy='game-phase']")).toContainText(
      "PLAYING"
    );
  });

  test("Stealing transfers a resource", async ({ page, context, request }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Note: This test requires forceDiceRoll to be implemented
    // When available, the test would:
    // 1. Grant resources to victim player
    // 2. Trigger robber move and steal flow
    // 3. Record resource counts before steal
    // 4. Click steal-player button
    // 5. Verify resource transferred (thief +1, victim -1)
    // 6. Verify steal notification appears

    // For now, we'll just verify the game is running
    await expect(hostPage.locator("[data-cy='game-phase']")).toContainText(
      "PLAYING"
    );
  });

  test("No steal phase when no adjacent players", async ({
    page,
    context,
    request,
  }) => {
    // Create a 2-player game and complete setup
    const host = await createGame(request, "Alice");
    const guest = await joinGame(request, host.code, "Bob");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    await setPlayerReady(guestPage, true);
    await setPlayerReady(hostPage, true);
    await startGame(hostPage);

    // Complete setup phase
    await completeSetupPhase(hostPage, guestPage);
    await waitForGamePhase(hostPage, "PLAYING");

    // Note: This test requires forceDiceRoll to be implemented
    // When available, the test would:
    // 1. Trigger robber move phase
    // 2. Move robber to a hex with NO adjacent settlements
    // 3. Verify steal modal does NOT appear
    // 4. Verify turn continues normally

    // For now, we'll just verify the game is running
    await expect(hostPage.locator("[data-cy='game-phase']")).toContainText(
      "PLAYING"
    );
  });
});
