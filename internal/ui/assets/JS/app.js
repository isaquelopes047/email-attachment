import { getConfig, putConfig, getLogs, execNow } from './api.js';
import { preencherCampos, coletarPayloadDoForm, renderLogs, setStatus, showView } from './ui.js';

function normalizarConfig(c) {
  const email = c.email ?? c.Email ?? {};
  return {
    pasta_destino: c.pasta_destino ?? c.PastaDestino ?? '',
    intervalo_minutos: c.intervalo_minutos ?? c.IntervaloMinutos ?? 10,
    caixa: c.caixa ?? c.Caixa ?? 'INBOX',
    marcar_como_lido: c.marcar_como_lido ?? c.MarcarComoLido ?? false,
    so_nao_lidos: c.so_nao_lidos ?? c.SoNaoLidos ?? true,
    email: {
      servidor: email.servidor ?? email.Servidor ?? '',
      porta: email.porta ?? email.Porta ?? 993,
      usuario: email.usuario ?? email.Usuario ?? '',
      usar_tls: email.usar_tls ?? email.UsarTLS ?? true
    }
  };
}

async function carregarConfig() {
  const cfgRaw = await getConfig();
  const cfg = normalizarConfig(cfgRaw);
  preencherCampos(cfg);
  // arquivo de log do dia (mostra na tela de status)
  const dataHoje = document.getElementById('status_data')?.textContent || '';
  const arq = `logs/${dataHoje}.log`;
  const el = document.getElementById('status_arquivo_log');
  if (el) el.textContent = arq;
}

async function carregarLogs() {
  const sel = document.getElementById('qtd');
  const n = sel ? Number(sel.value || 200) : 200;
  const js = await getLogs(n);
  renderLogs(js.linhas || []);
}

async function salvarConfig() {
  setStatus('cfgStatus', 'salvando...');
  const payload = coletarPayloadDoForm();
  try {
    await putConfig(payload);
    setStatus('cfgStatus', 'salvo!');
    await carregarConfig();
  } catch (e) {
    setStatus('cfgStatus', 'erro ao salvar');
  }
}

async function executarAgora() {
  setStatus('statusExecucao', 'enviando...');
  try {
    await execNow();
    setStatus('statusExecucao', 'execução agendada — acompanhe os logs');
    setTimeout(carregarLogs, 1200);
  } catch (e) {
    setStatus('statusExecucao', 'erro de rede');
  }
}

function wire() {
  const map = {
    'btnSalvarCfg': salvarConfig,
    'executar': executarAgora,
    'atualizar': carregarLogs
  };
  Object.entries(map).forEach(([id, fn]) => {
    const el = document.getElementById(id);
    if (el) el.addEventListener('click', fn);
  });

  // auto-refresh logs
  setInterval(() => {
    const h = location.hash || '#/status';
    if (h.startsWith('#/logs')) carregarLogs().catch(() => { });
  }, 5000);

  window.addEventListener('hashchange', route);
}

export function showView(name) {
  const views = ['status', 'logs', 'config'];
  views.forEach(v => {
    const el = document.getElementById(`view-${v}`);
    if (!el) return;
    const ativo = (v === name);
    if (ativo) {
      el.removeAttribute('hidden'); // padrão HTML
      el.style.display = '';        // reseta qualquer inline antigo
    } else {
      el.setAttribute('hidden', ''); // padrão HTML
      el.style.display = 'none';     // força caso algum CSS conflite
    }
  });

  // realça a aba ativa
  document.querySelectorAll('.tab').forEach(a => {
    a.classList.toggle('active', a.dataset.route === name);
  });
}

document.addEventListener('DOMContentLoaded', async () => {
  wire();
  if (!location.hash) location.hash = '#/status';
  await route();
});
