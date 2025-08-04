package main

import (
	"log"
	"os"
	"strconv"
)

// Config armazena as configurações da aplicação
type Config struct {
	Port        string // Porta do servidor
	BaseURL     string // URL base para geração de links
	StoragePath string // Diretório de armazenamento
	MaxFileSize int64  // Tamanho máximo de arquivo em bytes
}

// LoadConfig carrega as configurações das variáveis de ambiente com valores padrão
func LoadConfig() *Config {
	config := &Config{
		Port:        getEnv("PORT", "8080"),
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		StoragePath: getEnv("STORAGE_PATH", "./uploads"),
		MaxFileSize: getEnvAsInt64("MAX_FILE_SIZE_MB", 5) * 1024 * 1024, // Converte MB para bytes
	}

	// Log das configurações carregadas
	log.Printf("📋 Configurações carregadas:")
	log.Printf("   Porta: %s", config.Port)
	log.Printf("   URL Base: %s", config.BaseURL)
	log.Printf("   Diretório: %s", config.StoragePath)
	log.Printf("   Tamanho máximo: %.1f MB", float64(config.MaxFileSize)/(1024*1024))

	return config
}

// getEnv obtém uma variável de ambiente ou retorna um valor padrão
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt64 obtém uma variável de ambiente como int64 ou retorna um valor padrão
func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
		log.Printf("⚠️  Valor inválido para %s: %s. Usando valor padrão: %d", key, value, defaultValue)
	}
	return defaultValue
}
