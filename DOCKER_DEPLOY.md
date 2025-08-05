# FileShare Go - Docker Deployment

Este guia explica como usar a imagem Docker da aplicação FileShare Go.

## Usando a Imagem Docker

### 1. Docker Run

```bash
# Executar com configurações básicas
docker run -d \
  --name fileshare_server \
  -p 8080:8080 \
  -e API_KEY=sua_api_key_super_secreta_aqui \
  -e ALLOWED_USER_AGENT=ADAMA/1.0 \
  -v $(pwd)/uploads:/app/uploads \
  ghcr.io/joaquimp/fileshare_go:latest
```

### 2. Docker Compose

Copie o arquivo `docker-compose.yml` e ajuste as variáveis de ambiente:

```bash
# Iniciar com docker-compose
docker-compose up -d

# Parar o serviço
docker-compose down
```

### 3. Portainer

Para usar no Portainer:

1. Acesse Portainer
2. Vá em **Stacks** > **Add Stack**
3. Cole o conteúdo do `docker-compose.yml` ou use o repositório Git
4. Ajuste as variáveis de ambiente conforme necessário
5. Deploy

## Variáveis de Ambiente

| Variável | Padrão | Descrição |
|----------|---------|-----------|
| `PORT` | `8080` | Porta do servidor |
| `BASE_URL` | `http://localhost:8080` | URL base para downloads |
| `STORAGE_PATH` | `./uploads` | Diretório de uploads |
| `MAX_FILE_SIZE_MB` | `5` | Tamanho máximo em MB |
| `API_KEY` | **OBRIGATÓRIA** | Chave de autenticação |
| `ALLOWED_USER_AGENT` | - | User-Agent permitido |

## Volumes

- `/app/uploads` - Diretório onde os arquivos são armazenados

## Portas

- `8080` - Porta HTTP do servidor

## Segurança

⚠️ **IMPORTANTE**: Sempre altere a `API_KEY` padrão antes de usar em produção!

```bash
# Gerar uma API key segura
openssl rand -hex 32
```

## URLs Disponíveis

- `http://localhost:8080/` - Interface web com documentação
- `http://localhost:8080/status` - Status da aplicação
- `http://localhost:8080/upload` - Endpoint para upload (POST) - **Retorna JSON**
- `http://localhost:8080/file/{token}` - Download de arquivos - **MIME type correto**

## MIME Types Suportados

A aplicação detecta automaticamente o MIME type correto baseado na extensão:

| Extensão | MIME Type | Categoria |
|----------|-----------|-----------|
| `.pdf` | `application/pdf` | Documentos |
| `.doc/.docx` | `application/msword` | Office |
| `.jpg/.jpeg` | `image/jpeg` | Imagens |
| `.png` | `image/png` | Imagens |
| `.mp4` | `video/mp4` | Vídeos |
| `.mp3` | `audio/mpeg` | Áudio |
| `.txt` | `text/plain` | Texto |
| `.json` | `application/json` | Dados |
| `.zip` | `application/zip` | Compactados |
| Outros | `application/octet-stream` | Genérico |

**Features do Download:**

- ✅ MIME type correto detectado automaticamente
- ✅ Extensão adicionada se necessário
- ✅ Headers apropriados (Content-Length, Cache-Control)
- ✅ Nome de arquivo preservado

## Exemplo de Configuração Portainer

### Stack Configuration

```yaml
version: '3.8'
services:
  fileshare:
    image: ghcr.io/joaquimp/fileshare_go:latest
    container_name: fileshare_server
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      PORT: "8080"
      BASE_URL: "https://seu-dominio.com"
      MAX_FILE_SIZE_MB: "10"
      API_KEY: "cole_sua_api_key_aqui"
      ALLOWED_USER_AGENT: "ADAMA/1.0"
    volumes:
      - fileshare_uploads:/app/uploads

volumes:
  fileshare_uploads:
```

### Environment Variables no Portainer

No Portainer, você pode definir as variáveis na seção **Environment**:

```bash
API_KEY=sua_chave_api_super_secreta_de_32_caracteres
ALLOWED_USER_AGENT=ADAMA/1.0
BASE_URL=https://seu-dominio.com
MAX_FILE_SIZE_MB=10
PORT=8080
```

## Logs

Para visualizar os logs do container:

```bash
# Docker
docker logs fileshare_server

# Docker Compose
docker-compose logs fileshare

# Portainer
# Acesse Container > Logs na interface web
```

## Healthcheck

A aplicação responde no endpoint `/status` para verificação de saúde:

```bash
curl http://localhost:8080/status
```

Resposta esperada:

```json
{
  "status": "ok",
  "version": "1.0.0"
}
```
