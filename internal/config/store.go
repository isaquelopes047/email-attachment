package config

import (
	"bytes"
	"errors"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Store mantém a configuração em memória com acesso concorrente seguro.
type Store struct {
	mu      sync.RWMutex
	cfg     *Config
	arqYAML string
}

// NovaStore carrega a config do YAML e a mantém em memória.
func NovaStore(caminho string) (*Store, error) {
	cfg, err := CarregarConfig(caminho)
	if err != nil {
		return nil, err
	}
	return &Store{cfg: cfg, arqYAML: caminho}, nil
}

// Obter devolve um snapshot da config atual (cópia).
func (s *Store) Obter() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return *s.cfg
}

// Atualizar substitui campos da config atual e aplica defaults.
func (s *Store) Atualizar(nova Config) {
	// defaults
	if nova.IntervaloMinutos <= 0 {
		nova.IntervaloMinutos = 10
	}
	if nova.PastaDestino == "" {
		nova.PastaDestino = "Pendencias"
	}
	if nova.Caixa == "" {
		nova.Caixa = "INBOX"
	}
	s.mu.Lock()
	s.cfg = &nova
	s.mu.Unlock()
}

// Salvar persiste a config atual no YAML com escrita atômica.
func (s *Store) Salvar() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(s.cfg); err != nil {
		return err
	}
	_ = enc.Close()

	tmp := s.arqYAML + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.arqYAML)
}

// CarregarDoDisco recarrega do YAML (opcional).
func (s *Store) CarregarDoDisco() error {
	cfg, err := CarregarConfig(s.arqYAML)
	if err != nil {
		return err
	}
	s.Atualizar(*cfg)
	return nil
}

// Sanitizada esconde campos sigilosos para exibir no painel (ex.: senha).
func (s *Store) Sanitizada() Config {
	c := s.Obter()
	if c.Email.Senha != "" {
		c.Email.Senha = "********"
	}
	return c
}

// AplicarUpdateParcial permite atualizar só alguns campos vindos do painel.
func (s *Store) AplicarUpdateParcial(p map[string]any) (Config, error) {
	// Começa da atual
	atual := s.Obter()

	// Helper p/ extrair typed (evita panics)
	getStr := func(k string) (string, bool) {
		if v, ok := p[k]; ok && v != nil {
			if s, ok2 := v.(string); ok2 {
				return s, true
			}
		}
		return "", false
	}
	getBool := func(k string) (bool, bool) {
		if v, ok := p[k]; ok && v != nil {
			if b, ok2 := v.(bool); ok2 {
				return b, true
			}
		}
		return false, false
	}
	getInt := func(k string) (int, bool) {
		if v, ok := p[k]; ok && v != nil {
			switch vv := v.(type) {
			case float64: // JSON numérico chega como float64
				return int(vv), true
			case int:
				return vv, true
			}
		}
		return 0, false
	}

	// Campos de topo
	if s, ok := getStr("pasta_destino"); ok {
		atual.PastaDestino = s
	}
	if i, ok := getInt("intervalo_minutos"); ok {
		atual.IntervaloMinutos = i
	}
	if b, ok := getBool("marcar_como_lido"); ok {
		atual.MarcarComoLido = b
	}
	if b, ok := getBool("so_nao_lidos"); ok {
		atual.SoNaoLidos = b
	}
	if s, ok := getStr("caixa"); ok {
		atual.Caixa = s
	}

	// Subobjeto email (se vier)
	if v, ok := p["email"]; ok && v != nil {
		if m, ok2 := v.(map[string]any); ok2 {
			if s, ok := m["servidor"].(string); ok {
				atual.Email.Servidor = s
			}
			if i, ok := m["porta"].(float64); ok {
				atual.Email.Porta = int(i)
			}
			if s, ok := m["usuario"].(string); ok {
				atual.Email.Usuario = s
			}
			if s, ok := m["senha"].(string); ok && s != "" && s != "********" {
				atual.Email.Senha = s
			}
			if b, ok := m["usar_tls"].(bool); ok {
				atual.Email.UsarTLS = b
			}
		}
	}

	// Validações simples
	if atual.Email.Servidor == "" || atual.Email.Usuario == "" {
		return atual, errors.New("email.servidor e email.usuario são obrigatórios")
	}
	if atual.Email.Porta == 0 {
		atual.Email.Porta = 993
	}

	// Aplica
	s.Atualizar(atual)
	return s.Obter(), nil
}
