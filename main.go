package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	storagePath = "./uploads"
	tokenMap    = make(map[string]string) // token -> filepath
	mutex       = &sync.Mutex{}
)

func generateToken(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	rand.Seed(time.Now().UnixNano())
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		http.Error(w, "Erro ao processar o formulário", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Erro ao obter o arquivo", http.StatusBadRequest)
		return
	}
	defer file.Close()

	token := generateToken(12)
	filePath := filepath.Join(storagePath, token+"_"+header.Filename)

	out, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Erro ao salvar o arquivo", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	mutex.Lock()
	tokenMap[token] = filePath
	mutex.Unlock()

	publicURL := fmt.Sprintf("http://localhost:8080/file/%s", token)
	fmt.Fprintf(w, "Arquivo disponível para download: %s\n", publicURL)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	token := filepath.Base(r.URL.Path)

	mutex.Lock()
	filePath, exists := tokenMap[token]
	if exists {
		delete(tokenMap, token)
	}
	mutex.Unlock()

	if !exists {
		http.Error(w, "Arquivo não encontrado ou já baixado", http.StatusNotFound)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Erro ao abrir o arquivo", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Nome original do arquivo após "_"
	_, fileName := filepath.Split(filePath)
	split := filepath.Base(fileName)
	if i := len(token) + 1; len(split) > i {
		w.Header().Set("Content-Disposition", "attachment; filename="+split[i:])
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	io.Copy(w, file)

	// Apagar o arquivo
	os.Remove(filePath)
}

func main() {
	os.MkdirAll(storagePath, 0755)

	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/file/", downloadHandler)

	fmt.Println("Servidor iniciado em http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
