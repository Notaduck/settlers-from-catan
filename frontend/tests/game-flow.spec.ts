import {
  test,
  expect,
  type Page,
  type APIRequestContext,
} from "@playwright/test";

const API_BASE = "http://localhost:8080";

interface GameSession {
  code: string;
  sessionToken: string;
  playerId: string;
}

async function createGame(
  request: APIRequestContext,
  playerName: string
): Promise<GameSession> {
  const response = await request.post(`${API_BASE}/api/games`, {
    data: { playerName },
  });
  return response.json();
}

async function joinGame(
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

async function visitAsPlayer(page: Page, session: GameSession) {
  // First navigate to the page
  await page.goto("/");

  // Then set sessionStorage and reload to pick up the state
  await page.evaluate((s) => {
    sessionStorage.setItem("sessionToken", s.sessionToken);
    sessionStorage.setItem("gameCode", s.code);
    sessionStorage.setItem("playerId", s.playerId);
    localStorage.removeItem("sessionToken");
    localStorage.removeItem("gameCode");
    localStorage.removeItem("playerId");
  }, session);

  // Reload to connect with the session
  await page.reload();
}

async function waitForLobby(page: Page) {
  await expect(page.locator("[data-cy='game-loading']")).not.toBeVisible({
    timeout: 30000,
  });
  await expect(page.locator("[data-cy='game-waiting']")).toBeVisible({
    timeout: 30000,
  });
}

async function waitForGameBoard(page: Page) {
  await expect(page.locator("[data-cy='game-loading']")).not.toBeVisible({
    timeout: 30000,
  });
  await expect(page.locator("[data-cy='game-board-container']")).toBeVisible({
    timeout: 30000,
  });
}

test.describe("Lobby", () => {
  test("should create a new game and show lobby", async ({ page, request }) => {
    const session = await createGame(request, "Player1");
    await visitAsPlayer(page, session);
    await waitForLobby(page);

    await expect(page.getByText("Player1")).toBeVisible();
    // Use data-cy selector to be more specific
    await expect(page.locator("[data-cy='game-code']")).toBeVisible();
  });

  test("should join an existing game", async ({ page, request }) => {
    const host = await createGame(request, "Host");
    const guest = await joinGame(request, host.code, "Guest");

    await visitAsPlayer(page, guest);
    await waitForLobby(page);

    await expect(page.getByText("Host")).toBeVisible();
    await expect(page.getByText("Guest")).toBeVisible();
  });
});

test.describe("Ready System", () => {
  test("should allow players to toggle ready state", async ({
    page,
    request,
  }) => {
    const session = await createGame(request, "TestPlayer");
    await visitAsPlayer(page, session);
    await waitForLobby(page);

    await page.locator("[data-cy='ready-btn']").click();

    // Wait for button to change after WebSocket round-trip
    await expect(page.locator("[data-cy='cancel-ready-btn']")).toBeVisible({
      timeout: 10000,
    });
    await expect(page.getByText("✓ Ready")).toBeVisible();
  });
});

test.describe("Game Start", () => {
  test("should start game when both players ready", async ({
    page,
    context,
    request,
  }) => {
    const host = await createGame(request, "Host");
    const guest = await joinGame(request, host.code, "Guest");

    const hostPage = page;
    const guestPage = await context.newPage();

    await visitAsPlayer(hostPage, host);
    await waitForLobby(hostPage);

    await visitAsPlayer(guestPage, guest);
    await waitForLobby(guestPage);

    // Both ready up - wait for state updates after each click
    await guestPage.locator("[data-cy='ready-btn']").click();
    await expect(
      guestPage.locator("[data-cy='cancel-ready-btn']")
    ).toBeVisible({ timeout: 10000 });

    await hostPage.locator("[data-cy='ready-btn']").click();
    await expect(hostPage.locator("[data-cy='cancel-ready-btn']")).toBeVisible({
      timeout: 10000,
    });

    // Wait for start button to become enabled
    await expect(
      hostPage.locator("[data-cy='start-game-btn']")
    ).toBeEnabled({ timeout: 10000 });

    // Host starts game
    await hostPage.locator("[data-cy='start-game-btn']").click();

    // Both see game board
    await waitForGameBoard(hostPage);
    await waitForGameBoard(guestPage);

    await expect(hostPage.locator("[data-cy='board']")).toBeVisible();

    await guestPage.close();
  });
});
