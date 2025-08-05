# 🗂️ File Share Server

Um servidor simples em Go para compartilhamento temporário de arquivos com tokens seguros.

## 🌟 Características

- ✅ **Upload seguro**: Aceita arquivos até 10MB
- 🔐 **Tokens criptográficos**: Geração segura de tokens para acesso
- 🗑️ **Uso único**: Arquivos são removidos após o primeiro download
- 🔒 **Thread-safe**: Proteção contra condições de corrida
- 🧹 **Sanitização**: Nomes de arquivos são sanitizados automaticamente
- 📊 **Status**: Endpoint para verificar saúde do servidor

## 🏗️ Estrutura do Projeto

```text
fileShare_go/
├── main.go         # Ponto de entrada e configuração do servidor
├── config.go       # Gerenciamento de configurações via variáveis de ambiente
├── handlers.go     # Handlers HTTP para upload e download
├── types.go        # Estruturas de dados e tipos
├── utils.go        # Funções utilitárias
├── go.mod          # Dependências do Go
├── .env.example    # Exemplo de configuração
├── .gitignore      # Arquivos ignorados pelo Git
└── README.md       # Esta documentação
```

## 🚀 Como Usar

### Opção 1: Docker (Recomendado)

```bash
# Executar com Docker
docker run -d \
  --name fileshare_server \
  -p 8080:8080 \
  -e API_KEY=sua_api_key_super_secreta_aqui \
  -e ALLOWED_USER_AGENT=ADAMA/1.0 \
  -v $(pwd)/uploads:/app/uploads \
  ghcr.io/joaquimp/fileshare_go:latest
```

Para mais detalhes sobre deployment com Docker, veja [DOCKER_DEPLOY.md](./DOCKER_DEPLOY.md).

### Opção 2: Executar localmente

### Configurar autenticação

Primeiro, você precisa configurar a autenticação. Copie o arquivo de exemplo:

```bash
cp .env.example .env
```

Edite o arquivo `.env` e configure sua API key:

```bash
# Gerar uma API key segura
openssl rand -hex 32

# Adicionar ao arquivo .env
API_KEY=sua_api_key_gerada_aqui
ALLOWED_USER_AGENT=ADAMA/1.0
```

### Iniciar o servidor

```bash
go run .
```

O servidor iniciará em `http://localhost:8080`

### Fazer upload de um arquivo

⚠️ **Atenção**: Agora é necessário incluir a API key nos headers:

```bash
curl -X POST \
  -H "Authorization: Bearer sua_api_key_aqui" \
  -H "User-Agent: ADAMA/1.0" \
  -F "file=@meuarquivo.txt" \
  http://localhost:8080/upload
```

Resposta JSON:

```json
{
  "success": true,
  "message": "Arquivo enviado com sucesso",
  "token": "a1b2c3d4e5f6",
  "filename": "meuarquivo.txt",
  "safe_filename": "meuarquivo.txt",
  "size_bytes": 1024,
  "size_mb": 0.001,
  "download_url": "http://localhost:8080/file/a1b2c3d4e5f6",
  "uploaded_at": "2025-08-05T10:30:45Z",
  "note": "O arquivo será removido após o primeiro download"
}
```

### Fazer download

Use o token retornado no upload:

```bash
curl -O -J http://localhost:8080/file/a1b2c3d4e5f6
```

O arquivo será servido com:

- **MIME type correto** (detectado pela extensão)
- **Nome original** preservado
- **Headers apropriados** para download

## 📡 Endpoints

| Endpoint | Método | Descrição | Autenticação |
|----------|--------|-----------|--------------|
| `/` | GET | Página de instruções | ❌ Pública |
| `/upload` | POST | Upload de arquivo (campo: `file`) | ✅ API Key |
| `/file/{token}` | GET | Download do arquivo | ❌ Pública |
| `/status` | GET | Status do servidor | ❌ Pública |

## 🔧 Configuração

O servidor pode ser configurado através de variáveis de ambiente:

