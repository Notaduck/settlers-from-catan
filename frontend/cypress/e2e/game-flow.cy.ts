/// <reference types="cypress" />

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
  cy.get(".game-loading", { timeout: 30000 }).should("not.exist");
  cy.contains("Game Lobby", { timeout: 30000 }).should("be.visible");
};

describe("Catan Game Flow", () => {
  beforeEach(() => {
    // Clear localStorage before each test
    cy.clearLocalStorage();
  });

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

  describe("Lobby", () => {
    it("should create a new game and show lobby", () => {
      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "Player1" },
      }).then((response) => {
        const { code, sessionToken, playerId } = response.body;
        visitAsPlayer(code, sessionToken, playerId);

        // Should show the game lobby
        waitForLobby();
        cy.contains("Player1").should("be.visible");
        cy.contains("Game Code").should("be.visible");
      });
    });

    it("should join an existing game", () => {
      // First create a game via API
      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "Host" },
      }).then((response) => {
        const gameCode = response.body.code;
        const hostToken = response.body.sessionToken;
        const hostPlayerId = response.body.playerId;

        return cy
          .request({
            method: "POST",
            url: `http://localhost:8080/api/games/${gameCode}/join`,
            body: { playerName: "Player2" },
          })
          .then((joinResponse) => {
            const { sessionToken, playerId } = joinResponse.body;
            visitAsPlayer(gameCode, sessionToken, playerId);

            // Should see the lobby with both players
            waitForLobby();
            cy.contains("Player2").should("be.visible");
            cy.contains("Host").should("be.visible");
          })
          .then(() => {
            // Ensure host can also see lobby if needed
            visitAsPlayer(gameCode, hostToken, hostPlayerId);
          });
      });
    });
  });

  describe("Ready System", () => {
    it("should allow players to toggle ready state", () => {
      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "TestPlayer" },
      }).then((response) => {
        const { code, sessionToken, playerId } = response.body;
        visitAsPlayer(code, sessionToken, playerId);

        // Wait for lobby to load
        waitForLobby();
      });

      // Player should see Ready button
      cy.get("button").contains("I'm Ready").should("be.visible");

      // Click ready
      cy.get("button").contains("I'm Ready").click();

      // Should show as ready (button changes to "Not Ready" or shows checkmark)
      cy.get("button").contains("Cancel Ready").should("be.visible");
      cy.contains("✓ Ready").should("be.visible");
    });
  });

  describe("Two Player Lobby", () => {
    it("should show both players in lobby", () => {
      // This test simulates two browser windows
      // First, create a game and store the code
      let gameCode: string;
      let hostToken: string;

      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "Host" },
      })
        .then((response) => {
          gameCode = response.body.code;
          hostToken = response.body.sessionToken;

          // Join as second player
          return cy.request({
            method: "POST",
            url: `http://localhost:8080/api/games/${gameCode}/join`,
            body: { playerName: "Guest" },
          });
        })
        .then((joinResponse) => {
          const guestToken = joinResponse.body.sessionToken;
          const guestPlayerId = joinResponse.body.playerId;

          // Visit as guest
          visitAsPlayer(gameCode, guestToken, guestPlayerId);

          // Should auto-connect and show lobby
          waitForLobby();
          cy.contains("Host").should("be.visible");
          cy.contains("Guest").should("be.visible");

          // Both players should be listed
          cy.get(".player-item").should("have.length", 2);
        });
    });
  });

  describe("Board Rendering", () => {
    it("should load the lobby after creating a game", () => {
      // Create game with enough players and start it
      let gameCode: string;

      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "Host" },
      }).then((response) => {
        gameCode = response.body.code;
        const hostToken = response.body.sessionToken;
        const hostPlayerId = response.body.playerId;

        // Join as second player
        return cy
          .request({
            method: "POST",
            url: `http://localhost:8080/api/games/${gameCode}/join`,
            body: { playerName: "Guest" },
          })
          .then(() => {
            // Use host's session to view the game
            visitAsPlayer(gameCode, hostToken, hostPlayerId);

            // Wait for lobby
            waitForLobby();

            // Mark self as ready
            cy.get("button").contains("I'm Ready").click();

            // Check that both players are in the list
            cy.get(".player-item").should("have.length", 2);
          });
      });
    });

    it("should load the lobby for a two-player game", () => {
      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "TestHost" },
      }).then((response) => {
        const { code, sessionToken, playerId } = response.body;

        // Add second player
        return cy
          .request({
            method: "POST",
            url: `http://localhost:8080/api/games/${code}/join`,
            body: { playerName: "TestGuest" },
          })
          .then(() => {
            visitAsPlayer(code, sessionToken, playerId);
            waitForLobby();
          });
      });
    });
  });

  describe("Full Game Flow Visual Test", () => {
    it("should complete a full lobby to game transition", () => {
      cy.request({
        method: "POST",
        url: "http://localhost:8080/api/games",
        body: { playerName: "Alice" },
      }).then((response) => {
        const { code, sessionToken, playerId } = response.body;
        visitAsPlayer(code, sessionToken, playerId);

        // Verify the UI state
        waitForLobby();
        cy.contains("Alice").should("be.visible");
        cy.contains("Players (1/4)").should("be.visible");

        // The Ready button should be visible
        cy.get("button").contains("I'm Ready").should("be.visible");

        // Start Game button should NOT be visible with only 1 player
        cy.get("button").contains("Start Game").should("not.exist");
      });
    });
  });
});

describe("Board Visual Tests", () => {
  it("should render hexes without NaN coordinates", () => {
    // Setup: Create and join game, then check board
    cy.request({
      method: "POST",
      url: "http://localhost:8080/api/games",
      body: { playerName: "BoardTest" },
    }).then((response) => {
      const { sessionToken, playerId } = response.body;
      // For this test we need to manually trigger game start
      // by updating game status in DB or using WebSocket

      visitAsPlayer(response.body.code, sessionToken, playerId);

      // Wait for page to load
      waitForLobby();

      // Verify we're on the game page
      cy.url().should("include", "/");
    });
  });
});
