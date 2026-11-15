(() => {
  'use strict';
  const get = id => document.getElementById(id);
  const panel = get('panel');
  let home = '', closing = false, resizeTimer, initialToken = '', guest = false;
  const native = (name, ...args) => window.runtime?.[name]?.(...args);
  async function syncMaximized() {
    const maximized = await native('WindowIsMaximised');
    const label = maximized ? 'Restore window' : 'Maximize';
    get('maximize').setAttribute('aria-label', label);
    get('maximize').title = label;
    get('maximize-icon').setAttribute('d', maximized ? 'M8 6V4h12v12h-2M4 8h12v12H4Z' : 'M6 6h12v12H6Z');
  }
  function navigate() {
    if (!home || closing) return;
    get('loading').hidden = false;
    panel.src = home;
  }
  get('home').addEventListener('click', navigate);
  get('reload').addEventListener('click', navigate);
  get('data').addEventListener('click', () => { if (!closing) native('EventsEmit', 'desktop:data'); });
  get('minimize').addEventListener('click', () => native('WindowMinimise'));
  get('maximize').addEventListener('click', () => native('WindowToggleMaximise'));
  get('close').addEventListener('click', () => { if (!closing) native('Quit'); });
  panel.addEventListener('load', () => { if (home) get('loading').hidden = true; });
  window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(syncMaximized, 100);
  });
  // Session-only handoff. Never relay frame messages into native actions.
  window.addEventListener('message', event => {
    if (closing || !home || event.origin !== new URL(home).origin || event.source !== panel.contentWindow || event.data?.type !== 'np:desktop-ready') return;
    panel.contentWindow.postMessage({type: 'np:desktop-session', token: initialToken, guest}, new URL(home).origin);
    initialToken = ''; // A reload, preview, or logged-out page cannot reacquire it.
  });
  window.desktopShell = Object.freeze({
    init(target, token = '') {
      if (home || closing) return;
      const url = new URL(target);
      if (url.protocol !== 'http:' || url.hostname !== '127.0.0.1' || !url.port || url.username || url.password) return;
      home = url.origin + '/';
      initialToken = typeof token === 'string' ? token : '';
      guest = Boolean(initialToken);
      get('status').textContent = guest ? 'Local Guest · administrator' : 'Local panel';
      navigate();
      syncMaximized();
    },
    stopping() {
      closing = true;
      get('shutdown').hidden = false;
      get('status').textContent = 'Shutting down';
      for (const id of ['home', 'reload', 'data', 'close']) get(id).disabled = true;
    }
  });
})();
