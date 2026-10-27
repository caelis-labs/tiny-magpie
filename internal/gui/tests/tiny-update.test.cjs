const assert = require('node:assert/strict');
const fs = require('node:fs/promises');
const path = require('node:path');
const { test } = require('node:test');
const { pageFor } = require('./fixtures/browser.cjs');
const { fixture } = require('./fixtures/caller-keys.cjs');

// The app intentionally preserves the reader's scroll against programmatic
// scrolling. Use wheel events, as its other settings tests and a user do.
async function showRow(page,row) {
 await page.locator('#view-settings').hover();
 for(let i=0;i<40;i++){
  const b=await row.boundingBox(),v=await page.locator('#view-settings').boundingBox();
  if(b&&v&&b.y>=v.y+6&&b.y+b.height<=v.y+v.height-6)return;
  await page.mouse.wheel(0,b&&v&&b.y<v.y?-160:160);
  await page.waitForTimeout(30);
 }
 throw new Error('settings row could not be reached by scrolling');
}

for (const engine of ['chromium', 'webkit']) {
 for (const lang of ['en','zh','zh-TW','ja','de']) for (const width of (lang === 'en' ? [1100,440] : [440])) {
  test(`${engine} ${lang} ${width}: own update controls, check, download and confirmation`, async t=>{
   const page=await pageFor(t, engine, {viewport:{width,height:820},reducedMotion:'reduce'});
   page.setDefaultTimeout(10000);
   const errors=[],saves=[],calls=[];page.on('pageerror',e=>{ if(!/^ResizeObserver loop completed with undelivered notifications/.test(e.message))errors.push(e.message); });
   let updateState='found',pref={lang,theme:'light',version:'0.1.3',updateChannel:'preview'};
   const base=fixture(lang,'light',[],{web:false});
   await page.route('**/*',async route=>{
    const req=route.request(),u=new URL(req.url());
    if(u.pathname==='/api/settings') {
     if(req.method()==='POST'){pref={...pref,...req.postDataJSON()};saves.push(req.postDataJSON());}
     return route.fulfill({json:pref});
    }
    if(u.pathname==='/api/state')return route.fulfill({json:{agents:[],profiles:[],settings:pref}});
    if(u.pathname.startsWith('/api/update')){
     if(req.method()==='POST')calls.push(u.pathname);
     if(u.pathname==='/api/update/install')updateState='ready';
     if(u.pathname==='/api/update/notes')return route.fulfill({json:{latest:'0.1.4',releases:[{version:'0.1.4',notes:'Tiny fixture release',url:'https://github.com/caelis-labs/tiny-magpie/releases/tag/tiny-v0.1.4'}]}});
     return route.fulfill({json:{state:updateState,current:'0.1.3',latest:'0.1.4',notes:'Tiny fixture release',url:'https://github.com/caelis-labs/tiny-magpie/releases/tag/tiny-v0.1.4'}});
    }
    return base(route);
   });
   await page.goto('http://tiny.test/?view=providers');
   await page.locator('#update').waitFor({state:'visible'});
   await page.locator('#update').click();
   const box=page.locator('#tinyUpdates');
   await box.locator('.row').first().getByText('0.1.4',{exact:false}).waitFor();
   assert.equal(calls.length,0,'opening settings downloaded an update');
   // Explicit checking never posts install, regardless of automatic settings.
   await showRow(page,box.locator('.row').first());
   await box.locator('.row').first().locator('button').nth(1).click();
   assert.deepEqual(calls,['/api/update/check']);
   // Downloading stages it, but a separate Restart action and dialog are required.
   await box.locator('.row').first().locator('button').first().click();
   await page.waitForFunction(()=>document.querySelector('#tinyUpdates .row')?.textContent.match(/downloaded|已下载|已下載|ダウンロードしました|heruntergeladen/));
   assert.deepEqual(calls,['/api/update/check','/api/update/install']);
   await box.locator('.row').first().locator('button').first().click();
   await page.getByText('Tiny fixture release',{exact:false}).waitFor();
   assert.equal(calls.filter(x=>x==='/api/update/install').length,1,'restart happened before confirmation');
   await page.keyboard.press('Escape');
   // Both preferences survive a settings save/reload without reviving removed pages.
   await showRow(page,box.locator('.row').nth(1));
   await box.locator('.row').nth(1).locator('button').first().click();
   await page.waitForFunction(()=>document.querySelector('#tinyUpdates .row:nth-child(2) button')?.getAttribute('aria-pressed')==='true');
   await showRow(page,box.locator('.row').nth(2));
   await box.locator('.row').nth(2).locator('button').last().click();
   await page.waitForFunction(()=>prefs.updateChannel==='stable');
   assert.equal(saves.at(-1).updateChannel,'stable');
   assert.equal(saves.at(-1).noAutoUpdate,true);
   await page.reload();await box.locator('.row').nth(2).waitFor();
   assert.equal(await box.locator('.row').nth(2).locator('button').last().getAttribute('aria-pressed'),'true');
   if(process.env.ARTIFACT_DIR && lang==='zh' && width===1100){
    await fs.mkdir(process.env.ARTIFACT_DIR,{recursive:true});
    await page.screenshot({path:path.join(process.env.ARTIFACT_DIR,`tiny-updates-${engine}.png`)});
   }
   assert.deepEqual(errors,[]);
  });
 }
}
