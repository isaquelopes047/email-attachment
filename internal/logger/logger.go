package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	arquivoAtual *os.File
	dataAtual    string
)

// Iniciar cria (ou abre em append) o arquivo de log do dia.
func Iniciar() {
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Fatalf("erro criando pasta logs: %v", err)
	}

	rotacionarSeNecessario()

	log.SetOutput(arquivoAtual)
	log.SetFlags(log.LstdFlags)
}

// rotacionarSeNecessario verifica se mudou o dia e abre um novo arquivo
func rotacionarSeNecessario() {
	hoje := time.Now().Format("2006-01-02")

	if arquivoAtual != nil && hoje == dataAtual {
		return // ainda é o mesmo dia
	}

	if arquivoAtual != nil {
		arquivoAtual.Close()
	}

	caminho := filepath.Join("logs", fmt.Sprintf("%s.log", hoje))
	f, err := os.OpenFile(caminho, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("erro abrindo arquivo de log: %v", err)
	}

	arquivoAtual = f
	dataAtual = hoje
	log.SetOutput(arquivoAtual)
}

// Log escreve uma linha no log e garante rotação diária.
func Log(format string, v ...any) {
	rotacionarSeNecessario()
	log.Printf(format, v...)
}
