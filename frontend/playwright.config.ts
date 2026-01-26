import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests",
  fullyParallel: false,
  retries: 1,
  workers: 1,
  reporter: "list",
  timeout: 60000, // Increased from 30s to 60s for slower operations

  // Auto-start services for E2E tests
  webServer: [
    {
      command: "cd .. && cd backend && go run ./cmd/server",
      port: 8080,
      timeout: 60000,
      reuseExistingServer: false,
      env: {
        DEV_MODE: "true",
      },
    },
    {
      command: "npm run dev",
      port: 3000,
      timeout: 60000,
      reuseExistingServer: true,
    },
  ],

  use: {
    baseURL: "http://localhost:3000",
    trace: "on-first-retry",
    screenshot: "only-on-failure",
    // Add longer default timeout for UI actions
    actionTimeout: 10000,
  },

  projects: [
    {
      name: "chromium",
      use: { ...devices["Desktop Chrome"] },
    },
  ],
});
