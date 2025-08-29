// internal/ui/assets/js/app.js
import { getConfig, putConfig, getLogs, execNow } from './api.js';
import { preencherCampos, coletarPayloadDoForm, renderLogs, setStatus } from './ui.js';

async function carregarConfig() {
  const cfg = await getConfig();
  preencherCampos(cfg);
}

async function carregarLogs() {
  const select = document.getElementById('qtd');
  const n = select ? Number(select.value || 200) : 200;
  const js = await getLogs(n);
  renderLogs(js.linhas || []);
}

async function salvarConfig() {
  setStatus('cfgStatus', 'salvando...');
  const payload = coletarPayloadDoForm();
  try {
    await putConfig(payload);
    setStatus('cfgStatus', 'salvo!');
    await carregarConfig(); // reflete o que persistiu
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

function wireEvents() {
  const btnSalvar = document.getElementById('btnSalvarCfg');
  if (btnSalvar) btnSalvar.addEventListener('click', salvarConfig);

  const btnExec = document.getElementById('executar');
  if (btnExec) btnExec.addEventListener('click', executarAgora);

  const btnAtualizar = document.getElementById('atualizar');
  if (btnAtualizar) btnAtualizar.addEventListener('click', carregarLogs);

  // auto-refresh de logs
  setInterval(carregarLogs, 5000);
}

document.addEventListener('DOMContentLoaded', async () => {
  wireEvents();
  await carregarConfig();
  await carregarLogs();
});
