import {
  type Page,
  type BrowserContext,
  type APIRequestContext,
  expect,
} from "@playwright/test";

const API_BASE = "http://localhost:8080";

export interface GameSession {
  code: string;
  sessionToken: string;
  playerId: string;
}

// ============================================================================
// Core Game Setup Helpers
// ============================================================================

/**
 * Create a new game via API
 */
export async function createGame(
  request: APIRequestContext,
  playerName: string
): Promise<GameSession> {
  const response = await request.post(`${API_BASE}/api/games`, {
    data: { playerName },
  });
  return response.json();
}

/**
 * Join an existing game via API
 */
export async function joinGame(
  request: APIRequestContext,
  gameCode: string,
  playerName: string
): Promise<GameSession> {
  const response = await request.post(
    `${API_BASE}/api/games/${gameCode}/join`,
    {
      data: { playerName },
    }
  );
  const body = await response.json();
  return { code: gameCode, ...body };
}

/**
 * Visit the game as a specific player (set sessionStorage and reload)
 */
export async function visitAsPlayer(page: Page, session: GameSession) {
  await page.goto("/");
  await page.evaluate((s) => {
    sessionStorage.setItem("sessionToken", s.sessionToken);
    sessionStorage.setItem("gameCode", s.code);
    sessionStorage.setItem("playerId", s.playerId);
    localStorage.removeItem("sessionToken");
    localStorage.removeItem("gameCode");
    localStorage.removeItem("playerId");
  }, session);
  await page.reload();
}

/**
 * Wait for lobby screen to appear
 */
export async function waitForLobby(page: Page) {
  await expect(page.locator("[data-cy='game-loading']")).not.toBeVisible({
    timeout: 30000,
  });
  await expect(page.locator("[data-cy='game-waiting']")).toBeVisible({
    timeout: 30000,
  });
}

/**
 * Wait for game board to appear (after game starts)
 */
export async function waitForGameBoard(page: Page) {
  await expect(page.locator("[data-cy='game-loading']")).not.toBeVisible({
    timeout: 30000,
  });
  await expect(page.locator("[data-cy='game-board-container']")).toBeVisible({
    timeout: 30000,
  });
}

/**
 * Mark a player as ready in the lobby
 */
export async function setPlayerReady(page: Page, ready: boolean = true) {
  if (ready) {
    await page.locator("[data-cy='ready-btn']").click();
    await expect(page.locator("[data-cy='cancel-ready-btn']")).toBeVisible({
      timeout: 10000,
    });
  } else {
    await page.locator("[data-cy='cancel-ready-btn']").click();
    await expect(page.locator("[data-cy='ready-btn']")).toBeVisible({
      timeout: 10000,
    });
  }
}

/**
 * Start the game (clicks start button)
 */
export async function startGame(page: Page) {
  await expect(page.locator("[data-cy='start-game-btn']")).toBeEnabled({
    timeout: 10000,
  });
  await page.locator("[data-cy='start-game-btn']").click();
  await waitForGameBoard(page);
}

/**
 * Start a 2-player game (creates game, joins as 2 players, readies up, starts)
 * Returns { hostPage, guestPage }
 */
export async function startTwoPlayerGame(
  page: Page,
  context: BrowserContext,
  request: APIRequestContext
) {
  const host = await createGame(request, "Host");
  const guest = await joinGame(request, host.code, "Guest");

  const hostPage = page;
  const guestPage = await context.newPage();

  await visitAsPlayer(hostPage, host);
  await waitForLobby(hostPage);

  await visitAsPlayer(guestPage, guest);
  await waitForLobby(guestPage);

  await setPlayerReady(guestPage, true);
  await setPlayerReady(hostPage, true);
  await startGame(hostPage);

  return { hostPage, guestPage, hostSession: host, guestSession: guest };
}

// ============================================================================
// Setup Phase Helpers
// ============================================================================

/**
 * Place a settlement during setup phase (clicks first valid vertex)
 */
export async function placeSettlement(page: Page): Promise<string> {
  const placementMode = page.locator("[data-cy='placement-mode']");
  await expect(placementMode).toContainText("Place Settlement", {
    timeout: 30000,
  });
  const validVertex = page.locator("[data-cy^='vertex-'].vertex--valid").first();
  await expect(validVertex).toBeVisible({ timeout: 30000 });
  const dataCy = (await validVertex.getAttribute("data-cy")) ?? "";
  await validVertex.click();
  return dataCy;
}

