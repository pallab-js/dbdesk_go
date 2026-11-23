import { test, expect } from '@playwright/test';

test.describe('DBDesk UI/UX', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForSelector('.topbar', { timeout: 30000 });
    await page.waitForTimeout(2000);
  });

  test('should load without fatal errors', async ({ page }) => {
    const fatal = page.locator('.fatal');
    await expect(fatal).not.toBeVisible({ timeout: 10000 });
  });

  test('should show welcome screen when no database connected', async ({ page }) => {
    await expect(page.locator('.welcome, main').first()).toBeVisible({ timeout: 10000 });
    // Welcome screen should have some content
    await expect(page.locator('h1, h2').first()).toBeVisible({ timeout: 5000 });
  });

  test('should have working theme toggle', async ({ page }) => {
    const themeBtn = page.locator('button[aria-label^="Theme:"]').first();
    await expect(themeBtn).toBeVisible({ timeout: 10000 });
    
    const initialTheme = await themeBtn.getAttribute('aria-label');
    await themeBtn.click();
    await page.waitForTimeout(100);
    const newTheme = await themeBtn.getAttribute('aria-label');
    expect(newTheme).not.toBe(initialTheme);
  });

  test('should open demo database and show sidebar navigation', async ({ page }) => {
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await expect(demoBtn).toBeVisible({ timeout: 10000 });
    await demoBtn.click();
    
    // Wait for sidebar navigation items
    await expect(page.locator('.sidebar')).toBeVisible({ timeout: 20000 });
    await expect(page.locator('.sidebar button:has-text("Dashboard")')).toBeVisible();
    await expect(page.locator('.sidebar button:has-text("Data")')).toBeVisible();
    await expect(page.locator('.sidebar button:has-text("Editor")')).toBeVisible();
    await expect(page.locator('.sidebar button:has-text("History")')).toBeVisible();
    await expect(page.locator('.sidebar button:has-text("Import")')).toBeVisible();
    await expect(page.locator('.sidebar button:has-text("Maintenance")')).toBeVisible();
    await expect(page.locator('.sidebar button:has-text("Settings")')).toBeVisible();
  });

  test('sidebar navigation works', async ({ page }) => {
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('.sidebar')).toBeVisible({ timeout: 20000 });
    
    // Navigate through all views
    await page.locator('.sidebar button:has-text("Dashboard")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
    
    await page.locator('.sidebar button:has-text("Data")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
    
    await page.locator('.sidebar button:has-text("Editor")').click();
    await expect(page.locator('.cm-editor')).toBeVisible({ timeout: 10000 });
    
    await page.locator('.sidebar button:has-text("History")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
    
    await page.locator('.sidebar button:has-text("Import")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
    
    await page.locator('.sidebar button:has-text("Maintenance")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
    
    await page.locator('.sidebar button:has-text("Settings")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
  });

  test('dashboard shows stats after demo load', async ({ page }) => {
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    
    // Wait for dashboard view
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    // Check for key dashboard metrics
    await expect(page.locator('text=Database Size')).toBeVisible({ timeout: 15000 });
    await expect(page.locator('.kpi-label:has-text("Tables")')).toBeVisible();
  });

  test('data browser shows tables list', async ({ page }) => {
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    await page.locator('.sidebar button:has-text("Data")').click();
    await expect(page.locator('main')).toBeVisible({ timeout: 10000 });
    
    // Should show table list - look for any table name
    await expect(page.locator('text=mines').first()).toBeVisible({ timeout: 15000 });
    await expect(page.locator('text=plants').first()).toBeVisible();
    await expect(page.locator('text=products').first()).toBeVisible();
  });

  test('editor can run query', async ({ page }) => {
    test.setTimeout(60000);
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    await page.locator('.sidebar button:has-text("Editor")').click();
    // Editor takes time to load CodeMirror
    await page.waitForTimeout(3000);
    await expect(page.locator('.editor-host')).toBeVisible({ timeout: 40000 });
    
    await page.locator('.editor-host').click();
    await page.waitForTimeout(500);
    await page.keyboard.type('SELECT * FROM mines LIMIT 5');
    await page.locator('button:has-text("Run")').click();
    
    // Should show results
    await expect(page.locator('text=mines').first()).toBeVisible({ timeout: 20000 });
  });

  test('editor runs multiple query types', async ({ page }) => {
    test.setTimeout(90000);
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    await page.locator('.sidebar button:has-text("Editor")').click();
    await page.waitForTimeout(3000);
    await expect(page.locator('.editor-host')).toBeVisible({ timeout: 40000 });
    
    // Helper to run a query and wait for results
    async function runQuery(sql: string, expectResult: string | RegExp) {
      // Clear editor first
      await page.locator('button:has-text("Clear")').click();
      await page.waitForTimeout(200);
      await page.locator('.editor-host').click();
      await page.waitForTimeout(200);
      await page.keyboard.type(sql);
      await page.locator('button:has-text("Run")').click();
      await expect(page.locator('.result-panel')).toBeVisible({ timeout: 20000 });
      await expect(page.locator('.grid-wrap')).toBeVisible({ timeout: 15000 });
      await expect(page.locator('.grid-wrap')).toContainText(expectResult);
    }
    
    // Helper to run EXPLAIN
    async function runExplain(sql: string, expectResult: string | RegExp) {
      await page.locator('button:has-text("Clear")').click();
      await page.waitForTimeout(200);
      await page.locator('.editor-host').click();
      await page.waitForTimeout(200);
      await page.keyboard.type(sql);
      await page.locator('button:has-text("Explain")').click();
      await expect(page.locator('.result-panel')).toBeVisible({ timeout: 20000 });
      await expect(page.locator('.grid-wrap')).toBeVisible({ timeout: 15000 });
      await expect(page.locator('.grid-wrap')).toContainText(expectResult);
    }
    
    // Test 1: Simple SELECT
    await runQuery('SELECT name, state FROM mines ORDER BY name LIMIT 3', 'Borgang');
    await expect(page.locator('.grid-wrap')).toContainText('Dirang');
    await expect(page.locator('.grid-wrap')).toContainText('Dhansiri');
    
    // Test 2: JOIN query
    await runQuery(`
      SELECT m.name as mine, p.name as plant, p.capacity_tph
      FROM mines m
      JOIN plants p ON p.mine_id = m.id
      WHERE m.state = 'Assam'
    `, 'Numaligarh');
    await expect(page.locator('.grid-wrap')).toContainText('Borgang');
    
    // Test 3: Aggregate query
    await runQuery(`
      SELECT product_id, SUM(quantity_mt) as total_mt
      FROM production
      WHERE plant_id = 1
      GROUP BY product_id
      ORDER BY total_mt DESC
      LIMIT 5
    `, /total_mt/);
    
    // Test 4: View query (stock_status - may return no LOW rows)
    await page.locator('button:has-text("Clear")').click();
    await page.waitForTimeout(200);
    await page.locator('.editor-host').click();
    await page.waitForTimeout(200);
    await page.keyboard.type('SELECT * FROM stock_status LIMIT 10');
    await page.locator('button:has-text("Run")').click();
    await expect(page.locator('.result-panel')).toBeVisible({ timeout: 20000 });
    await expect(page.locator('.grid-wrap')).toBeVisible({ timeout: 15000 });
    // Just verify it runs without error - could show data or "No rows"
    
    // Test 5: EXPLAIN QUERY PLAN
    await runExplain('SELECT * FROM mines WHERE state = \'Assam\'', 'SCAN');
  });

  test('responsive layout on different viewport sizes', async ({ page }) => {
    await page.setViewportSize({ width: 1920, height: 1080 });
    await page.waitForTimeout(500);
    
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    await page.setViewportSize({ width: 1024, height: 768 });
    await page.waitForTimeout(500);
    await expect(page.locator('.topbar')).toBeVisible();
    
    await page.setViewportSize({ width: 1280, height: 800 });
  });

  test('toast notifications appear', async ({ page }) => {
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    await expect(page.locator('.toast, [role="status"]').first()).toBeVisible({ timeout: 10000 });
  });

  test('fullscreen toggle works', async ({ page }) => {
    const demoBtn = page.locator('.topbar button:has-text("Demo")').first();
    await demoBtn.click();
    await expect(page.locator('main')).toBeVisible({ timeout: 20000 });
    
    const fullscreenBtn = page.locator('button[aria-label^="Enter fullscreen"], button[aria-label^="Exit fullscreen"]').first();
    await expect(fullscreenBtn).toBeVisible({ timeout: 10000 });
    
    await fullscreenBtn.click();
    await page.waitForTimeout(500);
    
    await fullscreenBtn.click();
    await page.waitForTimeout(500);
  });
});