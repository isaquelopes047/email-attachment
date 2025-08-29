import { getConfig, execNow } from './api.js';
import { setText, setStatus } from './dom.js';

function normalizar(c){
  const e = c.email ?? c.Email ?? {};
  return {
    pasta_destino: c.pasta_destino ?? c.PastaDestino ?? '',
    intervalo_minutos: c.intervalo_minutos ?? c.IntervaloMinutos ?? 10,
    caixa: c.caixa ?? c.Caixa ?? 'INBOX',
    email: { servidor: e.servidor ?? e.Servidor ?? '' }
  };
}

async function carregar(){
  const cfg = normalizar(await getConfig());
  setText('status_pasta', cfg.pasta_destino);
  setText('status_intervalo', `${cfg.intervalo_minutos} min`);
  setText('status_caixa', cfg.caixa);
  setText('status_servidor', cfg.email.servidor);
}

async function executar(){
  setStatus('statusExecucao','enviando...');
  try{
    await execNow();
    setStatus('statusExecucao','execução agendada — confira em Logs');
  }catch{
    setStatus('statusExecucao','erro ao enviar');
  }
}

document.addEventListener('DOMContentLoaded', ()=>{
  carregar().catch(()=>{});
  const b = document.getElementById('executar');
  if (b) b.addEventListener('click', executar);
});