/**
 * Place a road during setup phase (clicks first valid edge)
 */
export async function placeRoad(page: Page) {
  const placementMode = page.locator("[data-cy='placement-mode']");
  await expect(placementMode).toContainText("Place Road", { timeout: 30000 });
  const validEdge = page.locator("[data-cy^='edge-'].edge--valid").first();
  await expect(validEdge).toBeVisible({ timeout: 30000 });
  await validEdge.click();
}

/**
 * Complete one round of setup (settlement + road for current player)
 */
export async function completeSetupRound(page: Page) {
  await placeSettlement(page);
  await placeRoad(page);
}

/**
 * Complete the entire setup phase for a 2-player game
 * (Host: S+R, Guest: S+R, Guest: S+R, Host: S+R)
 */
export async function completeSetupPhase(
  hostPage: Page,
  guestPage: Page
) {
  // Round 1: Host, then Guest
  await completeSetupRound(hostPage);
  await completeSetupRound(guestPage);

  // Round 2: Guest, then Host (reverse order)
  await completeSetupRound(guestPage);
  await completeSetupRound(hostPage);

  // Wait for main game phase
  await expect(hostPage.locator("[data-cy='game-phase']")).toContainText(
    "PLAYING",
    { timeout: 10000 }
  );
}

// ============================================================================
// Test-Only Backend Endpoints (DEV_MODE only)
// ============================================================================

/**
 * Check if DEV_MODE test endpoints are available
 */
export async function isDevModeAvailable(request: APIRequestContext): Promise<boolean> {
  try {
    const response = await request.post(`${API_BASE}/test/grant-resources`, {
      data: { gameCode: "nonexistent-game", playerId: "test", resources: {} },
    });
    const status = response.status();
    const text = await response.text();
    
    // If we get 404 with "Test endpoints not available", DEV_MODE is disabled
    // If we get 404 with "Game not found", DEV_MODE is enabled but game doesn't exist
    // If we get 400+, DEV_MODE is enabled
    if (status === 404) {
      return text.trim() === "Game not found"; // DEV_MODE enabled, but game not found
    }
    return status !== 404; // Any other status means DEV_MODE is enabled
  } catch {
    return false;
  }
}

/**
 * Grant resources to a player (test endpoint)
 * Only available when backend is running with DEV_MODE=true
 */
export async function grantResources(
  request: APIRequestContext,
  gameCode: string,
  playerId: string,
  resources: {
    wood?: number;
    brick?: number;
    sheep?: number;
    wheat?: number;
    ore?: number;
  }
) {
  const response = await request.post(
    `${API_BASE}/test/grant-resources`,
    {
      data: { gameCode, playerId, resources },
    }
  );
  if (!response.ok()) {
    const errorText = await response.text();
    throw new Error(`Failed to grant resources (${response.status()}): ${errorText}`);
  }
  return response.json();
}

/**
 * Grant a development card to a player (test endpoint)
 * Only available when backend is running with DEV_MODE=true
 */
export async function grantDevCard(
  request: APIRequestContext,
  gameCode: string,
  playerId: string,
  cardType: "knight" | "road_building" | "year_of_plenty" | "monopoly" | "victory_point"
) {
  const response = await request.post(
    `${API_BASE}/test/grant-dev-card`,
    {
      data: { gameCode, playerId, cardType },
    }
  );
  if (!response.ok()) {
    const errorText = await response.text();
    throw new Error(`Failed to grant dev card (${response.status()}): ${errorText}`);
  }
  return response.json();
}

/**
 * Force the next dice roll to a specific value (test endpoint)
 * Only available when backend is running with DEV_MODE=true
 */
export async function forceDiceRoll(
  request: APIRequestContext,
  gameCode: string,
  diceValue: number
) {
  const response = await request.post(
    `${API_BASE}/test/force-dice-roll`,
    {
      data: { gameCode, diceValue },
    }
  );
  if (!response.ok()) {
    throw new Error(`Failed to force dice roll: ${await response.text()}`);
  }
  return response.json();
}

