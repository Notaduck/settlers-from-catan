import {
  test,
  expect,
  type Page,
  type BrowserContext,
  type APIRequestContext,
} from "@playwright/test";
import {
  createGame,
  joinGame,
  visitAsPlayer,
  waitForLobby,
  setPlayerReady,
  startGame,
  waitForGamePhase,
  grantResources,
  forceDiceRoll,
  placeSettlement,
  placeRoad,
} from "./helpers";

const RESOURCE_TYPES = ["wood", "brick", "sheep", "wheat", "ore"] as const;

type ResourceType = (typeof RESOURCE_TYPES)[number];

async function startRobberGame(
  page: Page,
  context: BrowserContext,
  request: APIRequestContext
) {
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

  await placeSettlement(hostPage);
  await placeRoad(hostPage);

  const guestSettlement1 = await placeSettlement(guestPage);
  await placeRoad(guestPage);

  const guestSettlement2 = await placeSettlement(guestPage);
  await placeRoad(guestPage);

  await placeSettlement(hostPage);
  await placeRoad(hostPage);

  await waitForGamePhase(hostPage, "PLAYING");

  return {
    hostPage,
    guestPage,
    hostSession: host,
    guestSession: guest,
    guestSettlements: [guestSettlement1, guestSettlement2].filter(Boolean),
  };
}

async function getDiscardCount(page: Page): Promise<number> {
  const heading = await page
    .locator("[data-cy='discard-modal'] h2")
    .textContent();
  const match = heading?.match(/Discard\s+(\d+)\s+Cards/);
  if (!match) {
    throw new Error("Unable to parse discard count");
  }
  return Number(match[1]);
}

async function selectDiscardResources(
  page: Page,
  count: number,
  resource: ResourceType = "wood"
) {
  const addButton = page
    .locator(`[data-cy='discard-card-${resource}'] button`)
    .nth(1);
  for (let i = 0; i < count; i += 1) {
    await addButton.click();
  }
}

async function getResourceTotal(page: Page): Promise<number> {
  let total = 0;
  for (const resource of RESOURCE_TYPES) {
    const countText = await page
      .locator(`[data-cy='resource-${resource}'] .resource-count`)
      .textContent();
    total += Number(countText?.trim() ?? 0);
  }
  return total;
}

async function getRobberHexWithGuestAdjacency(
  page: Page,
  guestVertexDataCys: string[]
): Promise<string> {
  const dataCy = await page.evaluate((vertexDataCys) => {
    const parseTranslate = (el: Element | null) => {
      if (!el) return null;
      const transform = el.getAttribute("transform") || "";
      const match = /translate\(([-\d.]+),\s*([-\d.]+)\)/.exec(transform);
      if (!match) return null;
      return { x: Number(match[1]), y: Number(match[2]) };
    };

    const vertices = vertexDataCys
      .map((cy) => parseTranslate(document.querySelector(`[data-cy='${cy}']`)))
      .filter((pos): pos is { x: number; y: number } => Boolean(pos));
    if (vertices.length === 0) return null;

    const hexes = Array.from(
      document.querySelectorAll("[data-cy^='robber-hex-']")
    );
    if (hexes.length === 0) return null;

    const samplePolygon = hexes[0]?.querySelector("polygon");
    const points = samplePolygon?.getAttribute("points") || "";
    const radius = points
      .split(" ")
      .map((pair) => pair.split(",").map(Number))
      .filter((pair) => pair.length === 2 && !pair.some(Number.isNaN))
      .reduce((max, [x, y]) => Math.max(max, Math.hypot(x, y)), 0);

    const threshold = radius * 1.1;

    for (const hex of hexes) {
      const pos = parseTranslate(hex);
      if (!pos) continue;
      const hasGuestAdjacency = vertices.some((v) => {
        const dx = v.x - pos.x;
        const dy = v.y - pos.y;
        return Math.hypot(dx, dy) <= threshold;
      });
      if (hasGuestAdjacency) {
        return hex.getAttribute("data-cy");
      }
    }

    return null;
  }, guestVertexDataCys);

  if (!dataCy) {
    throw new Error("Unable to find robber hex adjacent to guest settlement");
  }

  return dataCy;
}

