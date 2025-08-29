package ui

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/transben/pendencias-downloader/internal/agendador"
	"github.com/transben/pendencias-downloader/internal/config"
	"github.com/transben/pendencias-downloader/internal/email"
	"github.com/transben/pendencias-downloader/internal/logger"
	"github.com/transben/pendencias-downloader/internal/servico"
)

//go:embed assets/templates/*.html assets/templates/*/*.html assets/css/*.css assets/js/*.js assets/img/*.svg
var assetsFS embed.FS

type DadosPagina struct {
	Porta         int
	DataHoje      string
	ArquivoLogDia string
	PastaDestino  string
	Intervalo     int
	Rota          string
	LogoURL       string
}

// valida se config está completa o suficiente p/ tentar conectar
func validarConfigMinima(cfg config.Config) (ok bool, faltando []string) {
	if cfg.PastaDestino == "" {
		faltando = append(faltando, "pasta_destino")
	}
	if cfg.IntervaloMinutos <= 0 {
		faltando = append(faltando, "intervalo_minutos")
	}
	if cfg.Email.Servidor == "" {
		faltando = append(faltando, "email.servidor")
	}
	if cfg.Email.Porta <= 0 {
		faltando = append(faltando, "email.porta")
	}
	if cfg.Email.Usuario == "" {
		faltando = append(faltando, "email.usuario")
	}
	// senha é opcional aqui se você usar outro método, mas p/ Gmail normalmente precisa:
	if cfg.Email.Senha == "" {
		faltando = append(faltando, "email.senha")
	}
	return len(faltando) == 0, faltando
}

// NovoMux cria um http.Handler com todas as rotas do painel.
func NovoMux(
	srv *servico.ServicoBaixarAnexos,
	porta int,
	_ string,
	_ int,
	store *config.Store,
	age *agendador.Agendador,
) http.Handler {
	mux := http.NewServeMux()

	// /static → serve css/js do embed
	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic(fmt.Errorf("embed 'assets' não encontrado: %w", err))
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	// helper para renderizar com base+partial+page
	render := func(w http.ResponseWriter, page string, data DadosPagina) {
		tpl, err := template.ParseFS(
			assetsFS,
			"assets/templates/layout/base.html",
			"assets/templates/partials/header.html",
			"assets/templates/"+page,
		)
		if err != nil {
			http.Error(w, "template inválido: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.ExecuteTemplate(w, "base", data)
	}

	// -------- PÁGINAS --------
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hoje := time.Now().Format("2006-01-02")
		cfg := store.Obter()
		render(w, "status.html", DadosPagina{
			Porta:         porta,
			DataHoje:      hoje,
			ArquivoLogDia: filepath.Join("logs", hoje+".log"),
			PastaDestino:  cfg.PastaDestino,
			Intervalo:     cfg.IntervaloMinutos,
			Rota:          "status",
			LogoURL:       fmt.Sprintf("https://picsum.photos/seed/%d/48/48", time.Now().UnixNano()%100000),
		})
	})

	// /logs (e /log como alias)
	logsHandler := func(w http.ResponseWriter, r *http.Request) {
		hoje := time.Now().Format("2006-01-02")
		render(w, "logs.html", DadosPagina{
			Porta:         porta,
			DataHoje:      hoje,
			ArquivoLogDia: filepath.Join("logs", hoje+".log"),
			Rota:          "logs",
			LogoURL:       fmt.Sprintf("https://picsum.photos/seed/%d/48/48", time.Now().UnixNano()%100000),
		})
	}
	mux.HandleFunc("/logs", logsHandler)
	mux.HandleFunc("/log", logsHandler)

	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		hoje := time.Now().Format("2006-01-02")
		cfg := store.Obter()
		render(w, "config.html", DadosPagina{
			Porta:         porta,
			DataHoje:      hoje,
			ArquivoLogDia: filepath.Join("logs", hoje+".log"),
			PastaDestino:  cfg.PastaDestino,
			Intervalo:     cfg.IntervaloMinutos,
			Rota:          "config",
			LogoURL:       fmt.Sprintf("https://picsum.photos/seed/%d/48/48", time.Now().UnixNano()%100000),
		})
	})

	// -------- APIs --------
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		n := 200
		if v := r.URL.Query().Get("linhas"); v != "" {
			if vv, err := strconvAtoiSafe(v); err == nil && vv > 0 && vv <= 2000 {
				n = vv
			}
		}
		hoje := time.Now().Format("2006-01-02")
		caminho := filepath.Join("logs", hoje+".log")
		linhas, _ := lerUltimasLinhas(caminho, n)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data":    hoje,
			"arquivo": caminho,
			"linhas":  linhas,
		})
	})

	mux.HandleFunc("/download/logs", func(w http.ResponseWriter, r *http.Request) {
		n := 200
		if v := r.URL.Query().Get("linhas"); v != "" {
			if vv, err := strconvAtoiSafe(v); err == nil && vv > 0 && vv <= 2000 {
				n = vv
			}
		}

		hoje := time.Now().Format("2006-01-02")
		caminho := filepath.Join("logs", hoje+".log")
		linhas, err := lerUltimasLinhas(caminho, n)
		if err != nil {
			http.Error(w, "sem logs para hoje", http.StatusNotFound)
			return
		}

		nome := fmt.Sprintf("logs-%s-%dlinhas.txt", hoje, n)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, nome))
		_, _ = w.Write([]byte(strings.Join(linhas, "\n")))
	})

	mux.HandleFunc("/api/executar", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
			return
		}
		go func() {
			logger.Log("[ui] execução manual solicitada pelo usuário")
			srv.ExecutarUmaVez()
		}()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"agendado"}`))
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodPut {
			http.Error(w, "método não permitido", http.StatusMethodNotAllowed)
			return
		}
		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(store.Sanitizada())
		case http.MethodPut:
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "json inválido", http.StatusBadRequest)
				return
			}
			_, err := store.AplicarUpdateParcial(payload)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := store.Salvar(); err != nil {
				http.Error(w, "falha ao salvar no YAML", http.StatusInternalServerError)
				return
			}
			if m, ok := payload["intervalo_minutos"]; ok && m != nil {
				age.SinalizarMudancaIntervalo()
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "config": store.Sanitizada()})
		}
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		cfg := store.Obter()
		cfgOK, faltando := validarConfigMinima(cfg)

		imapOK := false
		var errStr string
		if cfgOK {
			// ping leve: conecta e desconecta
			cl := email.NovoClienteIMAP(&cfg)
			if err := cl.Conectar(); err == nil {
				imapOK = true
				cl.Desconectar()
			} else {
				errStr = err.Error()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"config_ok": cfgOK,
			"faltando":  faltando,
			"imap_ok":   imapOK,
			"erro":      errStr,
			"ts":        time.Now().Format(time.RFC3339),
		})
	})

	return mux
}

// ---------- utilitários ----------
func strconvAtoiSafe(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

func lerUltimasLinhas(caminho string, n int) ([]string, error) {
	f, err := os.Open(caminho)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	const maxRead = 2 << 20
	info, err := f.Stat()
	if err != nil {
		return []string{}, err
	}
	var start int64 = 0
	if info.Size() > maxRead {
		start = info.Size() - maxRead
	}
	if start > 0 {
		_, _ = f.Seek(start, io.SeekStart)
	}

	data, _ := io.ReadAll(f)
	linhas := splitLines(string(data))
	if len(linhas) > n {
		linhas = linhas[len(linhas)-n:]
	}
	return linhas, nil
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := range s {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start <= len(s)-1 {
		out = append(out, s[start:])
	}
	return out
}
