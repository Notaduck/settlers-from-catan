/// <reference types="cypress" />

// Custom commands for Catan game testing

declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Create a new game with the given player name
       */
      createGame(playerName: string): Chainable<{
        gameCode: string;
        sessionToken: string;
        playerId: string;
      }>;

      /**
       * Join an existing game with the given code and player name
       */
      joinGame(
        gameCode: string,
        playerName: string
      ): Chainable<{ sessionToken: string; playerId: string }>;

      /**
       * Wait for WebSocket connection to be established
       */
      waitForConnection(): Chainable<void>;

      /**
       * Set player as ready in the lobby
       */
      setReady(): Chainable<void>;

      /**
       * Start the game (host only)
       */
      startGame(): Chainable<void>;
    }
  }
}

// Create a new game
Cypress.Commands.add("createGame", (playerName: string) => {
  return cy
    .request({
      method: "POST",
      url: "http://localhost:8080/api/games",
      body: { playerName },
    })
    .then((response) => {
      expect(response.status).to.eq(200);
      const { code, sessionToken, playerId } = response.body;

      // Store in localStorage for the app to use
      cy.window().then((win) => {
        win.localStorage.setItem("sessionToken", sessionToken);
        win.localStorage.setItem("gameCode", code);
        win.localStorage.setItem("playerId", playerId);
      });

      return { gameCode: code, sessionToken, playerId };
    });
});

// Join an existing game
Cypress.Commands.add("joinGame", (gameCode: string, playerName: string) => {
  return cy
    .request({
      method: "POST",
      url: `http://localhost:8080/api/games/${gameCode}/join`,
      body: { playerName },
    })
    .then((response) => {
      expect(response.status).to.eq(200);
      const { sessionToken, playerId } = response.body;

      // Store in localStorage
      cy.window().then((win) => {
        win.localStorage.setItem("sessionToken", sessionToken);
        win.localStorage.setItem("gameCode", gameCode);
        win.localStorage.setItem("playerId", playerId);
      });

      return { sessionToken, playerId };
    });
});

// Wait for WebSocket connection
Cypress.Commands.add("waitForConnection", () => {
  cy.get(".game-loading", { timeout: 5000 }).should("not.exist");
});

// Set player as ready
Cypress.Commands.add("setReady", () => {
  cy.get("button").contains("I'm Ready").click();
});

// Start the game
Cypress.Commands.add("startGame", () => {
  cy.get("button").contains("Start Game").click();
});

export {};