async function getRobberHexWithoutGuestAdjacency(
  page: Page,
  guestVertexDataCys: string[]
): Promise<string> {
  const dataCy = await page.evaluate((vertexDataCys) => {
    const parseTranslate = (el: Element | null) => {
      if (!el) return null;
      const transform = el.getAttribute("transform") || "";
      const match = /translate\(([-\d.]+),\s*([-\d.]+)\)/.exec(transform);
      if (!match) return null;
      return { x: Number(match[1]), y: Number(match[2]) };
    };

    const vertices = vertexDataCys
      .map((cy) => parseTranslate(document.querySelector(`[data-cy='${cy}']`)))
      .filter((pos): pos is { x: number; y: number } => Boolean(pos));
    if (vertices.length === 0) return null;

    const hexes = Array.from(
      document.querySelectorAll("[data-cy^='robber-hex-']")
    );
    if (hexes.length === 0) return null;

    const samplePolygon = hexes[0]?.querySelector("polygon");
    const points = samplePolygon?.getAttribute("points") || "";
    const radius = points
      .split(" ")
      .map((pair) => pair.split(",").map(Number))
      .filter((pair) => pair.length === 2 && !pair.some(Number.isNaN))
      .reduce((max, [x, y]) => Math.max(max, Math.hypot(x, y)), 0);

    const threshold = radius * 1.1;

    for (const hex of hexes) {
      const pos = parseTranslate(hex);
      if (!pos) continue;
      const hasGuestAdjacency = vertices.some((v) => {
        const dx = v.x - pos.x;
        const dy = v.y - pos.y;
        return Math.hypot(dx, dy) <= threshold;
      });
      if (!hasGuestAdjacency) {
        return hex.getAttribute("data-cy");
      }
    }

    return null;
  }, guestVertexDataCys);

  if (!dataCy) {
    throw new Error("Unable to find robber hex without guest adjacency");
  }

  return dataCy;
}

