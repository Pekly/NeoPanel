(() => {
  'use strict';
  const pending = new Set();
  async function rename(id) {
    if (S.user?.role !== 'admin' || !S.token || pending.has(id)) return;
    const node = (S.nodes || []).find(n => n.id === id);
    if (!node) return;
    const token = S.token, user = S.user.id;
    const live = () => S.token === token && S.user?.id === user && S.user?.role === 'admin';
    pending.add(id);
    try {
      const input = await uiPrompt({title:'Rename node', message:'Change the display name only. Server addresses, connection settings, and running servers stay unchanged.', defaultValue:displayNodeName(id,node.name), confirmText:'Save name'});
      if (!live() || input == null) return;
      const name = input.trim();
      if (!name || [...name].length > 100 || /[\p{Cc}]/u.test(name)) { toast('Enter a name of 1–100 characters without control characters.', 'bad'); return; }
      if (name === node.name) return;
      const result = await api(`/api/nodes/${encodeURIComponent(id)}/name`, {method:'PATCH', body:JSON.stringify({name})});
      if (!live()) return;
      if (result?.id !== id || result?.name !== name) throw Error('Unexpected rename response. Refresh nodes to check the saved name.');
      for (const item of S.nodes || []) if (item.id === id) item.name = name;
      for (const server of S.servers || []) if ((server.node_id || 'local') === id) server.node_name = name;
      toast('Node name saved', 'ok'); R();
    } catch (error) { if (live()) toast(error.message || 'Could not rename node', 'bad'); }
    finally { pending.delete(id); }
  }
  window.NPNodeNames = {rename};
})();
