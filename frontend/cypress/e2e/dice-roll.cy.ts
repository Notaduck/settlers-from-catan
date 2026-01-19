/// <reference types="cypress" />

/**
 * Dice Roll and Resource Distribution E2E Tests
 * Tests the Phase 2 functionality: dice rolling and resource distribution
 *
 * Note: These are simplified lobby tests until we have full game start flow.
 * The backend dice logic is thoroughly tested in dice_test.go
 */

const API_BASE = "http://localhost:8080";

interface CreateGameResponse {
  code: string;
  sessionToken: string;
  playerId: string;
}

interface JoinGameResponse {
  sessionToken: string;
  playerId: string;
}

const createGame = (
  playerName: string
): Cypress.Chainable<CreateGameResponse> => {
  return cy
    .request({
      method: "POST",
      url: `${API_BASE}/api/games`,
      body: { playerName },
    })
    .then((response) => response.body as CreateGameResponse);
};

const joinGame = (
  gameCode: string,
  playerName: string
): Cypress.Chainable<JoinGameResponse> => {
  return cy
    .request({
      method: "POST",
      url: `${API_BASE}/api/games/${gameCode}/join`,
      body: { playerName },
    })
    .then((response) => response.body as JoinGameResponse);
};

const visitAsPlayer = (
  gameCode: string,
  sessionToken: string,
  playerId: string
) => {
  cy.visit("/", {
    onBeforeLoad(win) {
      win.localStorage.setItem("sessionToken", sessionToken);
      win.localStorage.setItem("gameCode", gameCode);
      win.localStorage.setItem("playerId", playerId);
    },
  });
};

const waitForLobby = () => {
  cy.get("[data-cy='game-loading']", { timeout: 30000 }).should("not.exist");
  cy.get("[data-cy='game-waiting']", { timeout: 30000 }).should("be.visible");
};

describe("Dice Roll - Lobby Prerequisites", () => {
  beforeEach(() => {
    cy.clearLocalStorage();
  });

  it("should create a game and show the lobby", () => {
    createGame("DiceTestHost").then(({ code, sessionToken, playerId }) => {
      visitAsPlayer(code, sessionToken, playerId);
      waitForLobby();
      cy.contains("DiceTestHost").should("be.visible");
    });
  });

  it("should show two players in the lobby", () => {
    createGame("Host").then(({ code, sessionToken, playerId }) => {
      joinGame(code, "Guest").then(() => {
        visitAsPlayer(code, sessionToken, playerId);
        waitForLobby();
        cy.contains("Host").should("be.visible");
        cy.contains("Guest").should("be.visible");
        cy.contains("Players (2/4)").should("be.visible");
      });
    });
  });

  it("should allow host to ready up", () => {
    createGame("ReadyHost").then(({ code, sessionToken, playerId }) => {
      visitAsPlayer(code, sessionToken, playerId);
      waitForLobby();

      // Click ready button
      cy.get("[data-cy='ready-btn']").click();

      // Should show cancel ready button now
      cy.get("[data-cy='cancel-ready-btn']").should("be.visible");

      // Should show ready status
      cy.contains("✓ Ready").should("be.visible");
    });
  });
});
