export function setVal(id, v) {
  const el = document.getElementById(id);
  if (!el) return;
  if (el.type === 'checkbox') el.checked = !!v;
  else el.value = (v ?? '').toString();
}
export function setText(id, txt) {
  const el = document.getElementById(id);
  if (el) el.textContent = txt ?? '—';
}
export function renderLogs(lines) {
  setText('logs', (lines && lines.length) ? lines.join('\n') : '(sem logs para hoje)');
}
export function setStatus(id, text, timeoutMs = 2500) {
  const el = document.getElementById(id);
  if (!el) return;
  el.textContent = text;
  if (timeoutMs) setTimeout(()=> (el.textContent=''), timeoutMs);
}
export function valueOf(id){ const el=document.getElementById(id); return el?el.value:'' }
export function checked(id){ const el=document.getElementById(id); return el?!!el.checked:false }