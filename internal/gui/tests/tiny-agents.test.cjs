const assert = require('node:assert/strict');
const {test} = require('node:test');
const { pageFor } = require('./fixtures/browser.cjs');
const {fixture} = require('./fixtures/caller-keys.cjs');
for (const name of ['chromium','webkit']) for (const width of [1100,440]) {
 test(`${name} ${width}: Agent model picks and trimmed details`,async t=>{
  const page=await pageFor(t, name, {viewport:{width,height:820},reducedMotion:'reduce'});page.setDefaultTimeout(8000);
  const errors=[],requests=[];page.on('pageerror',e=>errors.push(e.message));
  const a={id:'codex',name:'Codex',icon:'generic',path:'~/.codex/config.toml',wired:true,fields:[
   {key:'model',label:'model',value:'relay/m',options:[{value:'relay/m',label:'Model A',ref:'relay/m'},{value:'relay/b',label:'Model B',ref:'relay/b'}]},
   {key:'effort',label:'effort',value:'high',options:[{value:'high',label:'high'},{value:'low',label:'low'}]}
  ]};
  const state=()=>({agents:[a],profiles:[],settings:{lang:'en',theme:'light'},served:2});
  const base=fixture('en','light',[],{web:true});
  await page.route('**/*',route=>{
   const req=route.request(),p=new URL(req.url()).pathname;
   if(p.startsWith('/api/agents/')||p==='/api/drift')requests.push(p);
   if(p==='/api/state')return route.fulfill({json:state()});
   if(p==='/api/set'){const b=req.postDataJSON();requests.push(b);a.fields.find(f=>f.key===b.field).value=b.value;return route.fulfill({json:state()});}
   if(p==='/api/drift')return route.fulfill({json:{}});
   if(p.startsWith('/api/agents/preview/'))return route.fulfill({json:{changes:[{path:'~/.codex/config.toml',lines:[{op:'~',text:'model = "original"'}]}]}});
   return base(route);
  });
  await page.goto('http://magpie.test/?view=agents');
  const row=page.locator('#agents .row.agent[data-id=codex]');await row.waitFor({state:'visible'});
  await row.locator('[data-key=model]').click();
  await page.locator('#list li[data-i]').filter({hasText:'Model B'}).click();
  await page.waitForFunction(()=>document.querySelector('#agents [data-key=model]')?.textContent.includes('Model B'));
  await row.locator('.ag-link').click();
  await row.locator('[data-key=effort]').click();
  await page.locator('#effortRange').press('End');
  await page.waitForFunction(()=>document.querySelector('#agents [data-key=effort]')?.textContent.toLowerCase().includes('low'));
  assert.equal(await page.locator('.profiles').isVisible(),false);
  assert.equal(await page.locator('.ag-prefs').count(),0);
  assert.equal(await page.locator('.ag-manage-bar').count(),0);
  assert.ok((await row.textContent()).includes('Disconnect and restore'));
  assert.ok(requests.some(x=>x.field==='model'&&x.value==='relay/b'));
  assert.ok(requests.some(x=>x.field==='effort'&&x.value==='low'));
  assert.equal(requests.some(x=>typeof x==='string'&&/\/cli|\/install/.test(x)),false);
  assert.ok(await page.locator('#view-agents').evaluate(e=>e.scrollWidth<=e.clientWidth+1));
  assert.deepEqual(errors,[]);
 });
}
