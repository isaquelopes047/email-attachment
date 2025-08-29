package config

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config representa o arquivo de configuração principal.
type Config struct {
	PastaDestino     string   `yaml:"pasta_destino" json:"pasta_destino"`
	IntervaloMinutos int      `yaml:"intervalo_minutos" json:"intervalo_minutos"`
	Email            EmailCfg `yaml:"email"            json:"email"`
	MarcarComoLido   bool     `yaml:"marcar_como_lido" json:"marcar_como_lido"`
	SoNaoLidos       bool     `yaml:"so_nao_lidos"     json:"so_nao_lidos"`
	Caixa            string   `yaml:"caixa"            json:"caixa"`
}

// EmailCfg agrega as credenciais e parâmetros de conexão IMAP.
type EmailCfg struct {
	Servidor string `yaml:"servidor"  json:"servidor"`
	Porta    int    `yaml:"porta"     json:"porta"`
	Usuario  string `yaml:"usuario"   json:"usuario"`
	Senha    string `yaml:"senha"     json:"senha"`
	UsarTLS  bool   `yaml:"usar_tls"  json:"usar_tls"`
}

// CarregarConfig lê e faz o parse do arquivo YAML de configuração e aplica overrides de ENV.
func CarregarConfig(caminho string) (*Config, error) {
	f, err := os.Open(caminho)
	if err != nil {
		// Se não existir YAML, começamos com defaults e só ENV
		cfg := &Config{}
		aplicarDefaults(cfg)
		aplicarEnv(cfg)
		return cfg, nil
	}
	defer f.Close()

	var cfg Config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	aplicarDefaults(&cfg)
	aplicarEnv(&cfg)
	return &cfg, nil
}

// --------- helpers ---------

func aplicarDefaults(cfg *Config) {
	if cfg.IntervaloMinutos <= 0 {
		cfg.IntervaloMinutos = 10
	}
	if cfg.PastaDestino == "" {
		cfg.PastaDestino = "Pendencias"
	}
	if cfg.Caixa == "" {
		cfg.Caixa = "INBOX"
	}
	if cfg.Email.Porta == 0 {
		cfg.Email.Porta = 993
	}
}

// aplicarEnv sobrescreve valores da config com variáveis de ambiente, se definidas.
func aplicarEnv(cfg *Config) {
	// strings
	if v := getenvTrim("PASTA_DESTINO"); v != "" {
		cfg.PastaDestino = v
	}
	if v := getenvTrim("CAIXA"); v != "" {
		cfg.Caixa = v
	}
	if v := getenvTrim("EMAIL_SERVIDOR"); v != "" {
		cfg.Email.Servidor = v
	}
	if v := getenvTrim("EMAIL_USUARIO"); v != "" {
		cfg.Email.Usuario = v
	}
	if v := getenvRaw("EMAIL_SENHA"); v != "" { // raw para manter espaços
		cfg.Email.Senha = v
	}

	// ints
	if n, ok := getenvInt("INTERVALO_MINUTOS"); ok {
		cfg.IntervaloMinutos = n
	}
	if n, ok := getenvInt("EMAIL_PORTA"); ok {
		cfg.Email.Porta = n
	}

	// bools
	if b, ok := getenvBool("MARCAR_COMO_LIDO"); ok {
		cfg.MarcarComoLido = b
	}
	if b, ok := getenvBool("SO_NAO_LIDOS"); ok {
		cfg.SoNaoLidos = b
	}
	if b, ok := getenvBool("EMAIL_USAR_TLS"); ok {
		cfg.Email.UsarTLS = b
	}
}

func getenvTrim(k string) string {
	return strings.TrimSpace(os.Getenv(k))
}
func getenvRaw(k string) string { // não trim para manter espaços no meio
	v := os.Getenv(k)
	return v
}
func getenvInt(k string) (int, bool) {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	return n, err == nil
}
func getenvBool(k string) (bool, bool) {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return false, false
	}
	// aceita: true/false/1/0/TRUE/FALSE etc.
	b, err := strconv.ParseBool(v)
	return b, err == nil
}
