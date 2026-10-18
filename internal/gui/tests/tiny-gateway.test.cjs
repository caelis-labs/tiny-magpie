const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const path = require('node:path');
const { test } = require('node:test');
const { chromium, webkit } = require('playwright');
const { fixture } = require('./fixtures/caller-keys.cjs');

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ['chromium', 'webkit'])) {
  for (const lang of ['en', 'zh', 'zh-TW', 'ja', 'de']) for (const width of [1100, 440]) {
    test(`${engine} ${lang} ${width}: tiny gateway pages and retained controls`, async t => {
      const browser = await (engine === 'webkit' ? webkit.launch() : chromium.launch({channel: 'chromium'}));
      t.after(() => browser.close());
      const page = await browser.newPage({viewport: {width, height: 820}, reducedMotion: 'reduce'});
      page.setDefaultTimeout(8000);
      const errors = [], mutations = [];
      page.on('pageerror', err => errors.push(err.message));
      await page.route('**/*', fixture(lang, 'dark', mutations, {web:false}));
      await page.goto('http://magpie.test/');
      await page.locator('#view-providers').waitFor({state:'visible'});
      await page.waitForFunction(() => document.querySelector('#providers')?.textContent.includes('Relay'));
      for (const view of ['agents','library','sessions']) {
        assert.equal(await page.locator(`#nav [data-view="${view}"]`).isVisible(), false, view);
      }
      for (const view of ['gateway','routing','usage','plugins']) {
        await page.locator(`#nav [data-view="${view}"]`).click();
        await page.locator(`#view-${view}`).waitFor({state:'visible'});
      }
      assert.equal(await page.locator('#usageTab').textContent().then(s => /Sessions|Context|会话|上下文/.test(s)), false);
      await page.locator('#prefs').click();
      await page.locator('#view-settings').waitFor({state:'visible'});
      assert.equal(await page.locator('#keepAwakeRow').isVisible(), false);
      assert.equal(await page.locator('#setPage-sync').isVisible(), false);
      assert.equal(await page.locator('#setTab-sync').isVisible(), false);
      assert.equal(await page.locator('#status.err').isVisible(), false);
      // Applying freshly fetched built-ins must never resurrect removed pages.
      await page.evaluate(() => applyFeatures([]));
      assert.equal(await page.locator('#nav [data-view="sessions"]').isVisible(), false);
      if (process.env.ARTIFACT_DIR) {
        await fs.mkdir(process.env.ARTIFACT_DIR, {recursive:true});
        await page.screenshot({path:path.join(process.env.ARTIFACT_DIR, `tiny-${engine}-${lang}-${width}.png`)});
      }
      assert.deepEqual(errors, []);
      assert.deepEqual(mutations, []);
    });
  }
}
