import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: 'html',
  use: {
    baseURL: 'http://localhost:34115',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    actionTimeout: 30000,
    navigationTimeout: 60000,
  },
  projects: [
    {
      name: 'webkit',
      use: { ...devices['Desktop WebKit'] },
    },
  ],
  webServer: {
    command: '/Users/pallabpc/go/bin/wails dev -tags webkit2_41',
    url: 'http://localhost:34115',
    reuseExistingServer: true,
    timeout: 180000,
  },
});