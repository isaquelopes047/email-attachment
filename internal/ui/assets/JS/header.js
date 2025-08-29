async function carregarSaude(){
  try{
    const r = await fetch('/api/health');
    if (!r.ok) return;
    const h = await r.json();
    const el = document.getElementById('header-alert');
    if (!el) return;

    // limpa classes e torna visível
    el.classList.remove('error','success');
    el.hidden = false;

    if (!h.config_ok) {
      const faltando = Array.isArray(h.faltando) && h.faltando.length ? h.faltando.join(', ') : 'campos obrigatórios';
      el.classList.add('error');
      el.innerHTML = `Configuração incompleta: ${faltando}. <a href="/config">Abrir configurações</a>`;
      return;
    }

    if (!h.imap_ok) {
      el.classList.add('error');
      const detalhe = h.erro ? ` — ${h.erro}` : '';
      el.innerHTML = `Falha na conexão IMAP${detalhe}. <a href="/config">Ver credenciais</a>`;
      return;
    }

    // ✅ tudo ok → verde discreto e permanente
    el.classList.add('success');
    const hora = h.ts ? new Date(h.ts).toLocaleTimeString() : '';
    el.innerHTML = `IMAP conectado com sucesso! <span class="muted">${hora ? '('+hora+')' : ''}</span>`;
  } catch (_) {
    // em caso de erro de rede, não muda o estado atual
  }
}

// inicial e refresh a cada 60s
document.addEventListener('DOMContentLoaded', ()=>{
  carregarSaude();
  setInterval(carregarSaude, 60000);
});