| Variável | Descrição | Padrão | Obrigatória |
|----------|-----------|--------|----|
| `PORT` | Porta do servidor | `8080` | ❌ |
| `BASE_URL` | URL base para links de download | `http://localhost:8080` | ❌ |
| `STORAGE_PATH` | Diretório de armazenamento | `./uploads` | ❌ |
| `MAX_FILE_SIZE_MB` | Tamanho máximo em MB | `5` | ❌ |
| `API_KEY` | Chave de autenticação | - | ✅ |
| `ALLOWED_USER_AGENT` | User-Agent permitido | - | ❌ |

### Exemplo de uso com variáveis de ambiente

```bash
# Definir limite de 10MB e porta 3000
export MAX_FILE_SIZE_MB=10
export PORT=3000
export BASE_URL=http://localhost:3000
export API_KEY=sua_api_key_super_secreta_aqui
go run .
```

### Arquivo .env

Você pode copiar o arquivo `.env.example` para `.env` e ajustar as configurações:

```bash
cp .env.example .env
# Edite o arquivo .env conforme necessário
```

### Limites

- **Tamanho máximo**: Configurável via `MAX_FILE_SIZE_MB` (padrão: 5MB)
- **Token**: 16 caracteres hexadecimais (8 bytes)
- **Uso**: Cada arquivo pode ser baixado apenas uma vez

## 🔒 Segurança

- **Autenticação**: API Key obrigatória para uploads
- **User-Agent**: Validação opcional para maior segurança
- Tokens gerados com `crypto/rand` (criptograficamente seguros)
- Sanitização automática de nomes de arquivos
- Validação de métodos HTTP
- Remoção automática de arquivos após download

Para configuração detalhada da autenticação, veja [AUTH_GUIDE.md](./AUTH_GUIDE.md).

## 🛠️ Melhorias Implementadas

### Reorganização do Código

- **Separação de responsabilidades**: Código dividido em arquivos lógicos
- **Estruturas próprias**: `FileStorage` e `Server` para encapsulamento
- **Funções utilitárias**: Isoladas em arquivo próprio

### Segurança Aprimorada

- **Tokens criptográficos**: Substituição do `math/rand` por `crypto/rand`
- **Sanitização**: Proteção contra nomes de arquivo maliciosos
- **Validação de métodos**: Verificação de GET/POST

### Experiência do Usuário

- **Mensagens melhoradas**: Feedback mais claro e amigável
- **Página de instruções**: Interface web com documentação
- **Endpoint de status**: Monitoramento da saúde do servidor

### Robustez

- **Tratamento de erros**: Validações mais abrangentes
- **Cleanup**: Remoção de arquivos parciais em caso de erro
- **Thread safety**: Proteção adequada para concorrência

## 🧪 Testando

### Upload

```bash
# Criar um arquivo de teste
echo "Conteúdo de teste" > teste.txt

# Fazer upload (com autenticação)
curl -X POST \
  -H "Authorization: Bearer sua_api_key_aqui" \
  -H "User-Agent: ADAMA/1.0" \
  -F "file=@teste.txt" \
  http://localhost:8080/upload
```

### Status

```bash
curl http://localhost:8080/status
```

### Página principal

Acesse `http://localhost:8080` no navegador para ver as instruções.

## 📝 TODO / Melhorias Futuras

- [x] ~~Configuração via variáveis de ambiente~~
- [x] ~~Autenticação com API Key~~
- [x] ~~Logs estruturados~~
- [x] ~~Imagem Docker~~
- [ ] Métricas e monitoramento
- [ ] Interface web para upload
- [ ] Expiração automática de arquivos por tempo
- [ ] Suporte a HTTPS
- [ ] Rate limiting
- [ ] Compressão de arquivos

## 🤝 Contribuindo

1. Faça um fork do projeto
2. Crie uma branch para sua feature
3. Commit suas mudanças
4. Abra um Pull Request

## 📄 Licença

Este projeto está sob a licença MIT. Veja o arquivo LICENSE para detalhes.
