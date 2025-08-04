package main

import "sync"

// FileStorage gerencia o armazenamento de arquivos temporários
type FileStorage struct {
	tokenMap     map[string]string // Mapeia tokens para caminhos de arquivo
	mutex        *sync.Mutex       // Protege acesso concorrente ao mapa
	storagePath  string            // Diretório onde os arquivos são armazenados
	maxFileSize  int64             // Tamanho máximo de arquivo em bytes
}

// NewFileStorage cria uma nova instância do gerenciador de arquivos
func NewFileStorage(storagePath string, maxFileSize int64) *FileStorage {
	return &FileStorage{
		tokenMap:    make(map[string]string),
		mutex:       &sync.Mutex{},
		storagePath: storagePath,
		maxFileSize: maxFileSize,
	}
}

// AddFile adiciona um arquivo ao storage com token associado
func (fs *FileStorage) AddFile(token, filePath string) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	fs.tokenMap[token] = filePath
}

// GetAndRemoveFile obtém o caminho do arquivo pelo token e o remove do mapa
// Retorna o caminho do arquivo e um booleano indicando se foi encontrado
func (fs *FileStorage) GetAndRemoveFile(token string) (string, bool) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()
	
	filePath, exists := fs.tokenMap[token]
	if exists {
		delete(fs.tokenMap, token)
	}
	return filePath, exists
}

// GetMaxFileSize retorna o tamanho máximo permitido para arquivos
func (fs *FileStorage) GetMaxFileSize() int64 {
	return fs.maxFileSize
}
