const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const path = require('node:path');
const { test } = require('node:test');
const { pageFor } = require('./fixtures/browser.cjs');
const { fixture } = require('./fixtures/caller-keys.cjs');

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ['chromium', 'webkit'])) {
  for (const lang of ['en', 'zh', 'zh-TW', 'ja', 'de']) for (const width of (lang === 'en' ? [1100, 560, 440] : [440])) {
    // Native windows stop at 560px (windowMinW); 440px exercises the Web phone layout.
    const web = width < 560;
    test(`${engine} ${lang} ${width} ${web ? 'Web' : 'desktop'}: tiny gateway pages and retained controls`, async t => {
      // Use the same pinned headless Chromium as the other retained suites.
      let page, completed = false;
      const errors = [], mutations = [];
      const diagnose = async () => {
        if (!completed && page && process.env.ARTIFACT_DIR) {
          await fs.mkdir(process.env.ARTIFACT_DIR, {recursive:true});
          let diagnosticTimer;
          const geometry = await Promise.race([page.evaluate(() => ({
            view: document.querySelector('.view:not([hidden])')?.id,
            header: document.querySelector('.top')?.className,
            visibility: document.visibilityState,
            tabs: [...document.querySelectorAll('#nav button:not([hidden])')].map(b => ({
              view:b.dataset.view, x:b.getBoundingClientRect().x, width:b.getBoundingClientRect().width,
            })),
          })), new Promise(resolve => { diagnosticTimer = setTimeout(() => resolve({unresponsive:true}), 2000); })]).finally(() => clearTimeout(diagnosticTimer));
          console.error('Navigation failure', geometry, errors);
          await page.screenshot({path:path.join(process.env.ARTIFACT_DIR, `failed-tiny-${engine}-${lang}-${width}.png`),timeout:2000}).catch(err => console.error('Failure screenshot unavailable:', err.message));
        }
      };
      page = await pageFor(t, engine, {viewport: {width, height: 820}, reducedMotion: 'reduce'}, diagnose);
      await page.addInitScript(() => Object.defineProperty(navigator, 'platform', {value:'MacIntel'}));
      page.setDefaultTimeout(8000);
      page.on('pageerror', err => errors.push(err.message));
      await page.route('**/*', fixture(lang, 'dark', mutations, {web,usageDelay:100,usageTiming:{decode_ms:1000,decode_out:592,timed:1,ttft_ms:164}}));
      await page.goto('http://magpie.test/');
      await page.locator('#view-providers').waitFor({state:'visible'});
      await page.waitForFunction(() => document.querySelector('#providers')?.textContent.includes('Relay'));
      const accountSpacing = await page.evaluate(() => ({
        add: document.querySelector('#view-providers .after-list').getBoundingClientRect().bottom,
        card: document.querySelector('#discoveryList').getBoundingClientRect().top,
      }));
      assert.ok(accountSpacing.card >= accountSpacing.add + 4, 'subscription card must clear the add-provider gradient');
      const bird = await page.locator('.brand .logo svg').boundingBox();
      assert.ok((width < 1100 && !bird) || bird && bird.width <= 22 && bird.height <= 22, `brand bird bounds: ${JSON.stringify(bird)}`);

      for (const view of ['library','sessions']) {
        assert.equal(await page.locator(`#nav [data-view="${view}"]`).isVisible(), false, view);
      }
      for (const view of ['gateway','routing','usage','plugins']) {
        await page.locator(`#nav [data-view="${view}"]`).click();
        await page.locator(`#view-${view}`).waitFor({state:'visible'});
      }
      assert.equal(await page.locator('#usageTab').textContent().then(s => /Sessions|Context|会话|上下文/.test(s)), false);
      await page.locator('#nav [data-view="usage"]').click();
      await page.locator('#stats .kpi b').filter({hasText:'592'}).waitFor({state:'visible'});
      assert.equal(await page.locator('#stats .kpi').count(), 6, 'overview retains tokens, cache, reasoning, calls, speed and first-token timing');
      assert.ok(await page.locator('#stats .kpi b').evaluateAll(xs => xs.every(e => e.scrollWidth <= e.clientWidth + 1)), 'summary values and timing units must not be clipped');
      assert.equal(await page.locator('#usageCost').isVisible(), true, 'estimated cost must stay visible');
      await page.locator(':is(#usageTab .opt, .nav-s):visible').nth(1).click();
      await page.locator('#ledKpi .blk').first().waitFor({state:'visible'});
      assert.equal(await page.locator('#ledScope').textContent().then(s => /session|会话|會話|セッション|Sitzungs/i.test(s)), false);
      await page.locator('#prefs').click();
      await page.locator('#view-settings').waitFor({state:'visible'});
      assert.equal(await page.locator('#keepAwakeRow').isVisible(), false);
      assert.equal(await page.locator('#setPage-sync').isVisible(), false);
      assert.equal(await page.locator('#setTab-sync').isVisible(), false);
      for (const id of ['privacy','models','usage','about']) assert.equal(await page.locator('#setTab-' + id).isVisible(), false, id);
      assert.equal(await page.locator('#uiFontRow').isVisible(), false);
      assert.equal(await page.locator('#codeFontRow').isVisible(), false);

      assert.equal(await page.locator('#status.err').isVisible(), false);
      // Applying freshly fetched built-ins must never resurrect removed pages.
      await page.evaluate(() => applyFeatures([]));
      assert.equal(await page.locator('#nav [data-view="sessions"]').isVisible(), false);

      assert.deepEqual(errors, []);
      assert.ok(mutations.some(e => e.action === 'ledger'), 'requests must load the retained gateway ledger');
      assert.deepEqual(mutations.filter(e => e.action !== 'ledger'), []);
      completed = true;
    });
  }
}

for (const engine of ['chromium', 'webkit']) {
  test(`${engine}: occupied gateway stays separate (Web, light theme)`, async t => {
    const page = await pageFor(t, engine, {viewport:{width:1100,height:820}});
    const mutations = [], errors = [];
    page.on('pageerror',e=>errors.push(e.message));
    await page.route('**/*',fixture('en','light',mutations,{web:true,taken:true}));
    await page.goto('http://magpie.test/?view=gateway');
    await page.getByText('Port 3999 is in use by another program: quit it, or move the gateway to another port in Settings.', {exact:true}).waitFor();
    assert.equal(await page.locator('#gateway .dot.on').count(),0);
    assert.equal(await page.locator('#gateway .gw-fix').count(),0);
    assert.deepEqual(errors,[]);
    assert.deepEqual(mutations,[]);
  });
}
