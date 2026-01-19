import {
  test,
  expect,
  type Page,
  type BrowserContext,
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
  await page.goto("/");
  await page.evaluate((s) => {
    localStorage.setItem("sessionToken", s.sessionToken);
    localStorage.setItem("gameCode", s.code);
    localStorage.setItem("playerId", s.playerId);
  }, session);
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

async function startTwoPlayerGame(
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

  await guestPage.locator("[data-cy='ready-btn']").click();
  await expect(guestPage.locator("[data-cy='cancel-ready-btn']")).toBeVisible({
    timeout: 10000,
  });

  await hostPage.locator("[data-cy='ready-btn']").click();
  await expect(hostPage.locator("[data-cy='cancel-ready-btn']")).toBeVisible({
    timeout: 10000,
  });

  await expect(hostPage.locator("[data-cy='start-game-btn']")).toBeEnabled({
    timeout: 10000,
  });
  await hostPage.locator("[data-cy='start-game-btn']").click();

  await waitForGameBoard(hostPage);

  return { hostPage, guestPage };
}

test.describe("Setup Phase UI", () => {
  test("shows setup phase banner and turn indicator", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    const banner = hostPage.locator("[data-cy='setup-phase-banner']");
    await expect(banner).toBeVisible({ timeout: 30000 });
    await expect(banner).toContainText("Setup Phase - Round 1");

    const turnIndicator = hostPage.locator(
      "[data-cy='setup-turn-indicator']"
    );
    await expect(turnIndicator).toBeVisible({ timeout: 30000 });
    await expect(turnIndicator).toContainText("Host");

    await guestPage.close();
  });
});
