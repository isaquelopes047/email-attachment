import { getConfig, putConfig } from './api.js';
import { setVal, setStatus, valueOf, checked } from './dom.js';

function normalizar(c){
  const e = c.email ?? c.Email ?? {};
  return {
    pasta_destino: c.pasta_destino ?? c.PastaDestino ?? '',
    intervalo_minutos: c.intervalo_minutos ?? c.IntervaloMinutos ?? 10,
    caixa: c.caixa ?? c.Caixa ?? 'INBOX',
    marcar_como_lido: c.marcar_como_lido ?? c.MarcarComoLido ?? false,
    so_nao_lidos: c.so_nao_lidos ?? c.SoNaoLidos ?? true,
    email: {
      servidor: e.servidor ?? e.Servidor ?? '',
      porta: e.porta ?? e.Porta ?? 993,
      usuario: e.usuario ?? e.Usuario ?? '',
      usar_tls: e.usar_tls ?? e.UsarTLS ?? true
    }
  };
}

function preencher(cfg){
  setVal('cfg_pasta', cfg.pasta_destino);
  setVal('cfg_intervalo', cfg.intervalo_minutos);
  setVal('cfg_caixa', cfg.caixa);
  document.getElementById('cfg_lido').checked = !!cfg.marcar_como_lido;
  document.getElementById('cfg_sonaolidos').checked = !!cfg.so_nao_lidos;

  setVal('cfg_servidor', cfg.email.servidor);
  setVal('cfg_porta', cfg.email.porta);
  setVal('cfg_usuario', cfg.email.usuario);
  document.getElementById('cfg_tls').checked = !!cfg.email.usar_tls;

  const s = document.getElementById('cfg_senha'); if (s) s.value = '';
}

function payload(){
  const p = {
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
  if (senha) p.email.senha = senha;
  return p;
}

async function carregar(){ preencher(normalizar(await getConfig())); }
async function salvar(){
  setStatus('cfgStatus','salvando...');
  try{
    await putConfig(payload());
    setStatus('cfgStatus','salvo!');
    await carregar();
  }catch{
    setStatus('cfgStatus','erro ao salvar');
  }
}

document.addEventListener('DOMContentLoaded', ()=>{
  carregar().catch(()=>{});
  const b = document.getElementById('btnSalvarCfg');
  if (b) b.addEventListener('click', salvar);
});
