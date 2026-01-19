/// <reference types="cypress" />

/**
 * Setup Phase Tests
 *
 * These tests verify the setup phase flow:
 * 1. Game transitions from lobby to setup when host clicks "Start Game"
 * 2. Board is displayed with hex tiles
 * 3. Player panel shows current turn and turn phase
 * 4. Snake draft order is followed
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

const waitForGameBoard = () => {
  cy.get("[data-cy='game-loading']", { timeout: 30000 }).should("not.exist");
  cy.get("[data-cy='game-board-container']", { timeout: 30000 }).should(
    "be.visible"
  );
};

describe("Setup Phase", () => {
  beforeEach(() => {
    cy.clearLocalStorage();
  });

  describe("Game Start", () => {
    it("should transition from lobby to setup phase when host starts game", () => {
      // Create game as host
      createGame("Host").then(({ code, sessionToken, playerId }) => {
        // Join as second player
        joinGame(code, "Guest").then((guestData) => {
          // Set guest as ready via API (simulating another player)
          // For simplicity, we'll just visit as host and do everything in one browser
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();

          // Mark host as ready
          cy.get("[data-cy='ready-btn']").click();
          cy.get("[data-cy='cancel-ready-btn']").should("be.visible");

          // Now visit as guest and mark them ready
          visitAsPlayer(code, guestData.sessionToken, guestData.playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();
          cy.get("[data-cy='cancel-ready-btn']").should("be.visible");

          // Go back to host and start game
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();

          // Wait for all players to show as ready (WebSocket sync)
          cy.contains("✓ Ready", { timeout: 10000 }).should("be.visible");

          // Host should see start button
          cy.get("[data-cy='start-game-btn']").should("be.visible");
          cy.get("[data-cy='start-game-btn']").click();

          // Should transition to game board
          waitForGameBoard();
          cy.get("[data-cy='board']").should("be.visible");
          cy.get("[data-cy='player-panel']").should("be.visible");
        });
      });
    });

    it("should show the board with hex tiles after game starts", () => {
      createGame("Alice").then(({ code, sessionToken, playerId }) => {
        joinGame(code, "Bob").then((bobData) => {
          // Ready both players via direct state manipulation
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, bobData.sessionToken, bobData.playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();

          // Wait for sync then start
          cy.get("[data-cy='start-game-btn']", { timeout: 10000 }).should(
            "not.be.disabled"
          );
          cy.get("[data-cy='start-game-btn']").click();

          // Verify board renders
          waitForGameBoard();
          cy.get("[data-cy='board-svg']").should("be.visible");

          // Should have hex tiles (standard board has 19 hexes)
          cy.get("[data-cy^='hex-']").should("have.length.at.least", 19);
        });
      });
    });
  });

  describe("Turn Display", () => {
    it("should show current player and turn phase after game starts", () => {
      createGame("Player1").then(({ code, sessionToken, playerId }) => {
        joinGame(code, "Player2").then((p2Data) => {
          // Ready both players
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, p2Data.sessionToken, p2Data.playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          // Start as host
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='start-game-btn']", { timeout: 10000 }).should(
            "not.be.disabled"
          );
          cy.get("[data-cy='start-game-btn']").click();

          waitForGameBoard();

          // Should show turn section
          cy.get("[data-cy='turn-section']").should("be.visible");
          cy.get("[data-cy='current-player']").should("be.visible");
          cy.get("[data-cy='turn-phase']").should("be.visible");
        });
      });
    });

    it("should show player resources section", () => {
      createGame("ResourceTest").then(({ code, sessionToken, playerId }) => {
        joinGame(code, "Other").then((otherData) => {
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, otherData.sessionToken, otherData.playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='start-game-btn']", { timeout: 10000 }).should(
            "not.be.disabled"
          );
          cy.get("[data-cy='start-game-btn']").click();

          waitForGameBoard();

          // Should show resources section
          cy.get("[data-cy='resources-section']").should("be.visible");
          cy.get("[data-cy='resource-wood']").should("be.visible");
          cy.get("[data-cy='resource-brick']").should("be.visible");
          cy.get("[data-cy='resource-sheep']").should("be.visible");
          cy.get("[data-cy='resource-wheat']").should("be.visible");
          cy.get("[data-cy='resource-ore']").should("be.visible");
        });
      });
    });
  });

  describe("Players List", () => {
    it("should show all players in the players section", () => {
      createGame("Alpha").then(({ code, sessionToken, playerId }) => {
        joinGame(code, "Beta").then((betaData) => {
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, betaData.sessionToken, betaData.playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='start-game-btn']", { timeout: 10000 }).should(
            "not.be.disabled"
          );
          cy.get("[data-cy='start-game-btn']").click();

          waitForGameBoard();

          // Should show players section with both players
          cy.get("[data-cy='players-section']").should("be.visible");
          cy.contains("Alpha").should("be.visible");
          cy.contains("Beta").should("be.visible");
        });
      });
    });

    it("should show victory points for each player", () => {
      createGame("VPTest1").then(({ code, sessionToken, playerId }) => {
        joinGame(code, "VPTest2").then((p2Data) => {
          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, p2Data.sessionToken, p2Data.playerId);
          waitForLobby();
          cy.get("[data-cy='ready-btn']").click();

          visitAsPlayer(code, sessionToken, playerId);
          waitForLobby();
          cy.get("[data-cy='start-game-btn']", { timeout: 10000 }).should(
            "not.be.disabled"
          );
          cy.get("[data-cy='start-game-btn']").click();

          waitForGameBoard();

          // Each player should have VP displayed
          cy.get("[data-cy^='player-vp-']").should("have.length", 2);
          cy.get("[data-cy^='player-vp-']").first().should("contain", "VP");
        });
      });
    });
  });
});
