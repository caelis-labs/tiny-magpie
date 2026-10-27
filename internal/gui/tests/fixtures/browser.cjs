const { after } = require('node:test');
const { chromium, webkit } = require('playwright');

// Node runs each test file in its own process. Share its browser processes,
// never its cookies, local storage, routes, pages or fixture state.
const browsers = new Map();
after(async () => {
  await Promise.all([...browsers.values()].map(async pending => {
    const browser = await pending;
    await browser.close();
  }));
});

async function pageFor(t, engine, options, beforeClose) {
  if (!['chromium', 'webkit'].includes(engine)) throw new Error(`Unknown browser: ${engine}`);
  if (!browsers.has(engine)) {
    // Use the pinned headless shell consistently with Tiny's core tests.
    browsers.set(engine, { chromium, webkit }[engine].launch());
  }
  const browser = await browsers.get(engine);
  const context = await browser.newContext(options);
  let page;
  // Node stops running subsequent after hooks when one throws. Keep optional
  // failure diagnostics and closure in the same hook so a broken artifact
  // directory cannot leave a context in the shared browser.
  t.after(async () => {
    try {
      if (page && beforeClose) await beforeClose(page);
    } finally {
      await context.close();
    }
  });
  page = await context.newPage();
  return page;
}

module.exports = { pageFor };