test.describe("Robber Flow", () => {
  test("Rolling 7 shows discard modal for players with >7 cards", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, guestPage, hostSession } = await startRobberGame(
      page,
      context,
      request
    );

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeVisible({
      timeout: 10000,
    });
    await expect(guestPage.locator("[data-cy='discard-modal']")).toBeHidden();
  });

  test("Discard modal enforces correct card count", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, hostSession } = await startRobberGame(
      page,
      context,
      request
    );

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    const discardModal = hostPage.locator("[data-cy='discard-modal']");
    await expect(discardModal).toBeVisible({ timeout: 10000 });

    const submitButton = hostPage.locator("[data-cy='discard-submit']");
    await expect(submitButton).toBeDisabled();

    const requiredCount = await getDiscardCount(hostPage);
    if (requiredCount > 1) {
      await selectDiscardResources(hostPage, requiredCount - 1);
      await expect(submitButton).toBeDisabled();
    }

    await selectDiscardResources(hostPage, 1);
    await expect(submitButton).toBeEnabled();

    await submitButton.click();
    await expect(discardModal).toBeHidden({ timeout: 10000 });
  });

  test("After discard, robber move UI appears", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, hostSession } = await startRobberGame(
      page,
      context,
      request
    );

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeVisible({
      timeout: 10000,
    });
    const requiredCount = await getDiscardCount(hostPage);
    await selectDiscardResources(hostPage, requiredCount);
    await hostPage.locator("[data-cy='discard-submit']").click();
    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeHidden({
      timeout: 10000,
    });

    await expect(hostPage.locator("[data-cy^='robber-hex-']").first()).toBeVisible({
      timeout: 10000,
    });
  });

  test("Clicking hex moves robber", async ({ page, context, request }) => {
    const { hostPage, hostSession, guestSettlements } = await startRobberGame(
      page,
      context,
      request
    );

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeVisible({
      timeout: 10000,
    });
    const requiredCount = await getDiscardCount(hostPage);
    await selectDiscardResources(hostPage, requiredCount);
    await hostPage.locator("[data-cy='discard-submit']").click();
    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeHidden({
      timeout: 10000,
    });

    await expect(hostPage.locator("[data-cy^='robber-hex-']").first()).toBeVisible({
      timeout: 10000,
    });

    const robberHexDataCy = await getRobberHexWithGuestAdjacency(
      hostPage,
      guestSettlements
    );
    await hostPage.locator(`[data-cy='${robberHexDataCy}']`).click();

    await expect(hostPage.locator("[data-cy^='robber-hex-']")).toHaveCount(0);
  });

  test("Steal UI shows adjacent players", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, hostSession, guestSession, guestSettlements } =
      await startRobberGame(page, context, request);

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeVisible({
      timeout: 10000,
    });
    const requiredCount = await getDiscardCount(hostPage);
    await selectDiscardResources(hostPage, requiredCount);
    await hostPage.locator("[data-cy='discard-submit']").click();
    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeHidden({
      timeout: 10000,
    });

    await expect(hostPage.locator("[data-cy^='robber-hex-']").first()).toBeVisible({
      timeout: 10000,
    });

    const robberHexDataCy = await getRobberHexWithGuestAdjacency(
      hostPage,
      guestSettlements
    );
    await hostPage.locator(`[data-cy='${robberHexDataCy}']`).click();

    await expect(hostPage.locator("[data-cy='steal-modal']")).toBeVisible({
      timeout: 10000,
    });
    await expect(
      hostPage.locator(`[data-cy='steal-player-${guestSession.playerId}']`)
    ).toBeVisible({ timeout: 10000 });
  });

  test("Stealing transfers a resource", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, hostSession, guestSession, guestSettlements } =
      await startRobberGame(page, context, request);

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });
    await grantResources(request, hostSession.code, guestSession.playerId, {
      brick: 2,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeVisible({
      timeout: 10000,
    });
    const requiredCount = await getDiscardCount(hostPage);
    await selectDiscardResources(hostPage, requiredCount);
    await hostPage.locator("[data-cy='discard-submit']").click();
    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeHidden({
      timeout: 10000,
    });

    await expect(hostPage.locator("[data-cy^='robber-hex-']").first()).toBeVisible({
      timeout: 10000,
    });

    const robberHexDataCy = await getRobberHexWithGuestAdjacency(
      hostPage,
      guestSettlements
    );
    await hostPage.locator(`[data-cy='${robberHexDataCy}']`).click();

    await expect(hostPage.locator("[data-cy='steal-modal']")).toBeVisible({
      timeout: 10000,
    });

    const beforeStealTotal = await getResourceTotal(hostPage);

    await hostPage
      .locator(`[data-cy='steal-player-${guestSession.playerId}']`)
      .click();

    await expect(hostPage.locator("[data-cy='steal-modal']")).toBeHidden({
      timeout: 10000,
    });

    await expect.poll(async () => getResourceTotal(hostPage)).toBe(
      beforeStealTotal + 1
    );
  });

  test("No steal phase when no adjacent players", async ({
    page,
    context,
    request,
  }) => {
    const { hostPage, hostSession, guestSettlements } = await startRobberGame(
      page,
      context,
      request
    );

    await grantResources(request, hostSession.code, hostSession.playerId, {
      wood: 10,
    });

    await forceDiceRoll(request, hostSession.code, 7);

    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeVisible({
      timeout: 10000,
    });
    const requiredCount = await getDiscardCount(hostPage);
    await selectDiscardResources(hostPage, requiredCount);
    await hostPage.locator("[data-cy='discard-submit']").click();
    await expect(hostPage.locator("[data-cy='discard-modal']")).toBeHidden({
      timeout: 10000,
    });

    await expect(hostPage.locator("[data-cy^='robber-hex-']").first()).toBeVisible({
      timeout: 10000,
    });

    const robberHexDataCy = await getRobberHexWithoutGuestAdjacency(
      hostPage,
      guestSettlements
    );
    await hostPage.locator(`[data-cy='${robberHexDataCy}']`).click();

    await expect(hostPage.locator("[data-cy='steal-modal']")).toBeHidden({
      timeout: 10000,
    });
  });
});
