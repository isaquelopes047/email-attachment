// internal/ui/assets/js/ui.js

export function setVal(id, v) {
  const el = document.getElementById(id);
  if (!el) return;
  if (el.type === 'checkbox') el.checked = !!v;
  else el.value = (v ?? '').toString();
}

export function preencherCampos(cfg) {
  setVal('cfg_pasta', cfg.pasta_destino);
  setVal('cfg_intervalo', cfg.intervalo_minutos);
  setVal('cfg_caixa', cfg.caixa);
  setVal('cfg_lido', cfg.marcar_como_lido);
  setVal('cfg_sonaolidos', cfg.so_nao_lidos);

  const email = cfg.email || {};
  setVal('cfg_servidor', email.servidor);
  setVal('cfg_porta', email.porta);
  setVal('cfg_usuario', email.usuario);
  setVal('cfg_tls', email.usar_tls);

  const senhaEl = document.getElementById('cfg_senha');
  if (senhaEl) senhaEl.value = ''; // nunca auto-preencher senha
}

export function coletarPayloadDoForm() {
  const payload = {
    pasta_destino: document.getElementById('cfg_pasta').value,
    intervalo_minutos: Number(document.getElementById('cfg_intervalo').value),
    caixa: document.getElementById('cfg_caixa').value,
    marcar_como_lido: document.getElementById('cfg_lido').checked,
    so_nao_lidos: document.getElementById('cfg_sonaolidos').checked,
    email: {
      servidor: document.getElementById('cfg_servidor').value,
      porta: Number(document.getElementById('cfg_porta').value),
      usuario: document.getElementById('cfg_usuario').value,
      usar_tls: document.getElementById('cfg_tls').checked,
    }
  };
  const senha = (document.getElementById('cfg_senha').value || '').trim();
  if (senha) payload.email.senha = senha; // só atualiza se usuário digitou
  return payload;
}

export function renderLogs(lines) {
  const pre = document.getElementById('logs');
  if (!pre) return;
  pre.textContent = (lines && lines.length) ? lines.join('\n') : '(sem logs para hoje)';
}

export function setStatus(id, text, timeoutMs = 2500) {
  const el = document.getElementById(id);
  if (!el) return;
  el.textContent = text;
  if (timeoutMs) setTimeout(() => (el.textContent = ''), timeoutMs);
}
