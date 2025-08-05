# 🐳 Guia Portainer - FileShare Go

Este guia mostra como deployar a aplicação FileShare Go usando Portainer.

## 📋 Pré-requisitos

- Portainer instalado e funcionando
- Acesso ao GitHub Container Registry (público)
- Conhecimento básico de Docker

## 🚀 Deploy via Portainer

### Método 1: Stack com Docker Compose

1. **Acesse Portainer**
   - Faça login no seu Portainer
   - Navegue para **Stacks** no menu lateral

2. **Criar Nova Stack**
   - Clique em **Add Stack**
   - Dê um nome: `fileshare-go`

3. **Configurar Stack**
   - Selecione **Web editor**
   - Cole o seguinte Docker Compose:

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
      - PORT=8080
      - BASE_URL=http://seu-servidor.com:8080
      - STORAGE_PATH=./uploads
      - MAX_FILE_SIZE_MB=10
      - API_KEY=ALTERE_ESTA_CHAVE_SECRETA_DE_32_CARACTERES
      - ALLOWED_USER_AGENT=ADAMA/1.0
    volumes:
      - fileshare_uploads:/app/uploads
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.fileshare.rule=Host(`files.seu-dominio.com`)"

volumes:
  fileshare_uploads:
    driver: local
```

4. **Configurar Variáveis de Ambiente**
   
   ⚠️ **IMPORTANTE**: Altere os seguintes valores:
   
   - `BASE_URL`: Coloque a URL real do seu servidor
   - `API_KEY`: Gere uma chave segura de 32 caracteres
   - `MAX_FILE_SIZE_MB`: Ajuste conforme necessário
   - Labels do Traefik (se usar): Configure seu domínio

5. **Gerar API Key Segura**
   
   Use um dos métodos abaixo:
   
   ```bash
   # Linux/macOS (Terminal)
   openssl rand -hex 32
   
   # Online (use sites confiáveis)
   # https://www.random.org/strings/
   # Configurar: 1 string, 32 caracteres, hex
   ```

6. **Deploy**
   - Clique em **Deploy the stack**
   - Aguarde o download da imagem e inicialização

### Método 2: Container Individual

1. **Acesse Containers**
   - Navegue para **Containers** no menu

2. **Adicionar Container**
   - Clique em **Add container**
   - Nome: `fileshare-server`

3. **Configurações da Imagem**
   - Image: `ghcr.io/joaquimp/fileshare_go:latest`
   - Always pull: ✅ Marcar

4. **Network & ports**
   - Port mapping:
     - Host: `8080`
     - Container: `8080`

5. **Volumes**
   - Volume mapping:
     - Container: `/app/uploads`
     - Bind: `/opt/fileshare/uploads` (ou outro caminho)

6. **Environment variables**

   ```bash
   PORT=8080
   BASE_URL=http://seu-servidor.com:8080
   MAX_FILE_SIZE_MB=10
   API_KEY=sua_chave_secreta_de_32_caracteres
   ALLOWED_USER_AGENT=ADAMA/1.0
   ```

7. **Restart policy**
   - Selecione: **Unless stopped**

8. **Deploy**
   - Clique em **Deploy the container**

## 🔧 Configuração Avançada

### Usando Traefik (Reverse Proxy)

Se você usa Traefik, adicione estas labels ao serviço:

```yaml
labels:
  - "traefik.enable=true"
  - "traefik.http.routers.fileshare.rule=Host(`files.seu-dominio.com`)"
  - "traefik.http.routers.fileshare.tls=true"
  - "traefik.http.routers.fileshare.tls.certresolver=letsencrypt"
  - "traefik.http.services.fileshare.loadbalancer.server.port=8080"
```

### HTTPS com Nginx Proxy Manager

1. **Configure o container** normalmente
2. **No Nginx Proxy Manager**:
   - Domain: `files.seu-dominio.com`
   - Forward Hostname: IP do Docker host
   - Forward Port: `8080`
   - Ative SSL/TLS

### Volumes Persistentes

Para dados persistentes, configure volumes:

```yaml
volumes:
  fileshare_uploads:
    driver: local
    driver_opts:
      type: none
      o: bind
      device: /opt/docker/fileshare/uploads
```

## 📊 Monitoramento

### Logs no Portainer

1. **Acesse o container**
2. **Clique em Logs**
3. **Configure auto-refresh** para monitoramento em tempo real

### Health Check

Adicione um health check ao container:

```yaml
healthcheck:
  test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/status"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 30s
```

## 🔒 Segurança

### Rede Docker

Crie uma rede isolada:

```yaml
networks:
  fileshare_network:
    driver: bridge
    internal: false
```

### Firewall

Configure firewall para expor apenas portas necessárias:

```bash
# UFW exemplo
sudo ufw allow 8080/tcp
sudo ufw reload
```

### API Key Rotation

Para trocar a API key:

1. **Gere nova chave**
2. **Edite a stack** no Portainer
3. **Atualize** a variável `API_KEY`
4. **Redeploy** a stack

## 🧪 Testando

### Verificar Status

```bash
curl http://seu-servidor.com:8080/status
```

### Upload de Teste

```bash
curl -X POST \
  -H "Authorization: Bearer SUA_API_KEY" \
  -H "User-Agent: ADAMA/1.0" \
  -F "file=@teste.txt" \
  http://seu-servidor.com:8080/upload
```

**Resposta JSON:**

```json
{
  "success": true,
  "message": "Arquivo enviado com sucesso",
  "token": "a1b2c3d4e5f6",
  "filename": "teste.txt",
  "download_url": "http://seu-servidor.com:8080/file/a1b2c3d4e5f6"
}
```

### Download de Teste

```bash
curl -O -J http://seu-servidor.com:8080/file/a1b2c3d4e5f6
```

**Headers de Download:**

- `Content-Type: text/plain` (detectado automaticamente)
- `Content-Disposition: attachment; filename="teste.txt"`
- `Content-Length: 1024`

## 🚨 Troubleshooting

### Container não inicia

1. **Verifique logs** no Portainer
2. **Confirme** se a API_KEY está definida
3. **Teste** se a porta 8080 está livre

### 401 Unauthorized

1. **Verifique** se a API_KEY está correta
2. **Confirme** o header `Authorization: Bearer`
3. **Teste** com `User-Agent: ADAMA/1.0`

### Uploads falhando

1. **Verifique** o tamanho do arquivo vs `MAX_FILE_SIZE_MB`
2. **Confirme** se o volume está montado corretamente
3. **Verifique** permissões da pasta de uploads

### Performance

1. **Monitor** uso de CPU/RAM no Portainer
2. **Ajuste** `MAX_FILE_SIZE_MB` conforme recursos
3. **Considere** usar SSD para volume de uploads

## 📞 Suporte

- **Documentação**: [README.md](./README.md)
- **Autenticação**: [AUTH_GUIDE.md](./AUTH_GUIDE.md)
- **Docker**: [DOCKER_DEPLOY.md](./DOCKER_DEPLOY.md)
- **Issues**: GitHub Issues do projeto

## 📈 Atualizações

Para atualizar para uma nova versão:

1. **Edit Stack** no Portainer
2. **Altere** a tag da imagem se necessário
3. **Click** em **Update the stack**
4. **Pull and redeploy** será executado automaticamente
