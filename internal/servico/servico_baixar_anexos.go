package servico

import (
	"io"
	"log"

	"github.com/transben/pendencias-downloader/internal/config"
	"github.com/transben/pendencias-downloader/internal/email"
	"github.com/transben/pendencias-downloader/internal/repositorio"
)

// ServicoBaixarAnexos orquestra a rotina de conexão e salvamento de anexos.
type ServicoBaixarAnexos struct {
	store *config.Store
}

// NovoServicoBaixarAnexos constrói o serviço com a config injetada.
func NovoServicoBaixarAnexos(store *config.Store) *ServicoBaixarAnexos {
	return &ServicoBaixarAnexos{store: store}
}

// ExecutarUmaVez realiza um ciclo: conecta no IMAP, busca mensagens alvo, baixa anexos e salva no disco.
func (s *ServicoBaixarAnexos) ExecutarUmaVez() {
	cfg := s.store.Obter()

	cl := email.NovoClienteIMAP(&cfg)

	// 1) Conectar
	if err := cl.Conectar(); err != nil {
		log.Printf("[servico] erro ao conectar IMAP: %v", err)
		return
	}
	defer cl.Desconectar()

	// 2) Selecionar caixa
	if _, err := cl.SelecionarCaixa(cfg.Caixa); err != nil {
		log.Printf("[servico] erro ao selecionar caixa: %v", err)
		return
	}

	// 3) Buscar mensagens (não lidas, se configurado)
	uids, err := cl.BuscarMensagensAlvo(cfg.SoNaoLidos)
	if err != nil {
		log.Printf("[servico] erro ao buscar mensagens: %v", err)
		return
	}
	if len(uids) == 0 {
		log.Printf("[servico] nenhuma mensagem para processar")
		return
	}

	log.Printf("[servico] processando %d mensagem(ns)...", len(uids))

	// 4) Baixar anexos e salvar
	onAnexo := func(nome string, r io.Reader) error {
		caminho, err := repositorio.SalvarAnexo(cfg.PastaDestino, nome, r)
		if err != nil {
			return err
		}
		log.Printf("[servico] anexo salvo em: %s", caminho)
		return nil
	}

	if err := cl.BaixarAnexosDeMensagens(uids, onAnexo, cfg.MarcarComoLido); err != nil {
		log.Printf("[servico] erro ao baixar anexos: %v", err)
		return
	}

	log.Printf("[servico] execução concluída")
}
