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
	"time"

	"github.com/transben/pendencias-downloader/internal/agendador"
	"github.com/transben/pendencias-downloader/internal/config"
	"github.com/transben/pendencias-downloader/internal/logger"
	"github.com/transben/pendencias-downloader/internal/servico"
)

var assetsFS embed.FS

type DadosPagina struct {
	Porta         int
	DataHoje      string
	ArquivoLogDia string
	PastaDestino  string
	Intervalo     int
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

	sub, err := fs.Sub(assetsFS, "assets")
	if err != nil {
		panic(fmt.Errorf("embed 'assets' não encontrado: %w", err))
	}

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		raw, err := fs.ReadFile(assetsFS, "assets/index.html")
		if err != nil {
			http.Error(w, "index não encontrado: "+err.Error(), http.StatusInternalServerError)
			return
		}
		tpl := template.Must(template.New("index").Parse(string(raw)))

		hoje := time.Now().Format("2006-01-02")
		cfg := store.Obter()

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.Execute(w, DadosPagina{
			Porta:         porta,
			DataHoje:      hoje,
			ArquivoLogDia: filepath.Join("logs", hoje+".log"),
			PastaDestino:  cfg.PastaDestino,
			Intervalo:     cfg.IntervaloMinutos,
		})
	})

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

	return mux
}

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
