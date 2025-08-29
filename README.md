# Pendências Downloader (Go)

Baixa **somente os anexos** de e‑mails via **IMAP** (preferencialmente **não lidos**) e salva em uma pasta local (ex.: `C:\Pendencias`).  
Roda em background com agendador (a cada _N_ minutos) e expõe um **painel web local** para status, logs e **edição dinâmica de configurações**.

---

## ✨ Recursos

- ⏱️ Agendador: executa a cada `intervalo_minutos` (dinâmico)
- 📎 Só baixa **anexos** (ignora corpo/HTML)
- 📥 Filtra **não lidos** (configurável)
- 🗂️ Destino configurável (ex.: `C:\Pendencias`)
- 🧾 Logs diários em `logs/YYYY-MM-DD.log`
- 🖥️ Painel local: `http://localhost:8080`
  - Ver últimos logs
  - **Executar agora**
  - **Editar configurações** (salvas em `config.yaml` e aplicadas em runtime)
- ⚙️ Configuração por **YAML** + **.env** (o `.env` vence o YAML no start)
- 🔐 Compatível com **Gmail (Senha de App)** e provedores IMAP comuns
- 🧩 Arquitetura modular, separada por responsabilidades

---

## 🧱 Estrutura do projeto

```
pendencias-downloader/
├─ cmd/
│  └─ pendencias-downloader/
│     └─ main.go
├─ internal/
│  ├─ agendador/
│  │  └─ agendador.go
│  ├─ config/
│  │  ├─ config.go
│  │  └─ store.go
│  ├─ email/
│  │  └─ cliente_imap.go
│  ├─ logger/
│  │  └─ logger.go
│  ├─ repositorio/
│  │  └─ arquivos.go
│  ├─ servico/
│  │  └─ servico_baixar_anexos.go
│  └─ ui/
│     ├─ handlers.go
│     └─ assets/
│        ├─ index.html
│        └─ js/
│           ├─ app.js
│           ├─ api.js
│           └─ ui.js
├─ configs/
│  ├─ config.yaml           # (gitignored)
│  └─ config.yaml.example   # (exemplo)
├─ .env                     # (gitignored)
├─ .env.example
├─ .gitignore
├─ go.mod
└─ README.md
```

---

## 🚀 Como rodar

### Pré‑requisitos
- Go **1.22+**

### Passo a passo
```bash
# 1) Dependências
go mod tidy

# 2) Arquivos locais (exemplos → reais)
cp configs/config.yaml.example configs/config.yaml
cp .env.example .env

# 3) Edite configs/config.yaml e/ou .env

# 4) Executar
go run ./cmd/pendencias-downloader
# Painel: http://localhost:8080
```

> **Windows (caminhos):** no YAML use `"C:\\Pendencias"` (duplo `\`).  
> **.env:** use aspas em valores com espaços, ex.: `EMAIL_SENHA="abcd efgh ijkl mnop"`.

---

## ⚙️ Configuração

### `configs/config.yaml` (exemplo)
```yaml
pasta_destino: "C:\\Pendencias"
intervalo_minutos: 10
email:
  servidor: imap.gmail.com
  porta: 993
  usuario: seu.email@dominio.com
  senha: ""          # recomendado deixar vazio e usar .env
  usar_tls: true
marcar_como_lido: true
so_nao_lidos: true
caixa: INBOX
```

### `.env` (override no start — tem prioridade sobre o YAML)
> Se a variável existir no `.env`, **ela vence** o YAML quando a aplicação inicia.

```env
PASTA_DESTINO="C:\Pendencias"
INTERVALO_MINUTOS=10
CAIXA=INBOX
MARCAR_COMO_LIDO=true
SO_NAO_LIDOS=true

EMAIL_SERVIDOR=imap.gmail.com
EMAIL_PORTA=993
EMAIL_USUARIO=seu.email@dominio.com
EMAIL_SENHA="abcd efgh ijkl mnop"   # Senha de App (Gmail)
EMAIL_USAR_TLS=true
```

**Dicas:**
- Mantenha **segredos** no `.env`; o painel salva no YAML (que está gitignored).
- `.env` é lido **apenas no start**; para aplicar mudanças no `.env`, reinicie o processo.
- Valores do YAML podem ser **editados pelo painel** e têm efeito imediato (o agendador reinicia se mudar o intervalo).

---

## 🖥️ Painel (UI)

- URL: `http://localhost:8080`
- **Ações**:
  - **Executar agora**: dispara uma execução imediata
  - **Logs**: exibe as últimas linhas do log do dia
  - **Configurações**: edita e persiste no `config.yaml`
- **APIs internas**:
  - `GET /api/config` → configuração atual (senha mascarada)
  - `PUT /api/config` → atualiza config (se incluir `intervalo_minutos`, reinicia o ticker)
  - `GET /api/logs?linhas=200` → últimas N linhas
  - `POST /api/executar` → executa uma vez

---

## 🪵 Logs

- Arquivos diários: `logs/YYYY-MM-DD.log`
- Rotação automática quando vira o dia
- O painel lê sempre o arquivo do **dia atual**

---

## 🔐 Gmail / Google Workspace

- Ative **IMAP** no Gmail (web).
- Ative **2FA** e gere **Senha de App** (IMAP não aceita “apps menos seguros”).
- Host/porta: `imap.gmail.com:993`, TLS = true.
- Em domínios corporativos, o admin pode **bloquear IMAP**; valide com TI.

---

## 📦 Build

```bash
# Linux
GOOS=linux   GOARCH=amd64 go build -o bin/pendencias ./cmd/pendencias-downloader

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/pendencias.exe ./cmd/pendencias-downloader

# macOS (Intel)
GOOS=darwin  GOARCH=amd64 go build -o bin/pendencias-macos ./cmd/pendencias-downloader
```

---

## 🔒 Segurança / Git

Adicione no **.gitignore**:
```
.env
configs/config.yaml
logs/
```

Versione apenas os exemplos:
- `configs/config.yaml.example`
- `.env.example`

Se segredos foram comitados por engano:
```bash
git rm --cached configs/config.yaml .env
git commit -m "Stop tracking secrets; ignore real config files"
git push
# → e rotacione as credenciais
```

---

## 🧪 Troubleshooting

**1) Gmail: `Application-specific password required`**  
→ Gere **Senha de App** após ativar **2FA** e use no lugar da senha normal.

**2) `Failed to load module script (MIME type text/html)` ao carregar `/static/js/app.js`**  
→ Caminho/“case” errado ou embed faltando. Garanta:
- `//go:embed assets/index.html assets/js/*.js`
- árvore: `internal/ui/assets/index.html` e `internal/ui/assets/js/app.js`
- `<script type="module" src="/static/js/app.js"></script>`
- Recompile após mover arquivos.

**3) `unknown revision` / versões quebradas no Go**  
→ Use `go get ...@latest` e `go mod tidy`. Evite travar em versões que não existem.

**4) Sem permissão para gravar em `C:\Pendencias`** (Windows)  
→ Rode o terminal como **Administrador** ou escolha outra pasta com permissão de escrita.

**5) Porta 8080 já em uso**  
→ Altere a porta do painel no código (const `portaPainel`) ou crie uma chave no YAML para isso.

**6) IMAP conecta mas não baixa anexos**  
→ Verifique filtros (`so_nao_lidos`, `caixa`) e se as mensagens realmente têm anexos.

---

