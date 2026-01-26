import { test, expect, type Page, type APIRequestContext } from "@playwright/test";
import {
  startTwoPlayerGame,
  completeSetupPhase,
  grantResources,
  grantResourcesAndWait,
  forceDiceRoll,
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
  async function enterBuildPhaseWithForcedRoll(
    page: Page,
    request: APIRequestContext,
    gameCode: string,
    diceValue: number = 8
  ) {
    await forceDiceRoll(request, gameCode, diceValue);
    await expect(page.locator("[data-cy='dice-result']")).toBeVisible({
      timeout: 10000,
    });
    const buildPhaseBtn = page.locator("[data-cy='build-phase-btn']");
    if (await buildPhaseBtn.isEnabled()) {
      await buildPhaseBtn.click();
    }
  }

  function parseRoadLength(text: string | null): number {
    const match = text?.match(/(\d+)/);
    return match ? Number.parseInt(match[1], 10) : 0;
  }

  async function expectRoadLengthAtLeast(
    page: Page,
    playerId: string,
    minimum: number
  ) {
    await expect(async () => {
      const text = await page
        .locator(`[data-cy='road-length-${playerId}']`)
        .textContent();
      expect(parseRoadLength(text)).toBeGreaterThanOrEqual(minimum);
    }).toPass({ timeout: 10000 });
  }

  function parseEdgeDataCy(dataCy: string | null): [string, string] | null {
    if (!dataCy) return null;
    const match = dataCy.match(
      /^edge-(-?\d+(?:\.\d+)?)-(-?\d+(?:\.\d+)?)-(-?\d+(?:\.\d+)?)-(-?\d+(?:\.\d+)?)$/
    );
    if (!match) {
      return null;
    }
    return [`${match[1]},${match[2]}`, `${match[3]},${match[4]}`];
  }

  async function buildConnectedRoads(page: Page, count: number) {
    const vertexCounts = new Map<string, number>();

    const bumpVertex = (vertex: string) => {
      vertexCounts.set(vertex, (vertexCounts.get(vertex) ?? 0) + 1);
    };

    const getEndpoints = () =>
      new Set(
        Array.from(vertexCounts.entries())
          .filter(([, count]) => count === 1)
          .map(([vertex]) => vertex)
      );

    for (let i = 0; i < count; i++) {
      await page.locator("[data-cy='build-road-btn']").click();

      const validEdges = page.locator("[data-cy^='edge-'].edge--valid");
      await expect(validEdges.first()).toBeVisible({ timeout: 5000 });

      const edgeData = await validEdges.evaluateAll((elements) =>
        elements.map((el) => el.getAttribute("data-cy"))
      );

      let chosenDataCy: string | null = null;
      const endpoints = getEndpoints();

      if (endpoints.size > 0) {
        for (const candidate of edgeData) {
          const vertices = parseEdgeDataCy(candidate);
          if (!vertices) continue;
          const [v1, v2] = vertices;
          const v1IsEndpoint = endpoints.has(v1);
          const v2IsEndpoint = endpoints.has(v2);
          const v1Seen = vertexCounts.has(v1);
          const v2Seen = vertexCounts.has(v2);

          if (v1IsEndpoint && !v2Seen) {
            chosenDataCy = candidate;
            break;
          }
          if (v2IsEndpoint && !v1Seen) {
            chosenDataCy = candidate;
            break;
          }
        }
      }

      if (!chosenDataCy && endpoints.size > 0) {
        for (const candidate of edgeData) {
          const vertices = parseEdgeDataCy(candidate);
          if (!vertices) continue;
          const [v1, v2] = vertices;
          if (endpoints.has(v1) || endpoints.has(v2)) {
            chosenDataCy = candidate;
            break;
          }
        }
      }

      if (!chosenDataCy) {
        chosenDataCy = edgeData[0] ?? null;
      }

      if (!chosenDataCy) {
        throw new Error("No valid edge found for road placement");
      }

      const targetEdge = page.locator(`[data-cy='${chosenDataCy}']`);
      await targetEdge.click();
      const placedVertices = parseEdgeDataCy(chosenDataCy);
      if (placedVertices) {
        bumpVertex(placedVertices[0]);
        bumpVertex(placedVertices[1]);
      }

      await page.waitForTimeout(500);
    }
  }

  test("5 connected roads awards Longest Road bonus", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession } =
      await startTwoPlayerGame(page, context, request);

    // Complete setup phase (each player places 2 settlements + 2 roads)
    await completeSetupPhase(hostPage, guestPage);

    // Grant resources to host for building 4 more roads (need 5 connected for bonus)
    // Setup phase gives 1-length chain, need 4 more = 4 wood + 4 brick
    await grantResourcesAndWait(
      request,
      hostPage,
      hostSession.code,
      hostSession.playerId,
      {
        wood: 8,
        brick: 8,
      }
    );

    await enterBuildPhaseWithForcedRoll(hostPage, request, hostSession.code);

    // Build additional roads until longest road length reaches 5
    await buildConnectedRoads(hostPage, 6);
    await expectRoadLengthAtLeast(hostPage, hostSession.playerId, 5);

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

    // Host builds 4 roads to get 5 connected (earns Longest Road)
    await grantResourcesAndWait(
      request,
      hostPage,
      hostSession.code,
      hostSession.playerId,
      {
        wood: 8,
        brick: 8,
      }
    );
    await hostPage.waitForTimeout(500);

    await enterBuildPhaseWithForcedRoll(hostPage, request, hostSession.code);
    await buildConnectedRoads(hostPage, 6);
    await expectRoadLengthAtLeast(hostPage, hostSession.playerId, 5);

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

    await enterBuildPhaseWithForcedRoll(guestPage, request, hostSession.code);

    // Guest builds more roads to exceed host's longest road
    await grantResourcesAndWait(
      request,
      guestPage,
      guestSession.code,
      guestSession.playerId,
      {
        wood: 10,
        brick: 10,
      }
    );
    await guestPage.waitForTimeout(500);

    await buildConnectedRoads(guestPage, 7);
    await expectRoadLengthAtLeast(guestPage, guestSession.playerId, 6);

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

    // Host builds 4 roads to get 5 connected (earns Longest Road)
    await grantResourcesAndWait(
      request,
      hostPage,
      hostSession.code,
      hostSession.playerId,
      {
        wood: 8,
        brick: 8,
      }
    );
    await hostPage.waitForTimeout(500);

    await enterBuildPhaseWithForcedRoll(hostPage, request, hostSession.code);
    await buildConnectedRoads(hostPage, 6);
    await expectRoadLengthAtLeast(hostPage, hostSession.playerId, 5);

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

    await enterBuildPhaseWithForcedRoll(guestPage, request, hostSession.code);

    // Guest builds enough roads to tie with host
    await grantResourcesAndWait(
      request,
      guestPage,
      guestSession.code,
      guestSession.playerId,
      {
        wood: 8,
        brick: 8,
      }
    );
    await guestPage.waitForTimeout(500);

    await buildConnectedRoads(guestPage, 6);
    await expectRoadLengthAtLeast(guestPage, guestSession.playerId, 5);

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

    // Host builds 4 roads to get 5 connected
    await grantResourcesAndWait(
      request,
      hostPage,
      hostSession.code,
      hostSession.playerId,
      {
        wood: 8,
        brick: 8,
      }
    );
    await hostPage.waitForTimeout(500);

    await enterBuildPhaseWithForcedRoll(hostPage, request, hostSession.code);
    await buildConnectedRoads(hostPage, 6);
    await expectRoadLengthAtLeast(hostPage, hostSession.playerId, 5);

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

    // Verify initial road counts (longest path is 1 from setup phase)
    await expect(
      hostPage.locator(`[data-cy='road-length-${hostSession.playerId}']`)
    ).toContainText("1");
    await expect(
      hostPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toContainText("1");

    // Host builds 1 more road
    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 1,
      brick: 1,
    });
    await hostPage.waitForTimeout(500);

    await enterBuildPhaseWithForcedRoll(hostPage, request, hostSession.code);
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
    ).toContainText("2");

    // Guest's road count should remain unchanged
    await expect(
      hostPage.locator(`[data-cy='road-length-${guestSession.playerId}']`)
    ).toContainText("1");
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

    // After setup, longest road length is 1 (below threshold)
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

    // Host builds 4 roads to get 5 connected
    await grantResourcesAndWait(
      request,
      hostPage,
      hostSession.code,
      hostSession.playerId,
      {
        wood: 8,
        brick: 8,
      }
    );
    await hostPage.waitForTimeout(500);

    await enterBuildPhaseWithForcedRoll(hostPage, request, hostSession.code);
    await buildConnectedRoads(hostPage, 6);
    await expectRoadLengthAtLeast(hostPage, hostSession.playerId, 5);

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
