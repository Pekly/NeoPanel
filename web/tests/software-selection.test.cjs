const test = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const vm = require('node:vm');
const app = fs.readFileSync(path.join(__dirname, '../static/js/app.js'), 'utf8');
function section(start, end) {
  const a=app.indexOf(start), b=app.indexOf(end,a);
  assert.ok(a>=0 && b>a, start);
  return app.slice(a,b);
}
function setup() {
  const names=['Paper','Fabric','Velocity','Steel'];
  const buttons=names.map(name=>({dataset:{softwareChoice:name},selected:false,pressed:'false',classList:{toggle(key,value){buttons.find(b=>b.dataset.softwareChoice===name).selected=value;}},setAttribute(key,value){this.pressed=value;}}));
  const elements={'#version-choice-title':{},'#version-search':{value:'old query'},'#version-family-filter':{hidden:false},'#version-family':{value:'Popular',_cselectRebuild(){this.rebuilt=true;}}};
  let rows=0,loads=0;
  const c={S:{view:'ver',sel:'Paper',core:'Paper',vq:'old query',vf:'Popular',ver:'1.21.4'},CORES:names.map(n=>[n,'']),document:{querySelectorAll:()=>buttons},$:id=>elements[id],isSteelEngine:n=>n==='Steel',renderVersionRows(){rows++;},loadVersionsCatalog(){loads++;return Promise.resolve();},R(){throw Error('Must not redraw the full page');}};
  vm.runInNewContext(section('function selectServerSoftware(', 'function renderVersionRows('),c);
  return {c,buttons,elements,rows:()=>rows,loads:()=>loads};
}
test('software selection updates the existing chooser without a page redraw or installation',()=>{
  const {c,buttons,elements,rows,loads}=setup();c.selectServerSoftware('Fabric');
  assert.equal(c.S.sel,'Fabric');assert.equal(c.S.core,'Paper');assert.equal(c.S.ver,'1.21.4');
  assert.equal(c.S.vq,'');assert.equal(c.S.vf,'All Families');assert.equal(elements['#version-search'].value,'');
  assert.equal(elements['#version-choice-title'].textContent,'2. Select Fabric Version');
  assert.equal(elements['#version-family'].value,'All Families');assert.equal(elements['#version-family'].rebuilt,true);
  assert.equal(buttons.filter(b=>b.selected).length,1);assert.equal(buttons[1].pressed,'true');
  assert.equal(rows(),1);assert.equal(loads(),1);
  assert.doesNotMatch(section('function selectServerSoftware(', 'function renderVersionRows('),/\bR\(|installSoftware\(|location\.|history\./);
});
test('proxy and native engines toggle only their relevant version controls',()=>{
  const {c,elements}=setup();c.selectServerSoftware('Velocity');assert.equal(elements['#version-family-filter'].hidden,true);
  c.selectServerSoftware('Steel');assert.match(elements['#version-search'].placeholder,/Steel release/);
  c.selectServerSoftware('Paper');assert.equal(elements['#version-family-filter'].hidden,false);assert.match(elements['#version-search'].placeholder,/Minecraft versions/);
});
test('duplicate, invalid, and off-page selections do nothing',()=>{
  const {c,rows,loads}=setup();c.selectServerSoftware('Paper');c.selectServerSoftware('invalid');c.S.view='con';c.selectServerSoftware('Fabric');assert.equal(rows(),0);assert.equal(loads(),0);assert.equal(c.S.sel,'Paper');
});
test('engine cards are non-submitting buttons using the local selection handler',()=>{
  const html=section('  ver() {','  bk() {');assert.match(html,/type="button" class="card sel/);assert.match(html,/onclick="selectServerSoftware\(this.dataset.softwareChoice\)"/);assert.doesNotMatch(html,/loadVersionsCatalog\(\); R\(\)/);
});
test('late catalog completion redraws rows for the latest choice rather than the page',async()=>{
  const {c,elements}=setup();let complete;const task=new Promise(resolve=>complete=resolve);let rendered='';c.renderVersionRows=()=>{rendered=c.S.sel;};c.loadVersionsCatalog=()=>task.then(()=>c.renderVersionRows());
  c.selectServerSoftware('Fabric');c.selectServerSoftware('Velocity');complete();await task;await Promise.resolve();
  assert.equal(rendered,'Velocity');assert.equal(elements['#version-choice-title'].textContent,'2. Select Velocity Version');
  assert.doesNotMatch(section('async function loadSteelVersions(', 'const CATS'),/\bR\(/);
});
test('the header address uses one consistent escaped renderer without a transient Join prefix',()=>{
  const c={esc:s=>String(s).replaceAll('&','&amp;').replaceAll('"','&quot;').replaceAll('<','&lt;'),ic:()=>''};
  vm.runInNewContext(section('function playerAddressButtonHtml(', 'function copyServerAddress('),c);
  const html=c.playerAddressButtonHtml('play.example:25565');assert.doesNotMatch(html,/Join:/);assert.match(html,/Copy player connection address/);assert.match(html,/data-addr="play.example:25565"/);
  assert.doesNotMatch(c.playerAddressButtonHtml('<img>"'),/<img>/);
  assert.equal((app.match(/\$\{playerAddressButtonHtml\(hostAddr\)\}/g)||[]).length,3);
  assert.doesNotMatch(section('function srvh()', 'function toggleCmdMenu()'),/Join:/);
});
test('partial release-list updates retain permission-based disabled controls',()=>{
  const rows={};let protectedRoot;
  const c={$:()=>rows,versionRowsHtml:()=>'<button>Install</button>',applyPermissionControls:root=>{protectedRoot=root;}};
  vm.runInNewContext(section('function renderVersionRows(', '// --- Navigation Render ---'),c);
  c.renderVersionRows();assert.equal(rows.innerHTML,'<button>Install</button>');assert.equal(protectedRoot,rows);
});
