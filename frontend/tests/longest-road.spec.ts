import { test, expect } from "@playwright/test";
import {
  startTwoPlayerGame,
  completeSetupPhase,
  grantResources,
} from "./helpers";

/**
 * Longest Road E2E Tests
 *
 * Tests the Longest Road bonus feature (2 VP for player with 5+ connected roads).
 *
 * Backend implementation: backend/internal/game/longestroad.go
 * Spec: specs/longest-road.md
 */

test.describe("Longest Road", () => {
  test("5 connected roads awards Longest Road bonus", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession } =
      await startTwoPlayerGame(page, context, request);

    // Complete setup phase (each player places 2 settlements + 2 roads)
    await completeSetupPhase(hostPage, guestPage);

    // Grant resources to host for building 3 more roads (need 5 total for bonus)
    // Setup phase gives 2 roads, need 3 more = 3 wood + 3 brick
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 3,
      brick: 3,
    });

    // Wait for resource grant to propagate
    await expect(
      hostPage.locator("[data-cy='resource-wood']")
    ).toContainText("3", { timeout: 10000 });

    // Build 3 additional roads to reach 5 total
    // Note: Need to find valid edge placements adjacent to existing roads
    const validEdges = hostPage.locator("[data-cy^='edge-'].edge--valid");
    await expect(validEdges.first()).toBeVisible({ timeout: 10000 });

    for (let i = 0; i < 3; i++) {
      await hostPage.locator("[data-cy='build-road-btn']").click();
      await expect(validEdges.first()).toBeVisible({ timeout: 5000 });
      await validEdges.first().click();
      
      // Wait for road to be placed
      await hostPage.waitForTimeout(500);
    }

    // Verify Longest Road badge appears for host
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).toBeVisible({ timeout: 10000 });
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(hostSession.playerId);

    // Verify road length is displayed
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toBeVisible();
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toContainText("5"); // Should show at least 5 roads

    // Verify host's VP increased by 2 (Longest Road bonus)
    const hostVP = hostPage.locator(
      `[data-cy='player-vp-${hostSession.playerId}']`
    );
    await expect(hostVP).toContainText("4 VP"); // 2 settlements + 2 VP bonus
  });

  test("Longer road takes Longest Road from another player", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession, guestSession } =
      await startTwoPlayerGame(page, context, request);

    await completeSetupPhase(hostPage, guestPage);

    // Host builds 3 roads to get 5 total (earns Longest Road)
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 3,
      brick: 3,
    });
    await hostPage.waitForTimeout(500);

    for (let i = 0; i < 3; i++) {
      await hostPage.locator("[data-cy='build-road-btn']").click();
      const validEdge = hostPage
        .locator("[data-cy^='edge-'].edge--valid")
        .first();
      await expect(validEdge).toBeVisible({ timeout: 5000 });
      await validEdge.click();
      await hostPage.waitForTimeout(500);
    }

    // Verify host has Longest Road
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(hostSession.playerId, { timeout: 10000 });

    // End host's turn
    await hostPage.locator("[data-cy='end-turn-btn']").click();

    // Wait for guest's turn
    await expect(
      guestPage.locator("[data-cy='current-player-name']")
    ).toContainText("Guest", { timeout: 10000 });

    // Guest rolls dice
    await guestPage.locator("[data-cy='roll-dice-btn']").click();
    await guestPage.waitForTimeout(1000);

    // Guest builds 4 roads to get 6 total (should take Longest Road)
    await grantResources(request, guestSession.code, guestSession.playerId, {
      wood: 4,
      brick: 4,
    });
    await guestPage.waitForTimeout(500);

    for (let i = 0; i < 4; i++) {
      await guestPage.locator("[data-cy='build-road-btn']").click();
      const validEdge = guestPage
        .locator("[data-cy^='edge-'].edge--valid")
        .first();
      await expect(validEdge).toBeVisible({ timeout: 5000 });
      await validEdge.click();
      await guestPage.waitForTimeout(500);
    }

    // Verify Longest Road transferred to guest
    await expect(
      guestPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(guestSession.playerId, { timeout: 10000 });

    // Verify guest's road length
    await expect(
      guestPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toContainText("6");

    // Verify guest's VP increased by 2
    const guestVP = guestPage.locator(
      `[data-cy='player-vp-${guestSession.playerId}']`
    );
    await expect(guestVP).toContainText("4 VP"); // 2 settlements + 2 VP bonus

    // Verify host lost the bonus
    const hostVP = hostPage.locator(
      `[data-cy='player-vp-${hostSession.playerId}']`
    );
    await expect(hostVP).toContainText("2 VP"); // 2 settlements only
  });

  test("Tie does not transfer Longest Road (current holder keeps)", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession, guestSession } =
      await startTwoPlayerGame(page, context, request);

    await completeSetupPhase(hostPage, guestPage);

    // Host builds 3 roads to get 5 total (earns Longest Road)
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 3,
      brick: 3,
    });
    await hostPage.waitForTimeout(500);

    for (let i = 0; i < 3; i++) {
      await hostPage.locator("[data-cy='build-road-btn']").click();
      const validEdge = hostPage
        .locator("[data-cy^='edge-'].edge--valid")
        .first();
      await expect(validEdge).toBeVisible({ timeout: 5000 });
      await validEdge.click();
      await hostPage.waitForTimeout(500);
    }

    // Verify host has Longest Road
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(hostSession.playerId, { timeout: 10000 });

    // End host's turn
    await hostPage.locator("[data-cy='end-turn-btn']").click();

    // Wait for guest's turn
    await expect(
      guestPage.locator("[data-cy='current-player-name']")
    ).toContainText("Guest", { timeout: 10000 });

    // Guest rolls dice
    await guestPage.locator("[data-cy='roll-dice-btn']").click();
    await guestPage.waitForTimeout(1000);

    // Guest builds 3 roads to also get 5 total (TIE)
    await grantResources(request, guestSession.code, guestSession.playerId, {
      wood: 3,
      brick: 3,
    });
    await guestPage.waitForTimeout(500);

    for (let i = 0; i < 3; i++) {
      await guestPage.locator("[data-cy='build-road-btn']").click();
      const validEdge = guestPage
        .locator("[data-cy^='edge-'].edge--valid")
        .first();
      await expect(validEdge).toBeVisible({ timeout: 5000 });
      await validEdge.click();
      await guestPage.waitForTimeout(500);
    }

    // Verify Longest Road stays with host (tie goes to current holder)
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(hostSession.playerId, { timeout: 10000 });

    // Verify both have same road length
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toContainText("5");
    await expect(
      guestPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toContainText("5");

    // Verify only host has VP bonus
    const hostVP = hostPage.locator(
      `[data-cy='player-vp-${hostSession.playerId}']`
    );
    await expect(hostVP).toContainText("4 VP"); // Still has bonus

    const guestVP = guestPage.locator(
      `[data-cy='player-vp-${guestSession.playerId}']`
    );
    await expect(guestVP).toContainText("2 VP"); // No bonus (tie)
  });

  test("Longest Road badge shows correct player name", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    await completeSetupPhase(hostPage, guestPage);

    // Host builds 3 roads to get 5 total
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 3,
      brick: 3,
    });
    await hostPage.waitForTimeout(500);

    for (let i = 0; i < 3; i++) {
      await hostPage.locator("[data-cy='build-road-btn']").click();
      const validEdge = hostPage
        .locator("[data-cy^='edge-'].edge--valid")
        .first();
      await expect(validEdge).toBeVisible({ timeout: 5000 });
      await validEdge.click();
      await hostPage.waitForTimeout(500);
    }

    // Verify Longest Road holder shows correct player name
    const longestRoadBadge = hostPage.locator("[data-cy='longest-road-holder']");
    await expect(longestRoadBadge).toBeVisible({ timeout: 10000 });

    // Should contain either "Host" (player name) or playerId
    // Implementation can decide which to display
    const badgeText = await longestRoadBadge.textContent();
    expect(
      badgeText?.includes("Host") || badgeText?.includes(hostSession.playerId)
    ).toBeTruthy();
  });

  test("Road length displayed for all players", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession, guestSession } =
      await startTwoPlayerGame(page, context, request);

    await completeSetupPhase(hostPage, guestPage);

    // Both players should have road length displayed (from setup phase)
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toBeVisible();
    await expect(
      hostPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toBeVisible();

    // Verify initial road counts (2 roads each from setup phase)
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toContainText("2");
    await expect(
      hostPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toContainText("2");

    // Host builds 1 more road
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 1,
      brick: 1,
    });
    await hostPage.waitForTimeout(500);

    await hostPage.locator("[data-cy='build-road-btn']").click();
    const validEdge = hostPage
      .locator("[data-cy^='edge-'].edge--valid")
      .first();
    await expect(validEdge).toBeVisible({ timeout: 5000 });
    await validEdge.click();
    await hostPage.waitForTimeout(500);

    // Verify host's road count updated
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toContainText("3");

    // Guest's road count should remain unchanged
    await expect(
      hostPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toContainText("2");
  });

  test("No Longest Road badge before 5 roads threshold", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    await completeSetupPhase(hostPage, guestPage);

    // After setup, each player has 2 roads (below threshold)
    // Verify no Longest Road badge is shown
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).not.toBeVisible();
  });

  test("Longest Road updates in real-time for all players", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    await completeSetupPhase(hostPage, guestPage);

    // Host builds 3 roads to get 5 total
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 3,
      brick: 3,
    });
    await hostPage.waitForTimeout(500);

    for (let i = 0; i < 3; i++) {
      await hostPage.locator("[data-cy='build-road-btn']").click();
      const validEdge = hostPage
        .locator("[data-cy^='edge-'].edge--valid")
        .first();
      await expect(validEdge).toBeVisible({ timeout: 5000 });
      await validEdge.click();
      await hostPage.waitForTimeout(500);
    }

    // Verify BOTH host and guest pages show the Longest Road update
    await expect(
      hostPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(hostSession.playerId, { timeout: 10000 });

    await expect(
      guestPage.locator("[data-cy='longest-road-holder']")
    ).toContainText(hostSession.playerId, { timeout: 10000 });

    // Verify VP update is visible to both players
    const hostVPonHost = hostPage.locator(
      `[data-cy='player-vp-${hostSession.playerId}']`
    );
    const hostVPonGuest = guestPage.locator(
      `[data-cy='player-vp-${hostSession.playerId}']`
    );

    await expect(hostVPonHost).toContainText("4 VP");
    await expect(hostVPonGuest).toContainText("4 VP");
  });
});
