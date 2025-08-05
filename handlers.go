package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Server representa o servidor de compartilhamento de arquivos
type Server struct {
	storage *FileStorage
	baseURL string
}

// NewServer cria uma nova instância do servidor
func NewServer(storage *FileStorage, baseURL string) *Server {
	return &Server{
		storage: storage,
		baseURL: baseURL,
	}
}

// uploadHandler processa o upload de arquivos
func (s *Server) uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica se o método é POST
	if r.Method != http.MethodPost {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Method not allowed","message":"Use POST para upload"}`, http.StatusMethodNotAllowed)
		return
	}

	// Limita o tamanho do upload usando a configuração
	maxSize := s.storage.GetMaxFileSize()
	err := r.ParseMultipartForm(maxSize)
	if err != nil {
		maxSizeMB := float64(maxSize) / (1024 * 1024)
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"error":   "File too large",
			"message": fmt.Sprintf("Arquivo muito grande (máximo %.1f MB)", maxSizeMB),
			"max_size_mb": maxSizeMB,
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Obtém o arquivo do formulário
	file, header, err := r.FormFile("file")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"No file provided","message":"Certifique-se de que o campo 'file' está presente"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Valida o nome do arquivo
	if header.Filename == "" {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Invalid filename","message":"Nome do arquivo não pode estar vazio"}`, http.StatusBadRequest)
		return
	}

	// Gera um token seguro para o arquivo
	token, err := generateSecureToken(8) // 16 caracteres hex
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Internal server error","message":"Erro ao gerar token"}`, http.StatusInternalServerError)
		return
	}

	// Cria o caminho do arquivo com o token como prefixo
	safeFilename := sanitizeFilename(header.Filename)
	filePath := filepath.Join(s.storage.storagePath, token+"_"+safeFilename)

	// Cria o arquivo no sistema de arquivos
	out, err := os.Create(filePath)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Storage error","message":"Erro ao salvar arquivo no servidor"}`, http.StatusInternalServerError)
		return
	}
	defer out.Close()

	// Copia o conteúdo do arquivo enviado para o arquivo local
	_, err = io.Copy(out, file)
	if err != nil {
		os.Remove(filePath) // Remove arquivo parcialmente salvo
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"Storage error","message":"Erro ao salvar arquivo"}`, http.StatusInternalServerError)
		return
	}

	// Obtém informações do arquivo
	fileInfo, _ := out.Stat()
	fileSize := fileInfo.Size()

	// Adiciona o arquivo ao storage
	s.storage.AddFile(token, filePath)

	// Log da operação de upload
	log.Printf("📤 [UPLOAD] Arquivo '%s' enviado com token %s - %s", safeFilename, token, r.RemoteAddr)

	// Retorna JSON com informações do upload
	publicURL := fmt.Sprintf("%s/file/%s", s.baseURL, token)
	w.Header().Set("Content-Type", "application/json")
	
	response := map[string]interface{}{
		"success":     true,
		"message":     "Arquivo enviado com sucesso",
		"token":       token,
		"filename":    header.Filename,
		"safe_filename": safeFilename,
		"size_bytes":  fileSize,
		"size_mb":     float64(fileSize) / (1024 * 1024),
		"download_url": publicURL,
		"uploaded_at": time.Now().Format(time.RFC3339),
		"note":        "O arquivo será removido após o primeiro download",
	}
	
	json.NewEncoder(w).Encode(response)
}

// downloadHandler processa o download de arquivos
func (s *Server) downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Verifica se o método é GET
	if r.Method != http.MethodGet {
		http.Error(w, "Método não permitido. Use GET para download.", http.StatusMethodNotAllowed)
		return
	}

	// Extrai o token da URL
	token := filepath.Base(r.URL.Path)
	if token == "" || token == "." {
		http.Error(w, "Token inválido.", http.StatusBadRequest)
		return
	}

	// Busca e remove o arquivo do storage (uso único)
	filePath, exists := s.storage.GetAndRemoveFile(token)
	if !exists {
		http.Error(w, "Arquivo não encontrado ou já foi baixado.", http.StatusNotFound)
		return
	}

	// Abre o arquivo para leitura
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Erro ao abrir o arquivo no servidor.", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Extrai o nome original do arquivo (remove o prefixo do token)
	_, fileName := filepath.Split(filePath)
	originalName := extractOriginalFilename(fileName, token)

	// Detecta o MIME type baseado na extensão do arquivo
	mimeType := mime.TypeByExtension(filepath.Ext(originalName))
	if mimeType == "" {
		// Se não conseguir detectar, usa application/octet-stream como fallback
		mimeType = "application/octet-stream"
	}

	// Obtém informações do arquivo
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Erro ao obter informações do arquivo.", http.StatusInternalServerError)
		return
	}

	// Define headers para download com MIME type correto
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", originalName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	// Envia o arquivo para o cliente
	_, err = io.Copy(w, file)
	if err != nil {
		// Se houve erro durante o envio, tenta remover o arquivo
		os.Remove(filePath)
		return
	}

	// Remove o arquivo após download bem-sucedido
	os.Remove(filePath)

	// Log da operação de download
	log.Printf("📥 [DOWNLOAD] Arquivo '%s' (%s) baixado e removido - %s", originalName, mimeType, r.RemoteAddr)
}

// sanitizeFilename remove caracteres perigosos do nome do arquivo
func sanitizeFilename(filename string) string {
	// Remove caracteres que podem ser problemáticos
	dangerous := []string{"/", "\\", "..", ":", "*", "?", "\"", "<", ">", "|"}
	safe := filename
	for _, char := range dangerous {
		safe = strings.ReplaceAll(safe, char, "_")
	}
	return safe
}

// extractOriginalFilename extrai o nome original do arquivo removendo o prefixo do token
func extractOriginalFilename(fileName, token string) string {
	prefix := token + "_"
	if strings.HasPrefix(fileName, prefix) {
		return fileName[len(prefix):]
	}
	return fileName
}
