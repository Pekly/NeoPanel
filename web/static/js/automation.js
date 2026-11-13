// Backup recovery and ordered schedule editing. Backend remains authoritative.
const AutomationUI = { tasks: [], server: null, scheduleID: '', saving: false, restoring: null };
const SCHEDULE_ACTIONS = [['command', 'Console command'], ['backup', 'Create backup'], ['stop', 'Stop server'], ['start', 'Start server'], ['restart', 'Restart server']];
function scheduleTasksFor(schedule) {
  const tasks = Array.isArray(schedule?.tasks) && schedule.tasks.length ? schedule.tasks : [{ action: schedule?.action || 'command', payload: schedule?.payload || '', time_offset: 0, continue_on_failure: false }];
  return tasks.map(task => ({ action: task.action, payload: task.payload || '', time_offset: Number(task.time_offset || 0), continue_on_failure: Boolean(task.continue_on_failure) }));
}
function validateScheduleTasks(tasks) {
  if (!Array.isArray(tasks) || tasks.length < 1 || tasks.length > 20) throw new Error('A schedule must contain between 1 and 20 steps.');
  return tasks.map((task, index) => {
    if (!SCHEDULE_ACTIONS.some(([action]) => action === task.action)) throw new Error(`Step ${index + 1}: choose a supported action.`);
    const offset = Number(task.time_offset);
    if (!Number.isInteger(offset) || offset < 0 || offset > 3600) throw new Error(`Step ${index + 1}: delay must be a whole number from 0 to 3600 seconds.`);
    const payload = String(task.payload || '').trim();
    if (task.action === 'command' && (!payload || new TextEncoder().encode(payload).length > 4096)) throw new Error(`Step ${index + 1}: enter a command of 1–4096 UTF-8 bytes.`);
    return { action: task.action, payload: task.action === 'command' ? payload : '', time_offset: offset, continue_on_failure: Boolean(task.continue_on_failure) };
  });
}
function renderScheduleTaskSummary(schedule) {
  const tasks = scheduleTasksFor(schedule);
  if (tasks.length === 1) return getScheduleActionBadge(tasks[0].action);
  return `<details class="np-task-summary"><summary>${tasks.length} steps</summary><ol>${tasks.map(task => `<li>${esc(SCHEDULE_ACTIONS.find(([action]) => action === task.action)?.[1] || task.action)}${task.time_offset ? ` · wait ${task.time_offset}s first` : ''}${task.payload ? `<div class="mono mu">${esc(task.payload)}</div>` : ''}${task.continue_on_failure ? '<span class="mu">Continue on failure</span>' : ''}</li>`).join('')}</ol></details>`;
}
function updateScheduleTask(index, field, value) {
  const task = AutomationUI.tasks[index];
  if (!task) return;
  task[field] = field === 'time_offset' ? Number(value) : value;
  if (field === 'action') renderScheduleTaskRows();
}
function addScheduleTask() {
  if (AutomationUI.tasks.length >= 20) return;
  AutomationUI.tasks.push({ action: 'command', payload: '', time_offset: 0, continue_on_failure: false });
  renderScheduleTaskRows();
}
function removeScheduleTask(index) {
  if (AutomationUI.tasks.length <= 1) return;
  AutomationUI.tasks.splice(index, 1); renderScheduleTaskRows();
}
function moveScheduleTask(index, direction) {
  const next = index + direction;
  if (next < 0 || next >= AutomationUI.tasks.length) return;
  [AutomationUI.tasks[index], AutomationUI.tasks[next]] = [AutomationUI.tasks[next], AutomationUI.tasks[index]];
  renderScheduleTaskRows();
}
function renderScheduleTaskRows() {
  const root = document.querySelector('#schedule-task-list');
  if (!root) return;
  closeAllCustomSelects();
  root.innerHTML = AutomationUI.tasks.map((task, index) => `<div class="np-schedule-step" data-step="${index}"><div class="np-team-heading"><b>Step ${index + 1}</b><div class="row"><button type="button" class="btn sm" ${index === 0 ? 'disabled' : ''} onclick="moveScheduleTask(${index}, -1)" aria-label="Move step ${index + 1} up">↑</button><button type="button" class="btn sm" ${index === AutomationUI.tasks.length - 1 ? 'disabled' : ''} onclick="moveScheduleTask(${index}, 1)" aria-label="Move step ${index + 1} down">↓</button><button type="button" class="btn sm dg" ${AutomationUI.tasks.length < 2 ? 'disabled' : ''} onclick="removeScheduleTask(${index})">Remove</button></div></div>
    <div class="grid g2"><label class="f"><span>Action</span><select onchange="updateScheduleTask(${index}, 'action', this.value)">${SCHEDULE_ACTIONS.map(([action, label]) => `<option value="${action}" ${task.action === action ? 'selected' : ''}>${label}</option>`).join('')}</select></label><label class="f"><span>Delay before step (seconds)</span><input type="number" min="0" max="3600" step="1" required value="${task.time_offset}" oninput="updateScheduleTask(${index}, 'time_offset', this.value)"></label></div>
    ${task.action === 'command' ? `<label class="f"><span>Command</span><input required maxlength="4096" class="mono" value="${esc(task.payload)}" placeholder="say Server restarting in 5 minutes!" oninput="updateScheduleTask(${index}, 'payload', this.value)"></label>` : ''}
    <label class="row np-continue-step"><input type="checkbox" ${task.continue_on_failure ? 'checked' : ''} onchange="updateScheduleTask(${index}, 'continue_on_failure', this.checked)"><span>Continue to the next step if this one fails</span></label></div>`).join('');
  const add = document.querySelector('#schedule-add-step'); if (add) add.disabled = AutomationUI.tasks.length >= 20;
  const count = document.querySelector('#schedule-step-count'); if (count) count.textContent = `${AutomationUI.tasks.length} / 20 steps`;
  enhanceSelects(root);
}
function openScheduleChainModal(scheduleID = null) {
  const srv = curServer(); if (!srv) return;
  const existing = S.serverSchedules?.find(s => String(s.id) === String(scheduleID));
  AutomationUI.server = { ...srv }; AutomationUI.scheduleID = existing?.id || ''; AutomationUI.tasks = scheduleTasksFor(existing);
  document.querySelector('#modal-container').innerHTML = `<div class="pal" onclick="if(event.target===this&&!AutomationUI.saving)closeModal()"><div class="card pbox np-schedule-modal" role="dialog" aria-modal="true" aria-labelledby="schedule-chain-title"><div class="np-team-heading"><h2 id="schedule-chain-title">${existing ? 'Edit schedule' : 'Create schedule'}</h2><button class="btn sm" onclick="closeModal()" aria-label="Close schedule">×</button></div><p class="mu">Run ordered steps on <b>${esc(srv.name)}</b>. Times are UTC. Delays occur before each step; failed runs remain visible even when later steps continue.</p>
    <form id="schedule-chain-form" onsubmit="submitScheduleModal(event, '${esc(String(AutomationUI.scheduleID))}')"><label class="f"><span>Schedule name</span><input id="sch-name" required maxlength="120" value="${esc(existing?.name || '')}" placeholder="Nightly maintenance"></label>
    <div class="grid g2"><label class="f"><span>Frequency</span><select id="sch-frequency" onchange="applyScheduleTimePicker()"><option value="custom">Custom cron</option><option value="daily">Daily</option><option value="hourly">Hourly (minute from time)</option><option value="weekly">Weekly</option></select></label><label class="f"><span>Time (UTC)</span><input id="sch-time" type="time" value="04:00" onchange="applyScheduleTimePicker()"></label><label class="f"><span>Weekly day</span><select id="sch-weekday" onchange="applyScheduleTimePicker()">${['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'].map((day, index) => `<option value="${index}">${day}</option>`).join('')}</select></label></div>
    <label class="f"><span>Cron expression</span><input id="sch-cron" required class="mono" value="${esc(existing?.cron_expr || '0 4 * * *')}" oninput="document.querySelector('#sch-frequency').value='custom';updateSchedulePreview()"></label><p id="sch-preview" class="mu" role="status"></p>
    <div class="np-team-heading"><h3>Execution steps</h3><span id="schedule-step-count" class="mu"></span></div><div id="schedule-task-list"></div><button id="schedule-add-step" type="button" class="btn sm" onclick="addScheduleTask()">${ic('plus', 13)} Add step</button>
    <label class="row np-continue-step"><input id="sch-active" type="checkbox" ${existing ? existing.is_active ? 'checked' : '' : 'checked'}><span>Enable schedule</span></label><p id="schedule-chain-error" class="upload-error" role="alert"></p><div class="row" style="justify-content:flex-end"><button type="button" class="btn" onclick="closeModal()">Cancel</button><button id="sch-submit-btn" type="submit" class="btn pr">${existing ? 'Save schedule' : 'Create schedule'}</button></div></form></div></div>`;
  renderScheduleTaskRows(); updateSchedulePreview(); document.querySelector('#sch-name').focus();
}
async function submitScheduleChain(event) {
  event.preventDefault();
  const srv = AutomationUI.server;
  if (AutomationUI.saving || !srv || !sameQolServer(srv, curServer())) return;
  const form = document.querySelector('#schedule-chain-form'); if (!form?.reportValidity()) return;
  const error = document.querySelector('#schedule-chain-error');
  let tasks;
  try { tasks = validateScheduleTasks(AutomationUI.tasks); }
  catch (err) { error.textContent = err.message; return; }
  const body = { name: document.querySelector('#sch-name').value.trim(), cron_expr: document.querySelector('#sch-cron').value.trim(), is_active: document.querySelector('#sch-active').checked, tasks };
  const id = AutomationUI.scheduleID, button = document.querySelector('#sch-submit-btn');
  AutomationUI.saving = true; setBtnLoading(button, true, 'Saving…');
  try {
    await api(srvApiPath(srv, id ? `/schedules/${encodeURIComponent(id)}` : '/schedules'), { method: id ? 'PUT' : 'POST', body: JSON.stringify(body), suppressToast: true });
    closeModal(); toast(id ? 'Schedule updated.' : 'Schedule created.', 'ok');
    if (sameQolServer(srv, curServer())) await loadServerSchedules();
  } catch (err) { error.textContent = err.message; }
  finally { AutomationUI.saving = false; if (button?.isConnected) setBtnLoading(button, false); }
}
async function restoreBackup(name, button) {
  const srv = curServer();
  if (!srv || !name || AutomationUI.restoring) return;
  AutomationUI.restoring = { ...srv };
  try {
    const snapshot = await api(srvApiPath(srv, ''), { suppressToast: true });
    const status = String(snapshot.stats?.status || snapshot.status || '').toLowerCase();
    if (!['offline', 'stopped'].includes(status)) {
      await uiAlert({ title: 'Stop the server first', message: 'A backup can only be restored while the server is fully stopped. Stop it from the console, then retry.' }); return;
    }
    const confirmed = await uiConfirm({ title: 'Restore server backup?', message: `<p>Restore <b>${esc(name)}</b> on <b>${esc(srv.name)}</b>?</p><p>This replaces the current server files, including worlds, plugins, and configuration. Files absent from the backup will be removed. The archive is checked before replacement; the server must remain stopped.</p>`, confirmText: 'Replace files & restore', isDanger: true });
    if (!confirmed || !sameQolServer(srv, curServer())) return;
    if (button) setBtnLoading(button, true, 'Restoring…');
    const result = await api(srvApiPath(srv, '/backups/restore'), { method: 'POST', body: JSON.stringify({ name, confirm: true }), suppressToast: true });
    toast(result.message || 'Backup restored. The server remains stopped.', 'ok');
    AutomationUI.lastRestore = { server: { ...srv }, safetyCopy: result.safety_copy || '' };
    if (sameQolServer(srv, curServer()) && S.view === 'bk') await loadBackupsList();
  } catch (err) { toast(err.message || 'Backup restore failed.', 'bad'); }
  finally { AutomationUI.restoring = null; if (button?.isConnected) setBtnLoading(button, false); }
}

function renderRestoreNotice() {
  const restore = AutomationUI.lastRestore;
  if (!restore?.safetyCopy || !sameQolServer(restore.server, curServer())) return '';
  return `<div class="card np-restore-notice" role="status"><b>Backup restored — previous files retained</b><p class="mu">Safety copy, relative to the host’s servers directory: <code>${esc(restore.safetyCopy)}</code>. Verify the restored server before removing that copy manually. It continues to use disk space.</p></div>`;
}
