(() => {
 'use strict';
 let target='local', state=null;
 const escv=value=>esc(String(value ?? ''));
 function renderPool() {
  if(S.user?.role!=='admin') return '';
  return `<section class="card" id="np-port-pool" style="margin-bottom:18px"><h2>Member port pool</h2><p class="mu">Members receive a random free port from their hosting node’s pool, up to their server allowance. Existing allocations stay unchanged. The node must also allow member port changes.</p><div id="np-port-pool-form"></div></section>`;
 }
 const key=()=>`${S.user?.id}|${S.token||''}|${target}`;
 function current(st) {return state===st && st.key===key() && S.user?.role==='admin'}
 function draw() {
  const host=document.getElementById('np-port-pool-form'),st=state;if(!host||!st)return;
  const nodes=[{id:'local',name:'Local node'},...(S.nodes||[]).filter(n=>n.id!=='local')];
  host.innerHTML=`<div class="np-feature-grid"><label class="f"><span>Hosting node</span><select id="port-pool-node" ${st.saving?'disabled':''}>${nodes.map(n=>`<option value="${escv(n.id)}" ${target===n.id?'selected':''}>${escv(n.name||n.id)}</option>`).join('')}</select></label><div class="grid g2" style="gap:12px"><label class="f"><span>First port</span><input id="port-pool-start" type="number" min="1024" max="65535" value="${escv(st.start)}" ${st.loading||st.saving?'disabled':''}></label><label class="f"><span>Last port</span><input id="port-pool-end" type="number" min="1024" max="65535" value="${escv(st.end)}" ${st.loading||st.saving?'disabled':''}></label></div></div><div class="row" style="gap:12px"><button class="btn" id="port-pool-save" ${st.loading||st.saving?'disabled':''}>${st.saving?'Saving…':'Save pool'}</button><span class="mu" role="status">${escv(st.message||'')}</span></div>`;
  host.querySelector('#port-pool-node').addEventListener('change',e=>{target=e.target.value;state=null;mount()});
  for(const field of ['start','end'])host.querySelector('#port-pool-'+field).addEventListener('input',e=>{st[field]=e.target.value});
  host.querySelector('#port-pool-save').addEventListener('click',savePool);
  if(typeof enhanceSelects==='function')enhanceSelects(host);
 }
 function mount() {
  if(!document.getElementById('np-port-pool-form')||S.user?.role!=='admin')return;
  if(!state||state.key!==key()) {state={key:key(),url:nodeApiPath(target,'/api/admin/port-pool'),start:'',end:'',loading:true,message:'Loading…'};void load(state)}
  draw();
 }
 async function load(st) {
  try{const p=await api(st.url,{suppressToast:true});if(!current(st))return;if(!Number.isInteger(p?.start)||!Number.isInteger(p?.end))throw Error('The node did not return a port pool.');st.start=p.start;st.end=p.end;st.message='';}
  catch(e){if(current(st))st.message=e.message||'Choose a port range before members can add ports.'}
  finally{if(current(st)){st.loading=false;draw()}}
 }
 async function savePool() {
  const st=state;if(!st||!current(st)||st.loading||st.saving)return;
  const start=Number(st.start),end=Number(st.end);
  if(!Number.isInteger(start)||!Number.isInteger(end)||start<1024||end>65535||end<start){st.message='Choose a range within 1024–65535.';draw();return}
  st.saving=true;draw();
  try{const p=await api(st.url,{method:'PUT',body:JSON.stringify({start,end}),suppressToast:true});if(!current(st))return;if(p?.start!==start||p?.end!==end)throw Error('Save was not confirmed. Reload before retrying.');st.message='Port pool saved';}
  catch(e){if(current(st))st.message=e.message||'Port pool could not be saved.'}
  finally{if(current(st)){st.saving=false;draw()}}
 }
 async function saveAllowance() {
  if(S.user?.role!=='admin')return;
  const srv=curServer(),input=document.getElementById('port-allowance');if(!srv||!input||input.disabled)return;
  const allowed=Number(input.value),url=srvApiPath(srv,'/ports/allowance');
  if(!Number.isInteger(allowed)||allowed<1){toast('Enter a valid total allowance.','bad');return}
  input.disabled=true;
  try{const summary=await api(url,{method:'PUT',body:JSON.stringify({allowed})});if(!curServer()||srvApiPath(curServer(),'/ports/allowance')!==url)return;if(!Number.isInteger(summary?.allowed)||!Number.isInteger(summary?.used))throw Error('Allowance was not confirmed.');srv.ports_allowed=summary.allowed;srv.ports_used=summary.used;toast('Port allowance saved','ok');R()}
  catch(e){toast(e.message||'Could not save allowance','bad')}
  finally{if(input.isConnected)input.disabled=false}
 }
 function reset(){state=null;target='local'}
 window.NPPorts={renderPool,mount,saveAllowance,reset};
})();
