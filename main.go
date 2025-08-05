package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Carrega as configurações das variáveis de ambiente
	config := LoadConfig()

	// Verifica se API Key está configurada
	if config.APIKey == "" {
		log.Fatal("❌ API_KEY deve ser configurada para proteger o servidor")
	}

	// Cria o diretório de uploads se não existir
	err := os.MkdirAll(config.StoragePath, 0755)
	if err != nil {
		log.Fatalf("❌ Erro ao criar diretório de uploads: %v", err)
	}

	// Inicializa o storage de arquivos com a configuração
	storage := NewFileStorage(config.StoragePath, config.MaxFileSize)

	// Inicializa o servidor
	server := NewServer(storage, config.BaseURL)

	// Inicializa a autenticação
	auth := NewAuthConfig(config.APIKey, config.UserAgent)

	// Configura as rotas PROTEGIDAS
	http.HandleFunc("/upload", auth.RequireAuth(server.uploadHandler))

	// Rotas PÚBLICAS (download e status)
	http.HandleFunc("/file/", server.downloadHandler)

	// Adiciona uma rota de status para verificar se o servidor está funcionando
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		maxSizeMB := float64(config.MaxFileSize) / (1024 * 1024)
		authRequired := config.APIKey != ""
		fmt.Fprintf(w, `{"status": "ok", "service": "file-share-server", "max_file_size_mb": %.1f, "auth_required": %t}`, maxSizeMB, authRequired)
	})

	// Rota de instruções para a raiz
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		maxSizeMB := float64(config.MaxFileSize) / (1024 * 1024)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>File Share Server - Protegido</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>🔐 File Share Server (Protegido)</h1>
    <p>Servidor de compartilhamento temporário de arquivos com autenticação</p>
    
    <h2>🚨 Acesso Restrito</h2>
    <p><strong>Este servidor requer autenticação via API Key.</strong></p>
    <p>Apenas aplicativos autorizados podem fazer upload.</p>
    
    <h2>📤 Como fazer upload (aplicativo autorizado):</h2>
    <p>Envie uma requisição POST para <code>/upload</code> com:</p>
    <ul>
        <li>Arquivo no campo <code>file</code></li>
        <li>Header: <code>Authorization: Bearer YOUR_API_KEY</code></li>
        <li>User-Agent correto (se configurado)</li>
    </ul>
    <p><strong>Tamanho máximo:</strong> %.1f MB</p>
    
    <h3>Exemplo com curl:</h3>
    <pre><code>curl -H "Authorization: Bearer YOUR_API_KEY" \
     -F "file=@meuarquivo.txt" \
     %s/upload</code></pre>
    
    <h2>📥 Downloads:</h2>
    <p>Links de download são públicos e não requerem autenticação.</p>
    
    <h2>🔍 Status:</h2>
    <p><a href="/status">/status</a> - Verificar se o servidor está funcionando</p>
    
    <h2>⚙️ Configurações:</h2>
    <ul>
        <li><strong>Autenticação:</strong> Habilitada</li>
        <li><strong>Tamanho máximo:</strong> %.1f MB</li>
        <li><strong>User-Agent filtro:</strong> %s</li>
    </ul>
</body>
</html>`, maxSizeMB, config.BaseURL, maxSizeMB, config.UserAgent)
	})

	serverAddr := ":" + config.Port
	fmt.Printf("🔐 File Share Server iniciado em %s\n", config.BaseURL)
	if config.UserAgent != "" {
		fmt.Printf("📱 Restrito ao User-Agent: %s\n", config.UserAgent)
	}
	fmt.Printf("ℹ️  Acesse %s para instruções\n", config.BaseURL)

	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
