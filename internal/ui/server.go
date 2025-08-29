package ui

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"
)

// ServidorUI gerencia o servidor HTTP do painel.
type ServidorUI struct {
	httpSrv *http.Server
	porta   int
}

// NovoServidorUI cria a instância do servidor de UI na porta indicada.
func NovoServidorUI(porta int, mux http.Handler) *ServidorUI {
	addr := fmt.Sprintf("127.0.0.1:%d", porta)
	return &ServidorUI{
		httpSrv: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
			IdleTimeout:       60 * time.Second,
		},
		porta: porta,
	}
}

// Iniciar inicia o servidor HTTP em goroutine e faz shutdown gracioso quando o contexto for cancelado.
func (s *ServidorUI) Iniciar(ctx context.Context) error {
	ln, err := net.Listen("tcp4", s.httpSrv.Addr)
	if err != nil {
		return fmt.Errorf("falha ao abrir porta do painel (%s): %w", s.httpSrv.Addr, err)
	}

	go func() {
		log.Printf("[ui] painel disponível em http://%s", s.httpSrv.Addr)
		if err := s.httpSrv.Serve(ln); err != nil && err != http.ErrServerClosed {
			log.Printf("[ui] erro no servidor HTTP: %v", err)
		}
	}()

	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.httpSrv.Shutdown(shCtx)
	}()

	return nil
}
