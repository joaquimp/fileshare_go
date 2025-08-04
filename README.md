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

```
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

### Iniciar o servidor

```bash
go run .
```

O servidor iniciará em `http://localhost:8080`

### Fazer upload de um arquivo

```bash
curl -F "file=@meuarquivo.txt" http://localhost:8080/upload
```

Resposta:
```
✅ Arquivo enviado com sucesso!
📥 Link para download: http://localhost:8080/file/a1b2c3d4e5f6
⚠️  Atenção: O arquivo será removido após o primeiro download.
```

### Fazer download

Acesse o link retornado ou use curl:

```bash
curl -O -J http://localhost:8080/file/a1b2c3d4e5f6
```

## 📡 Endpoints

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/` | GET | Página de instruções |
| `/upload` | POST | Upload de arquivo (campo: `file`) |
| `/file/{token}` | GET | Download do arquivo |
| `/status` | GET | Status do servidor |

## 🔧 Configuração

O servidor pode ser configurado através de variáveis de ambiente:

| Variável | Descrição | Padrão |
|----------|-----------|--------|
| `PORT` | Porta do servidor | `8080` |
| `BASE_URL` | URL base para links de download | `http://localhost:8080` |
| `STORAGE_PATH` | Diretório de armazenamento | `./uploads` |
| `MAX_FILE_SIZE_MB` | Tamanho máximo em MB | `5` |

### Exemplo de uso com variáveis de ambiente

```bash
# Definir limite de 10MB e porta 3000
export MAX_FILE_SIZE_MB=10
export PORT=3000
export BASE_URL=http://localhost:3000
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

- Tokens gerados com `crypto/rand` (criptograficamente seguros)
- Sanitização automática de nomes de arquivos
- Validação de métodos HTTP
- Remoção automática de arquivos após download

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

# Fazer upload
curl -F "file=@teste.txt" http://localhost:8080/upload
```

### Status
```bash
curl http://localhost:8080/status
```

### Página principal
Acesse `http://localhost:8080` no navegador para ver as instruções.

## 📝 TODO / Melhorias Futuras

- [ ] Configuração via variáveis de ambiente
- [ ] Logs estruturados
- [ ] Métricas e monitoramento
- [ ] Autenticação opcional
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
