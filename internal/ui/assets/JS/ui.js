export function setVal(id, v) {
  const el = document.getElementById(id);
  if (!el) return;
  if (el.type === 'checkbox') el.checked = !!v;
  else el.value = (v ?? '').toString();
}

export function preencherCampos(cfg) {
  // Geral
  setVal('cfg_pasta', cfg.pasta_destino);
  setVal('cfg_intervalo', cfg.intervalo_minutos);
  setVal('cfg_caixa', cfg.caixa);
  setVal('cfg_lido', cfg.marcar_como_lido);
  setVal('cfg_sonaolidos', cfg.so_nao_lidos);
  // IMAP
  const email = cfg.email || {};
  setVal('cfg_servidor', email.servidor);
  setVal('cfg_porta', email.porta);
  setVal('cfg_usuario', email.usuario);
  setVal('cfg_tls', email.usar_tls);
  const senhaEl = document.getElementById('cfg_senha'); if (senhaEl) senhaEl.value = '';

  // Status view
  setText('status_pasta', cfg.pasta_destino);
  setText('status_intervalo', (cfg.intervalo_minutos ?? '') + ' min');
  setText('status_caixa', cfg.caixa);
  setText('status_servidor', email.servidor);
}

export function coletarPayloadDoForm() {
  const payload = {
    pasta_destino: valueOf('cfg_pasta'),
    intervalo_minutos: Number(valueOf('cfg_intervalo') || 0),
    caixa: valueOf('cfg_caixa'),
    marcar_como_lido: checked('cfg_lido'),
    so_nao_lidos: checked('cfg_sonaolidos'),
    email: {
      servidor: valueOf('cfg_servidor'),
      porta: Number(valueOf('cfg_porta') || 0),
      usuario: valueOf('cfg_usuario'),
      usar_tls: checked('cfg_tls'),
    }
  };
  const senha = (valueOf('cfg_senha') || '').trim();
  if (senha) payload.email.senha = senha;
  return payload;
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

export function showView(name){
  const views = ['status','logs','config'];
  views.forEach(v => {
    const el = document.getElementById(`view-${v}`);
    if (el) el.hidden = (v !== name);
  });
  document.querySelectorAll('.tab').forEach(a=>{
    a.classList.toggle('active', a.dataset.route===name);
  });
}

function valueOf(id){ const el=document.getElementById(id); return el?el.value:'' }
function checked(id){ const el=document.getElementById(id); return el?!!el.checked:false }
function setText(id, txt){ const el=document.getElementById(id); if(el) el.textContent = txt ?? '—' }
