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
  await expect(
    guestPage.locator("[data-cy='cancel-ready-btn']")
  ).toBeVisible({ timeout: 10000 });

  await hostPage.locator("[data-cy='ready-btn']").click();
  await expect(hostPage.locator("[data-cy='cancel-ready-btn']")).toBeVisible({
    timeout: 10000,
  });

  await expect(
    hostPage.locator("[data-cy='start-game-btn']")
  ).toBeEnabled({ timeout: 10000 });
  await hostPage.locator("[data-cy='start-game-btn']").click();

  await waitForGameBoard(hostPage);

  return { hostPage, guestPage };
}

async function placeSettlement(page: Page) {
  const placementMode = page.locator("[data-cy='placement-mode']");
  await expect(placementMode).toContainText("Place Settlement", {
    timeout: 30000,
  });
  const validVertex = page.locator("[data-cy^='vertex-'].vertex--valid").first();
  await expect(validVertex).toBeVisible({ timeout: 30000 });
  await validVertex.click();
}

async function placeRoad(page: Page) {
  const placementMode = page.locator("[data-cy='placement-mode']");
  await expect(placementMode).toContainText("Place Road", { timeout: 30000 });
  const validEdge = page.locator("[data-cy^='edge-'].edge--valid").first();
  await expect(validEdge).toBeVisible({ timeout: 30000 });
  await validEdge.click();
}

test.describe("Interactive Board", () => {
  test("vertices render on board", async ({ page, context, request }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    const vertices = hostPage.locator("[data-cy^='vertex-']");
    await expect(vertices.first()).toBeVisible({ timeout: 30000 });

    await guestPage.close();
  });

  test("edges render on board", async ({ page, context, request }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    const edges = hostPage.locator("[data-cy^='edge-']");
    await expect(edges.first()).toBeVisible({ timeout: 30000 });

    await guestPage.close();
  });

  test("shows placement mode for setup turn", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    const placementMode = hostPage.locator("[data-cy='placement-mode']");
    await expect(placementMode).toBeVisible({ timeout: 30000 });
    await expect(placementMode).toContainText("Place Settlement");

    await guestPage.close();
  });

  test("clicking vertex during setup places settlement", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    await placeSettlement(hostPage);
    await expect(hostPage.locator(".vertex--occupied")).toHaveCount(1, {
      timeout: 30000,
    });
    await expect(hostPage.locator("[data-cy='placement-mode']")).toContainText(
      "Place Road",
      { timeout: 30000 }
    );

    await guestPage.close();
  });

  test("clicking edge during setup places road", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    await placeSettlement(hostPage);
    await placeRoad(hostPage);

    await expect(hostPage.locator(".edge--occupied")).toHaveCount(1, {
      timeout: 30000,
    });

    await guestPage.close();
  });

  test("invalid vertices are not clickable during setup", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    await placeSettlement(hostPage);
    await placeRoad(hostPage);

    await expect(guestPage.locator("[data-cy='placement-mode']")).toContainText(
      "Place Settlement",
      { timeout: 30000 }
    );

    const invalidVertices = guestPage.locator(
      "[data-cy^='vertex-'].vertex--empty:not(.vertex--valid)"
    );
    const invalidCount = await invalidVertices.count();
    expect(invalidCount).toBeGreaterThan(0);

    const occupiedBefore = await guestPage.locator(".vertex--occupied").count();
    await invalidVertices.first().click();
    await expect(guestPage.locator(".vertex--occupied")).toHaveCount(
      occupiedBefore,
      { timeout: 10000 }
    );

    await guestPage.close();
  });

  test("highlights valid placements during setup", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage } = await startTwoPlayerGame(
      page,
      context,
      request
    );

    const validVertices = hostPage.locator(
      "[data-cy^='vertex-'].vertex--valid"
    );
    const validCount = await validVertices.count();
    expect(validCount).toBeGreaterThan(0);

    await guestPage.close();
  });
});
