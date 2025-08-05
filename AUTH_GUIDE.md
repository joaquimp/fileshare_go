# 🔐 Guia de Autenticação - File Share Server

Este guia explica como configurar e usar a autenticação no File Share Server.

## 🔑 Configuração da API Key

### 1. Gerar uma API Key Segura

```bash
# Método 1: Usando openssl (recomendado)
openssl rand -hex 32

# Método 2: Usando uuidgen (macOS/Linux)
uuidgen

# Método 3: Online (https://www.uuidgenerator.net/)
```

### 2. Configurar Variáveis de Ambiente

```bash
# Copiar arquivo de exemplo
cp .env.example .env

# Editar .env
nano .env
```

**Exemplo de .env:**

```bash
API_KEY=a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456
ALLOWED_USER_AGENT=MeuAppIOS/1.0
MAX_FILE_SIZE_MB=5
PORT=8080
```

## 📱 Implementação no iOS

### Swift - URLSession

```swift
import Foundation

class FileUploadService {
    private let apiKey = "sua_api_key_aqui"
    private let baseURL = "http://seu-servidor.com:8080"

    func uploadFile(fileURL: URL, completion: @escaping (Result<String, Error>) -> Void) {
        guard let uploadURL = URL(string: "\(baseURL)/upload") else {
            completion(.failure(URLError(.badURL)))
            return
        }

        var request = URLRequest(url: uploadURL)
        request.httpMethod = "POST"

        // Header de autenticação (OBRIGATÓRIO)
        request.setValue("Bearer \(apiKey)", forHTTPHeaderField: "Authorization")

        // User-Agent personalizado (OPCIONAL)
        request.setValue("MeuAppIOS/1.0", forHTTPHeaderField: "User-Agent")

        // Criar form data
        let boundary = UUID().uuidString
        request.setValue("multipart/form-data; boundary=\(boundary)", forHTTPHeaderField: "Content-Type")

        let formData = createFormData(fileURL: fileURL, boundary: boundary)

        URLSession.shared.uploadTask(with: request, from: formData) { data, response, error in
            // Processar resposta...
        }.resume()
    }

    private func createFormData(fileURL: URL, boundary: String) -> Data {
        var formData = Data()

        // Adicionar arquivo
        formData.append("--\(boundary)\r\n".data(using: .utf8)!)
        formData.append("Content-Disposition: form-data; name=\"file\"; filename=\"\(fileURL.lastPathComponent)\"\r\n".data(using: .utf8)!)
        formData.append("Content-Type: application/octet-stream\r\n\r\n".data(using: .utf8)!)

        if let fileData = try? Data(contentsOf: fileURL) {
            formData.append(fileData)
        }

        formData.append("\r\n--\(boundary)--\r\n".data(using: .utf8)!)

        return formData
    }
}
```

## 🧪 Testes com curl

### ❌ Upload sem autenticação (falha)

```bash
curl -F "file=@teste.txt" http://localhost:8080/upload
# Retorna: {"error":"Unauthorized","message":"API key inválida ou aplicativo não autorizado"}
```

### ✅ Upload com autenticação (sucesso)

```bash
curl -H "Authorization: Bearer sua_api_key_aqui" \
     -F "file=@teste.txt" \
     http://localhost:8080/upload
# Retorna: Link de download
```

### ✅ Download (público - sem autenticação)

```bash
curl -O -J http://localhost:8080/file/TOKEN_RETORNADO
```

### 📊 Status do servidor

```bash
curl http://localhost:8080/status
# Retorna: {"status": "ok", "service": "file-share-server", "max_file_size_mb": 5.0, "auth_required": true}
```

## 🛡️ Segurança Implementada

### 1. **API Key Validation**

- Header `Authorization: Bearer YOUR_API_KEY` obrigatório
- Comparação segura contra timing attacks (`subtle.ConstantTimeCompare`)
- API Key mascarada nos logs

### 2. **User-Agent Filtering (Opcional)**

- Valida se requisição vem do aplicativo correto
- Configurável via `ALLOWED_USER_AGENT`

### 3. **Endpoints Protegidos**

- ✅ `/upload` - PROTEGIDO (requer autenticação)
- ✅ `/file/*` - PÚBLICO (download sem autenticação)
- ✅ `/status` - PÚBLICO (informações do servidor)
- ✅ `/` - PÚBLICO (página de instruções)

## 🚨 Troubleshooting

### Servidor não inicia

```text
❌ API_KEY deve ser configurada para proteger o servidor
```

**Solução:** Configure a variável `API_KEY` no .env ou como variável de ambiente.

### Upload retorna 401 Unauthorized

```json
{"error":"Unauthorized","message":"API key inválida ou aplicativo não autorizado"}
```

**Possíveis causas:**

1. Header `Authorization` ausente
2. Formato incorreto (deve ser `Bearer YOUR_API_KEY`)
3. API Key incorreta
4. User-Agent não permitido (se configurado)

### Como verificar configuração

```bash
curl http://localhost:8080/status
```

## 🔧 Configurações Avançadas

### Múltiplos User-Agents

Atualmente suporta apenas um User-Agent. Para múltiplos, modifique a função `ValidateRequest` em `auth.go`.

### Rate Limiting

Para implementar rate limiting, adicione middleware antes da autenticação.

### HTTPS

Para produção, configure proxy reverso (nginx) com SSL/TLS.

## 📚 Exemplo Completo

1 - **Configurar servidor:**

```bash
export API_KEY="$(openssl rand -hex 32)"
export ALLOWED_USER_AGENT="MeuAppIOS/1.0"
go run .
```

2 - **Upload do iOS:**

```swift
// Usar o código Swift mostrado acima
```

3 - **Verificar upload:**

```bash
curl -H "Authorization: Bearer $API_KEY" \
     -H "User-Agent: MeuAppIOS/1.0" \
     -F "file=@teste.txt" \
     http://localhost:8080/upload
```

## 🎯 Resumo

- ✅ Apenas aplicativos com API Key válida podem fazer upload
- ✅ Downloads são públicos (não requerem autenticação)
- ✅ User-Agent opcional para maior segurança
- ✅ Logs seguros (API Key mascarada)
- ✅ Proteção contra timing attacks
