// internal/ui/assets/js/api.js

export async function getConfig() {
  const r = await fetch('/api/config');
  if (!r.ok) throw new Error('GET /api/config falhou');
  return r.json();
}

export async function putConfig(payload) {
  const r = await fetch('/api/config', {
    method: 'PUT',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(payload),
  });
  if (!r.ok) throw new Error('PUT /api/config falhou');
  return r.json();
}

export async function getLogs(n = 200) {
  const r = await fetch(`/api/logs?linhas=${encodeURIComponent(n)}`);
  if (!r.ok) throw new Error('GET /api/logs falhou');
  return r.json();
}

export async function execNow() {
  const r = await fetch('/api/executar', { method: 'POST' });
  if (!r.ok) throw new Error('POST /api/executar falhou');
  return r.json();
}