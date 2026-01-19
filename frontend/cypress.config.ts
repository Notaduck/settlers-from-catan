import { defineConfig } from "cypress";

export default defineConfig({
  reporter: "mochawesome",
  reporterOptions: {
    reportDir: "cypress/results",
    overwrite: false,
    html: false,
    json: true,
  },

  e2e: {
    baseUrl: "http://localhost:3000",
    supportFile: "cypress/support/e2e.ts",
    specPattern: "cypress/e2e/**/*.cy.{js,jsx,ts,tsx}",
    viewportWidth: 1280,
    viewportHeight: 800,
    video: false,
    screenshotOnRunFailure: true,
    defaultCommandTimeout: 10000,
    setupNodeEvents(on, config) {
      // implement node event listeners here
    },
  },
});

// import { defineConfig } from "cypress";

// export default defineConfig({
//   e2e: {
//     baseUrl: "http://localhost:3000",
//     supportFile: "cypress/support/e2e.ts",
//     specPattern: "cypress/e2e/**/*.cy.{js,jsx,ts,tsx}",
//     viewportWidth: 1280,
//     viewportHeight: 800,
//     video: true,
//     screenshotOnRunFailure: true,
//     defaultCommandTimeout: 10000,
//     setupNodeEvents(on, config) {
//       // implement node event listeners here
//     },
//   },
// });
