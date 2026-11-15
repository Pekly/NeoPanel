(() => {
  'use strict';
  const cache = new Map();
  let context = null;
  const steps = [
    { label: 'Your account', icon: 'shield', title: 'Make this account yours',
      text: 'You are the administrator. This account can manage every server and invite other people, so keep its credentials private.',
      items: ['Add two-factor authentication or a passkey in Account → Security.', 'Before inviting anyone, configure SMTP and the public panel URL in Admin. Password-reset emails need both.', 'Use separate accounts for your team instead of sharing your administrator password.'],
      action: 'Open account security', target: 'security', note: 'Come back to Your servers to continue this guide.' },
    { label: 'Choose a host', icon: 'server', title: 'Decide where Minecraft will run',
      text: 'The panel is your control room. A host (shown as a node) is the machine that actually runs a game server.',
      items: ['Choose Local to run a server on the same machine as this panel. No extra worker is needed.', 'For another machine, install a worker and connect it under Admin → Nodes. Protect worker APIs with HTTPS or a trusted private network.', 'Use configured Linux isolation for untrusted workloads. Windows local hosting is trusted-only; process limits are not a filesystem or network sandbox.'],
      action: 'Review hosts and nodes', target: 'nodes', note: 'Players connect to the game host and port, not the panel’s web address.' },
    { label: 'Create a server', icon: 'plus', title: 'Start with one server',
      text: 'The Create server form lets you choose the host, server software, Minecraft version, memory, and game port.',
      items: ['Paper is a common choice for a plugin-based Java server; Vanilla keeps the original server software. Pick the engine your world needs.', 'Choose a compatible Java runtime on the selected host. If Java is missing, use the runtime controls in the form.', 'Leave memory for the operating system and use an unused game port. Review the Minecraft EULA before running Minecraft.'],
      action: 'Create a server', target: 'create', note: 'This opens the form only. You stay in control of installation and startup.' },
    { label: 'Run and protect', icon: 'terminal', title: 'Know your way around',
      text: 'Open a server card to reach its console. That is where you start and stop it, read startup errors, and send commands.',
      items: ['Wait for the console to confirm the server is ready before connecting. The server address includes the game port.', 'Files is where you edit configuration and manage plugins. Stop the server before changes that require a restart.', 'Use Backups before upgrades, keep an independent copy, and stop the server for a consistent world snapshot. Restore replaces active files.'],
      action: 'Back to your servers', target: 'dashboard', note: 'Finishing this guide marks the tutorial as read. It does not certify your server or security configuration.' }
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
      <div class="np-guide-intro"><span class="np-guide-kicker">GETTING STARTED</span><h2>Let’s get you ready to host.</h2><p>A short tour of the essentials. Take it at your own pace.</p>
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
