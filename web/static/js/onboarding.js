(() => {
  'use strict';
  const cache = new Map();
  let context = null;
  const steps = [
    { label: 'Your account', icon: 'shield', title: 'Make this account yours',
      text: 'Your account manages your servers and team.',
      items: ['Add two-factor authentication or a passkey in Account → Security.', 'Set up email in Admin before inviting others or using password resets.', 'Give each teammate their own account.'],
      action: 'Open account security', target: 'security', note: 'Return to Your servers to continue.' },
    { label: 'Choose a host', icon: 'server', title: 'Decide where Minecraft will run',
      text: 'Choose the computer that will run your server.',
      items: ['Choose Local to use this computer.', 'To use another computer, connect a worker in Admin → Nodes over HTTPS or a trusted private network.', 'Only run trusted servers on Windows. Use isolated Linux workers for untrusted workloads.'],
      action: 'Review hosts and nodes', target: 'nodes', note: 'Players connect to the game server’s address, not the panel.' },
    { label: 'Create a server', icon: 'plus', title: 'Start with one server',
      text: 'Choose your server software, Minecraft version, and memory.',
      items: ['Choose Paper for plugins or Vanilla for the original game.', 'Select a compatible Java version; install it here if needed.', 'Leave some memory free for your computer and review the Minecraft EULA.'],
      action: 'Create a server', target: 'create', note: 'Nothing starts until you choose to start it.' },
    { label: 'Run and protect', icon: 'terminal', title: 'Know your way around',
      text: 'Open a server to start it and see its console.',
      items: ['Wait until the console says the server is ready, then join.', 'Use Files to change settings or manage plugins.', 'Back up before upgrades. Stop the server for a consistent backup; restoring replaces current files.'],
      action: 'Back to your servers', target: 'dashboard', note: 'You can reopen this guide anytime.' }
  ];
  const eligible = state => Boolean(state?.token && state.user?.id && state.user.role === 'admin');
  const keyFor = user => `neopanel_getting_started_v1_${user.id}`;
  function read(user) {
    const key = keyFor(user);
    if (cache.has(key)) return cache.get(key);
    let saved;
    try { saved = JSON.parse(localStorage.getItem(key) || 'null'); } catch (_) {}
    const value = { step: Number.isInteger(saved?.step) ? Math.max(0, Math.min(3, saved.step)) : 0,
      status: ['active', 'paused', 'done'].includes(saved?.status) ? saved.status : 'paused' };
    cache.set(key, value);
    return value;
  }
  function save(user, value) {
    cache.set(keyFor(user), value);
    try { localStorage.setItem(keyFor(user), JSON.stringify(value)); } catch (_) { /* The in-memory guide still works. */ }
  }
  const icon = name => typeof window.ic === 'function' ? window.ic(name, 20) : '';
  function content(state) {
    const saved = read(state.user);
    if (saved.status !== 'active') return '';
    const step = steps[saved.step];
    return `<section class="np-guide" aria-labelledby="np-guide-title">
      <div class="np-guide-intro"><span class="np-guide-kicker">GETTING STARTED</span><h2>Set up your first server.</h2><p>Four quick steps. Skip or come back anytime.</p>
        <ol class="np-guide-steps">${steps.map((s, i) => `<li><button type="button" data-guide-step="${i}" ${i === saved.step ? 'aria-current="step"' : ''}><span>${i + 1}</span>${s.label}</button></li>`).join('')}</ol>
        <button type="button" class="btn sm" data-guide-action="pause">Not now</button>
      </div>
      <div class="np-guide-detail"><div class="np-guide-meta"><span class="np-guide-icon" aria-hidden="true">${icon(step.icon)}</span><span>Guide · ${saved.step + 1} of ${steps.length}</span></div>
        <h3 id="np-guide-title" tabindex="-1">${step.title}</h3><p>${step.text}</p>
        <ul>${step.items.map(item => `<li>${item}</li>`).join('')}</ul>
        <button type="button" class="btn pr" data-guide-action="${step.target}">${step.action}</button>
        <p class="np-guide-note">${step.note}</p>
        <div class="np-guide-bottom"><span>Saved for this account in this browser</span><div><button type="button" class="btn sm" data-guide-action="back" ${saved.step === 0 ? 'disabled' : ''}>Back</button><button type="button" class="btn sm" data-guide-action="next">${saved.step === 3 ? 'Finish guide' : 'Next'}</button></div></div>
      </div>
    </section>`;
  }
  function update(focus = true) {
    const root = document.getElementById('np-onboarding-root');
    if (!root || !eligible(context)) return;
    root.innerHTML = content(context);
    if (focus) {
      const heading = document.getElementById('np-guide-title');
      (heading || document.getElementById('np-guide-open'))?.focus();
    }
  }
  function change(step) {
    if (!eligible(context) || !Number.isInteger(step)) return;
    save(context.user, { step: Math.max(0, Math.min(3, step)), status: 'active' });
    update();
  }
  document.addEventListener('click', event => {
    const button = event.target.closest?.('[data-guide-action], [data-guide-step]');
    if (!button || !eligible(context) || !button.closest('#np-onboarding-root')) return;
    const saved = read(context.user);
    if (button.hasAttribute('data-guide-step')) { change(Number(button.dataset.guideStep)); return; }
    switch (button.dataset.guideAction) {
      case 'next':
        if (saved.step < 3) { change(saved.step + 1); return; }
        save(context.user, {step: 3, status: 'done'}); update(); break;
      case 'back': change(saved.step - 1); break;
      case 'pause': save(context.user, {...saved, status: 'paused'}); update(); break;
      case 'security': window.go('acc', {acctTab: 1}); break;
      case 'nodes': window.go('admin', {adminTab: 1}); break;
      case 'create': window.openCreateServerModal(); break;
      case 'dashboard': document.getElementById('srv-card-grid')?.scrollIntoView({block: 'start'}); break;
    }
  });
  window.NPOnboarding = Object.freeze({
    begin(user) { if (user?.id && user.role === 'admin') save(user, {step: 0, status: 'active'}); },
    render(state) { context = state; return eligible(state) ? `<div id="np-onboarding-root">${content(state)}</div>` : ''; },
    open() {
      if (!eligible(context)) return;
      const saved = read(context.user);
      change(saved.status === 'done' ? 0 : saved.step);
      document.getElementById('np-onboarding-root')?.scrollIntoView({block: 'start'});
    },
    resetSession() { context = null; cache.clear(); }
  });
})();
