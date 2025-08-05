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

// Mapa de MIME types adicionais para melhor detecção
var customMimeTypes = map[string]string{
	".pdf":  "application/pdf",
	".doc":  "application/msword",
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xls":  "application/vnd.ms-excel",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".ppt":  "application/vnd.ms-powerpoint",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
	".zip":  "application/zip",
	".rar":  "application/x-rar-compressed",
	".7z":   "application/x-7z-compressed",
	".tar":  "application/x-tar",
	".gz":   "application/gzip",
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".mp4":  "video/mp4",
	".avi":  "video/x-msvideo",
	".mov":  "video/quicktime",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".svg":  "image/svg+xml",
	".webp": "image/webp",
	".txt":  "text/plain",
	".html": "text/html",
	".htm":  "text/html",
	".css":  "text/css",
	".js":   "application/javascript",
	".json": "application/json",
	".xml":  "application/xml",
	".csv":  "text/csv",
}

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
	mimeType := detectMimeType(originalName)
	
	// Garante que o arquivo tenha extensão correta
	finalFilename := ensureFileExtension(originalName, mimeType)

	// Obtém informações do arquivo
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Erro ao obter informações do arquivo.", http.StatusInternalServerError)
		return
	}

	// Define headers para download com MIME type correto
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", finalFilename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

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
	log.Printf("📥 [DOWNLOAD] Arquivo '%s' → '%s' (%s) baixado e removido - %s", originalName, finalFilename, mimeType, r.RemoteAddr)
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

// detectMimeType detecta o MIME type baseado na extensão do arquivo
func detectMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	// Primeiro verifica no mapa customizado
	if mimeType, exists := customMimeTypes[ext]; exists {
		return mimeType
	}
	
	// Depois tenta usar a biblioteca padrão do Go
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	
	// Fallback para binário genérico
	return "application/octet-stream"
}

// ensureFileExtension garante que o arquivo tenha extensão baseada no MIME type
func ensureFileExtension(filename, mimeType string) string {
	ext := filepath.Ext(filename)
	if ext != "" {
		return filename // Já tem extensão
	}
	
	// Mapa reverso para adicionar extensão baseada no MIME type
	mimeToExt := map[string]string{
		"application/pdf":       ".pdf",
		"image/jpeg":            ".jpg",
		"image/png":             ".png",
		"image/gif":             ".gif",
		"text/plain":            ".txt",
		"application/json":      ".json",
		"video/mp4":             ".mp4",
		"audio/mpeg":            ".mp3",
		"application/zip":       ".zip",
		"application/msword":    ".doc",
	}
	
	if extension, exists := mimeToExt[mimeType]; exists {
		return filename + extension
	}
	
	return filename // Retorna sem modificação se não encontrar
}
