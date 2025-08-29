package repositorio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// GarantirPasta cria a pasta, caso não exista.
func GarantirPasta(pasta string) error {
	return os.MkdirAll(pasta, 0o755)
}

// SalvarAnexo salva um anexo no disco, tratando colisões de nome.
func SalvarAnexo(pastaDestino, nome string, r io.Reader) (string, error) {
	if err := GarantirPasta(pastaDestino); err != nil {
		return "", fmt.Errorf("erro criando pasta destino: %w", err)
	}

	caminho := filepath.Join(pastaDestino, nome)

	// Se já existe, versiona com timestamp
	if _, err := os.Stat(caminho); err == nil {
		ext := filepath.Ext(nome)
		base := nome[:len(nome)-len(ext)]
		nome = fmt.Sprintf("%s_%d%s", base, time.Now().Unix(), ext)
		caminho = filepath.Join(pastaDestino, nome)
	}

	f, err := os.Create(caminho)
	if err != nil {
		return "", fmt.Errorf("erro criando arquivo: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, r); err != nil {
		return "", fmt.Errorf("erro escrevendo arquivo: %w", err)
	}

	return caminho, nil
}
