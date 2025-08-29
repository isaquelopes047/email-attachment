package agendador

import (
	"context"
	"time"

	"github.com/transben/pendencias-downloader/internal/config"
	"github.com/transben/pendencias-downloader/internal/logger"
	"github.com/transben/pendencias-downloader/internal/servico"
)

type Agendador struct {
	servico *servico.ServicoBaixarAnexos
	store   *config.Store

	chTroca chan struct{}
}

func NovoAgendador(s *servico.ServicoBaixarAnexos, store *config.Store) *Agendador {
	return &Agendador{servico: s, store: store, chTroca: make(chan struct{}, 1)}
}

func (a *Agendador) notificarTroca() {
	select {
	case a.chTroca <- struct{}{}:
	default:
	}
}

func (a *Agendador) Iniciar(ctx context.Context) error {
	a.servico.ExecutarUmaVez()

	intervalo := time.Duration(a.store.Obter().IntervaloMinutos) * time.Minute
	if intervalo <= 0 {
		intervalo = 10 * time.Minute
	}
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Log("[agendador] encerrando...")
			return ctx.Err()
		case <-ticker.C:
			a.servico.ExecutarUmaVez()
		case <-a.chTroca:
			ticker.Stop()
			novo := time.Duration(a.store.Obter().IntervaloMinutos) * time.Minute
			if novo <= 0 {
				novo = 10 * time.Minute
			}
			ticker = time.NewTicker(novo)
			logger.Log("[agendador] intervalo atualizado para %v", novo)
		}
	}
}

// Expor para UI chamar quando /api/config mudar
func (a *Agendador) SinalizarMudancaIntervalo() { a.notificarTroca() }
