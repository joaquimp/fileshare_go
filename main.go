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

	// Cria o diretório de uploads se não existir
	err := os.MkdirAll(config.StoragePath, 0755)
	if err != nil {
		log.Fatalf("❌ Erro ao criar diretório de uploads: %v", err)
	}

	// Inicializa o storage de arquivos com a configuração
	storage := NewFileStorage(config.StoragePath, config.MaxFileSize)

	// Inicializa o servidor
	server := NewServer(storage, config.BaseURL)

	// Configura as rotas
	http.HandleFunc("/upload", server.uploadHandler)
	http.HandleFunc("/file/", server.downloadHandler)

	// Adiciona uma rota de status para verificar se o servidor está funcionando
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		maxSizeMB := float64(config.MaxFileSize) / (1024 * 1024)
		fmt.Fprintf(w, `{"status": "ok", "service": "file-share-server", "max_file_size_mb": %.1f}`, maxSizeMB)
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
    <title>File Share Server</title>
    <meta charset="utf-8">
</head>
<body>
    <h1>🗂️ File Share Server</h1>
    <p>Servidor de compartilhamento temporário de arquivos</p>
    
    <h2>📤 Como fazer upload:</h2>
    <p>Envie uma requisição POST para <code>/upload</code> com o arquivo no campo <code>file</code></p>
    <p><strong>Tamanho máximo:</strong> %.1f MB</p>
    
    <h3>Exemplo com curl:</h3>
    <pre><code>curl -F "file=@meuarquivo.txt" %s/upload</code></pre>
    
    <h2>📥 Como fazer download:</h2>
    <p>Use o link retornado após o upload. O arquivo será removido após o primeiro download.</p>
    
    <h2>🔍 Status do servidor:</h2>
    <p><a href="/status">/status</a> - Verificar se o servidor está funcionando</p>
    
    <h2>⚙️ Configurações:</h2>
    <ul>
        <li><strong>Tamanho máximo:</strong> %.1f MB</li>
        <li><strong>Diretório:</strong> %s</li>
        <li><strong>Porta:</strong> %s</li>
    </ul>
</body>
</html>`, maxSizeMB, config.BaseURL, maxSizeMB, config.StoragePath, config.Port)
	})

	serverAddr := ":" + config.Port
	fmt.Printf("🚀 Servidor iniciado em %s\n", config.BaseURL)
	fmt.Printf("📁 Diretório de uploads: %s\n", config.StoragePath)
	fmt.Printf("📏 Tamanho máximo de arquivo: %.1f MB\n", float64(config.MaxFileSize)/(1024*1024))
	fmt.Printf("ℹ️  Acesse %s para ver as instruções\n", config.BaseURL)
	
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}