/**
 * Advance game to a specific phase (test endpoint)
 * Only available when backend is running with DEV_MODE=true
 */
export async function advanceToPhase(
  request: APIRequestContext,
  gameCode: string,
  phase: string
) {
  const response = await request.post(
    `${API_BASE}/test/set-game-state`,
    {
      data: { gameCode, phase },
    }
  );
  if (!response.ok()) {
    throw new Error(`Failed to advance to phase: ${await response.text()}`);
  }
  return response.json();
}

// ============================================================================
// Game Phase Helpers
// ============================================================================

/**
 * Wait for game to reach a specific phase
 */
export async function waitForGamePhase(page: Page, phase: string) {
  await expect(page.locator("[data-cy='game-phase']")).toContainText(phase, {
    timeout: 30000,
  });
}

/**
 * Roll dice (clicks roll dice button)
 */
export async function rollDice(page: Page) {
  await page.locator("[data-cy='roll-dice-btn']").click();
  // Wait for dice result to appear
  await expect(page.locator("[data-cy='dice-result']")).toBeVisible({
    timeout: 10000,
  });
  // Wait a bit more for game state to update to trade phase
  await page.waitForTimeout(1000);
}

/**
 * End turn (clicks end turn button)
 */
export async function endTurn(page: Page) {
  await page.locator("[data-cy='end-turn-btn']").click();
}

// ============================================================================
// Build Helpers
// ============================================================================

/**
 * Build a settlement (assumes player has resources and valid placement)
 */
export async function buildSettlement(page: Page, vertexSelector: string) {
  await page.locator("[data-cy='build-settlement-btn']").click();
  await page.locator(vertexSelector).click();
  await expect(page.locator(vertexSelector)).toHaveClass(/vertex--occupied/, {
    timeout: 10000,
  });
}

/**
 * Build a road (assumes player has resources and valid placement)
 */
export async function buildRoad(page: Page, edgeSelector: string) {
  await page.locator("[data-cy='build-road-btn']").click();
  await page.locator(edgeSelector).click();
  await expect(page.locator(edgeSelector)).toHaveClass(/edge--occupied/, {
    timeout: 10000,
  });
}

/**
 * Build a city (assumes player has resources and valid upgrade)
 */
export async function buildCity(page: Page, vertexSelector: string) {
  await page.locator("[data-cy='build-city-btn']").click();
  await page.locator(vertexSelector).click();
}

/**
 * Wait for specific resources to be updated in the UI after granting via API
 */
export async function waitForResourcesUpdated(
  page: Page,
  expectedResources: {
    ore?: number;
    wheat?: number;
    sheep?: number;
    wood?: number;
    brick?: number;
  }
) {
  // Wait for each expected resource to be at least the specified amount
  for (const [resource, minAmount] of Object.entries(expectedResources)) {
    if (minAmount === undefined) continue;
    
    await expect(async () => {
      const resourceElement = page.locator(`[data-cy='resource-${resource}'] .resource-count`);
      const currentText = await resourceElement.textContent();
      const currentAmount = parseInt(currentText || '0', 10);
      expect(currentAmount).toBeGreaterThanOrEqual(minAmount);
    }).toPass({
      timeout: 15000, // Extended timeout for WebSocket propagation
      intervals: [1000] // Check every second
    });
  }
}

/**
 * Grant resources and wait for them to be reflected in the UI
 */
export async function grantResourcesAndWait(
  request: APIRequestContext,
  page: Page,
  gameCode: string,
  playerId: string,
  resources: {
    wood?: number;
    brick?: number;
    sheep?: number;
    wheat?: number;
    ore?: number;
  }
) {
  // Grant the resources via API
  await grantResources(request, gameCode, playerId, resources);
  
  // Wait for the resources to be reflected in the UI
  await waitForResourcesUpdated(page, resources);
}

/**
 * Buy a development card (assumes player has resources)
 */
export async function buyDevelopmentCard(page: Page) {
  // Wait for button to be enabled (resources should be granted first)
  await expect(page.locator("[data-cy='buy-dev-card-btn']")).toBeEnabled({
    timeout: 10000
  });
  await page.locator("[data-cy='buy-dev-card-btn']").click();
  // Wait for resource count to update or card to appear
  await page.waitForTimeout(500);
}
