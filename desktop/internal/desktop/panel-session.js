(() => {
  'use strict';
  if (window.parent === window || location.protocol !== 'http:' || location.hostname !== '127.0.0.1' || !location.port || !['/', '/index.html'].includes(location.pathname)) return;
  window.NPDesktopSession = new Promise(resolve => {
    let timer;
    const finish = token => {
      window.removeEventListener('message', receive);
      clearTimeout(timer);
      resolve(token);
    };
    const receive = event => {
      if (event.source !== window.parent || event.origin !== 'http://wails.localhost' || event.data?.type !== 'np:desktop-session' || typeof event.data.token !== 'string') return;
      window.NPDesktopGuest = Boolean(event.data.guest);
      finish(event.data.token);
    };
    window.addEventListener('message', receive);
    timer = setTimeout(() => finish(''), 5000);
    window.parent.postMessage({type: 'np:desktop-ready'}, 'http://wails.localhost');
  });
})();
