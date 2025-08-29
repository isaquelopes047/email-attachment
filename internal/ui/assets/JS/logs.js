import { getLogs } from './api.js';
import { renderLogs } from './dom.js';

function updateDownloadLink() {
  const sel = document.getElementById('qtd');
  const n = sel ? Number(sel.value || 200) : 200;
  const a = document.getElementById('baixar');
  if (a) a.href = `/download/logs?linhas=${encodeURIComponent(n)}`;
}

async function carregar(){
  const sel = document.getElementById('qtd');
  const n = sel ? Number(sel.value || 200) : 200;
  const js = await getLogs(n);
  renderLogs(js.linhas || []);
  updateDownloadLink(); // mantém o botão coerente com a tela
}

document.addEventListener('DOMContentLoaded', ()=>{
  carregar().catch(()=>{});

  const btn = document.getElementById('atualizar');
  if (btn) btn.addEventListener('click', ()=>carregar());

  const sel = document.getElementById('qtd');
  if (sel) sel.addEventListener('change', ()=>{
    updateDownloadLink(); // atualiza href ao trocar a quantidade
    carregar().catch(()=>{}); // opcional: já recarrega as linhas
  });

  setInterval(()=>carregar().catch(()=>{}), 5000);
});
