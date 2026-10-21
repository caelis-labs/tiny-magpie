const assert = require('node:assert/strict');
const {test} = require('node:test');
const fs = require('node:fs/promises');
const path = require('node:path');
const {chromium, webkit} = require('playwright');
const {fixture} = require('./fixtures/caller-keys.cjs');

for (const [engine,driver] of [['chromium',chromium],['webkit',webkit]]) {
 for (const width of [440,360]) {
  test(`${engine} ${width}: tray retains available allowances and readable timing`,async t=>{
   const browser=await driver.launch();t.after(()=>browser.close());
   const page=await browser.newPage({viewport:{width,height:800},reducedMotion:'reduce'});
   let available=false;
   const base=fixture('zh','light',[],{web:false,usageTiming:{decode_ms:1000,decode_out:592,timed:1,ttft_ms:164}});
   await page.route('**/*',route=>new URL(route.request().url()).pathname==='/api/usage/quotas'
    ?route.fulfill({json:available?[{provider:'codex',name:'ChatGPT fixture',user:'fixture@example.test',windows:[{name:'5 hours',used:28,resetsAt:new Date(Date.now()+3600000).toISOString()}]}]:[]})
    :base(route));
   await page.goto('http://magpie.test/?mode=panel');
   await page.waitForFunction(()=>Array.isArray(quotas));
   const quotaTab=page.locator('#ptabs [data-ptab=usage]');
   assert.equal(await quotaTab.isVisible(),false,'upstream hides allowances only while no card is available');
   available=true;
   await page.evaluate(()=>loadQuotas());
   await quotaTab.waitFor({state:'visible'});await quotaTab.click();
   await page.locator('#panelQuota .pq-card').first().waitFor({state:'visible'});
   assert.equal(await page.locator('#ptabs [data-ptab=agents]').isVisible(),true);
   assert.ok((await page.locator('#panelQuota').textContent()).includes('ChatGPT fixture'));
   await page.locator('#ptabs [data-ptab=stats]').click();
   await page.waitForFunction(()=>document.querySelector('#panelUsage .pu-tot')?.textContent.includes('592'));
   assert.equal(await page.locator('.pu-tot .blk').count(),6);
   assert.ok(await page.locator('.pu-tot .v').evaluateAll(xs=>xs.every(e=>e.scrollWidth<=e.clientWidth+1)),'tray totals must not overflow their cards');
  });
 }
 for (const locale of ['en','zh','zh-TW','ja','de']) for (const width of [1100,440]) {
  test(`${engine} ${locale} ${width}: explicit classification and retrieval APIs`, async t=>{
   const browser=await driver.launch();t.after(()=>browser.close());
   const page=await browser.newPage({viewport:{width,height:820},reducedMotion:'reduce'});
   await page.addInitScript(()=>Object.defineProperty(navigator,'platform',{value:'MacIntel'}));
   const errors=[],writes=[];page.on('pageerror',e=>errors.push(e.message));
   const base=fixture(locale,'light',writes,{web:false});
   await page.route('**/*',async route=>{
    if(new URL(route.request().url()).pathname==='/api/providers')return route.fulfill({json:{presets:[],excluded:[],plugins:[],providers:[
     {id:'typesafe',name:'TypeSafe Jev',icon:'generic',key:{set:true},ready:true,decide:'http://fixture/v1',models:[{id:'jev-latest',name:'Jev',on:true}],agents:[]},
     {id:'relay',name:'Relay',icon:'generic',key:{set:true},ready:true,chat:'http://fixture/v1',models:[{id:'chat',name:'Chat',on:true}],retrievals:[{id:'vector',name:'Vector',kind:'embedding'},{id:'rank',name:'Rank',kind:'rerank'}],agents:[]}
    ],gateway:{running:true,mine:true,url:'http://127.0.0.1:3999',calls:[],groups:[]}}});
    return base(route);
   });
   await page.goto('http://magpie.test/?view=gateway&lang=curl');
   await page.locator('#connect').waitFor({state:'visible'});
   for (const [name,path,id] of [['System One','systemone','typesafe/jev-latest'],['Embedding','embeddings','relay/vector'],['Rerank','rerank','relay/rank']]) {
    await page.locator('#connect .opt').filter({hasText:new RegExp('^'+name+'$')}).click();
    await page.waitForFunction(({path,id})=>document.querySelector('#connect').textContent.includes('/v1/'+path)&&document.querySelector('#connect').textContent.includes(id),{path,id});
    assert.ok(await page.locator('#connect .hint').allTextContents().then(xs=>xs.some(x=>x.includes('/v1/'+path))), 'endpoint hint must match selected API');
   }
   assert.ok(await page.locator('#view-gateway').evaluate(e=>e.scrollWidth<=e.clientWidth+1),'gateway overflows horizontally');
   if(process.env.ARTIFACT_DIR){await fs.mkdir(process.env.ARTIFACT_DIR,{recursive:true});await page.screenshot({path:path.join(process.env.ARTIFACT_DIR,`core-${engine}-${locale}-${width}.png`)});}
   await page.locator('#nav [data-view=plugins]').click();
   await page.locator('#view-plugins').waitFor({state:'visible'});
   assert.equal(await page.getByText('Built-in',{exact:true}).count(),0);
   await page.locator('#nav [data-view=routing]').click();
   await page.locator('.rt-gsec').waitFor({state:'visible'});
   const groupTop=await page.locator('.rt-gsec').boundingBox();
   assert.ok(groupTop.y<350,'groups must be reachable before the live diagram');
   await page.locator('.rt-gnew').click();
   const advanced=page.locator('.tiny-group-advanced');
   assert.equal(await advanced.getAttribute('open'),null);
   assert.equal(await advanced.locator('.inner').isVisible(),false);
   await advanced.locator('summary').click();
   await advanced.locator('.inner').waitFor({state:'visible'});
   assert.deepEqual(errors,[]);assert.deepEqual(writes,[]);
  });
 }
 for (const theme of ['light','dark']) {
  test(`${engine} ${theme}: tray bird stays small at 150 percent`,async t=>{
   const browser=await driver.launch();t.after(()=>browser.close());
   const page=await browser.newPage({viewport:{width:440,height:720},reducedMotion:'reduce'});
   await page.addInitScript(()=>Object.defineProperty(navigator,'platform',{value:'MacIntel'}));
   await page.route('**/*',fixture('zh',theme,[],{web:false}));
   await page.goto('http://magpie.test/?mode=panel');
   await page.locator('.brand .logo svg').waitFor({state:'visible'});
   await page.evaluate(()=>document.documentElement.style.zoom='1.5');
   const bird=await page.locator('.brand .logo svg').boundingBox();
   assert.ok(bird.width<=28&&bird.height<=28,JSON.stringify(bird));
   const header=await page.locator('.top').boundingBox();
   assert.ok(bird.y>=header.y&&bird.y+bird.height<=header.y+header.height+1,'bird escapes titlebar');
  });
 }
}
