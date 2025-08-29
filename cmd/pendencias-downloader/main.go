package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv" // <— ADD

	"github.com/transben/pendencias-downloader/internal/agendador"
	"github.com/transben/pendencias-downloader/internal/config"
	"github.com/transben/pendencias-downloader/internal/logger"
	"github.com/transben/pendencias-downloader/internal/servico"
	"github.com/transben/pendencias-downloader/internal/ui"
)

func main() {
	// Carrega .env (se existir). Se não existir, segue normal.
	_ = godotenv.Load(".env")

	logger.Iniciar()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Store dinâmica já aplica overrides de ENV (ver passo 4)
	store, err := config.NovaStore("configs/config.yaml")
	if err != nil {
		log.Fatalf("erro ao carregar config: %v", err)
	}

	serv := servico.NovoServicoBaixarAnexos(store)
	age := agendador.NovoAgendador(serv, store)

	go func() {
		if err := age.Iniciar(ctx); err != nil {
			logger.Log("agendador finalizado com erro: %v", err)
		}
	}()

	const portaPainel = 8080
	cfgSnap := store.Obter()
	mux := ui.NovoMux(serv, portaPainel, cfgSnap.PastaDestino, cfgSnap.IntervaloMinutos, store, age)
	painel := ui.NovoServidorUI(portaPainel, mux)
	if err := painel.Iniciar(ctx); err != nil {
		logger.Log("erro ao iniciar painel: %v", err)
	}

	logger.Log("pendencias-downloader iniciado. Intervalo: %d min. Pasta destino: %s",
		cfgSnap.IntervaloMinutos, cfgSnap.PastaDestino)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	<-sigs
	logger.Log("sinal recebido, finalizando...")
	cancel()
}
