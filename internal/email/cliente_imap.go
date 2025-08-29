package email

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"

	"github.com/transben/pendencias-downloader/internal/config"
)

// ClienteIMAP encapsula a conexão e operações com IMAP.
type ClienteIMAP struct {
	cfg    *config.Config
	client *client.Client
}

// NovoClienteIMAP cria um cliente IMAP baseado na configuração.
func NovoClienteIMAP(cfg *config.Config) *ClienteIMAP {
	return &ClienteIMAP{cfg: cfg}
}

// Conectar abre a conexão IMAP (com ou sem TLS) e faz login.
// Retorna erro se não conseguir autenticar.
func (c *ClienteIMAP) Conectar() error {
	endpoint := fmt.Sprintf("%s:%d", c.cfg.Email.Servidor, c.cfg.Email.Porta)

	var (
		cl  *client.Client
		err error
	)
	if c.cfg.Email.UsarTLS {
		cl, err = client.DialTLS(endpoint, &tls.Config{ServerName: c.cfg.Email.Servidor, MinVersion: tls.VersionTLS12})
	} else {
		cl, err = client.Dial(endpoint)
	}
	if err != nil {
		return fmt.Errorf("falha ao conectar IMAP: %w", err)
	}

	if err := cl.Login(c.cfg.Email.Usuario, c.cfg.Email.Senha); err != nil {
		_ = cl.Logout()
		return fmt.Errorf("falha ao autenticar IMAP: %w", err)
	}

	c.client = cl
	return nil
}

// Desconectar finaliza a sessão IMAP.
func (c *ClienteIMAP) Desconectar() {
	if c.client != nil {
		_ = c.client.Logout()
	}
}

// SelecionarCaixa seleciona a mailbox (ex.: "INBOX") para leitura.
func (c *ClienteIMAP) SelecionarCaixa(nome string) (*imap.MailboxStatus, error) {
	if c.client == nil {
		return nil, fmt.Errorf("cliente IMAP não conectado")
	}
	mbox, err := c.client.Select(nome, false)
	if err != nil {
		return nil, fmt.Errorf("erro ao selecionar caixa %s: %w", nome, err)
	}
	return mbox, nil
}

// BuscarMensagensAlvo retorna sequências UID das mensagens alvo.
func (c *ClienteIMAP) BuscarMensagensAlvo(soNaoLidos bool) ([]uint32, error) {
	if c.client == nil {
		return nil, fmt.Errorf("cliente IMAP não conectado")
	}
	criteria := imap.NewSearchCriteria()
	if soNaoLidos {
		criteria.WithoutFlags = []string{imap.SeenFlag}
	}
	// Poderíamos aplicar intervalos de data, termo no assunto, etc., futuramente.
	uids, err := c.client.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar mensagens: %w", err)
	}
	return uids, nil
}

// BaixarAnexosDeMensagens itera pelas mensagens (por UID) e chama o callback para cada anexo encontrado.
func (c *ClienteIMAP) BaixarAnexosDeMensagens(uids []uint32, onAnexo func(nome string, r io.Reader) error, marcarComoLido bool) error {
	if c.client == nil {
		return fmt.Errorf("cliente IMAP não conectado")
	}
	if len(uids) == 0 {
		return nil
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(uids...)

	section := &imap.BodySectionName{}
	items := []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope, imap.FetchUid}

	messages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.client.UidFetch(seqset, items, messages)
	}()

	for msg := range messages {
		if msg == nil {
			continue
		}
		body := msg.GetBody(section)
		if body == nil {
			continue
		}

		mr, err := mail.CreateReader(body)
		if err != nil {
			log.Printf("falha ao ler MIME UID=%d: %v", msg.Uid, err)
			continue
		}

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("falha ao iterar partes UID=%d: %v", msg.Uid, err)
				break
			}

			switch h := p.Header.(type) {
			case *mail.AttachmentHeader:
				filename, _ := h.Filename()
				if filename == "" {
					filename = fmt.Sprintf("anexo_%d_%d.bin", msg.Uid, time.Now().Unix())
				}
				if err := onAnexo(filename, p.Body); err != nil {
					log.Printf("erro ao salvar anexo (UID=%d, arquivo=%s): %v", msg.Uid, filename, err)
				}
			default:
				// Ignora texto inline, HTML, etc.
			}
		}

		// Marca como lido, se configurado
		if marcarComoLido {
			if err := c.marcarUIDComoLido(msg.Uid); err != nil {
				log.Printf("erro ao marcar como lida UID=%d: %v", msg.Uid, err)
			}
		}
	}

	if err := <-done; err != nil {
		return fmt.Errorf("erro no fetch IMAP: %w", err)
	}
	return nil
}

// marcarUIDComoLido aplica a flag \Seen na mensagem.
func (c *ClienteIMAP) marcarUIDComoLido(uid uint32) error {
	seq := new(imap.SeqSet)
	seq.AddNum(uid)
	item := imap.FormatFlagsOp(imap.AddFlags, true)
	flags := []interface{}{imap.SeenFlag}
	return c.client.UidStore(seq, item, flags, nil)
}
