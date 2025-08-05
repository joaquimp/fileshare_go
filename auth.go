package main

import (
	"crypto/subtle"
	"log"
	"net/http"
	"strings"
)

// AuthConfig armazena as configurações de autenticação
type AuthConfig struct {
	APIKey           string // Chave de API para autenticação
	AllowedUserAgent string // User-Agent permitido (opcional)
}

// NewAuthConfig cria uma nova configuração de autenticação
func NewAuthConfig(apiKey, userAgent string) *AuthConfig {
	return &AuthConfig{
		APIKey:           apiKey,
		AllowedUserAgent: userAgent,
	}
}

// ValidateRequest valida se a requisição vem do aplicativo autorizado
func (auth *AuthConfig) ValidateRequest(r *http.Request) bool {
	// Verifica API Key no header Authorization
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Printf("❌ [AUTH] Tentativa de acesso sem Authorization header - %s", r.RemoteAddr)
		return false
	}

	// Formato esperado: "Bearer YOUR_API_KEY"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Printf("❌ [AUTH] Formato Authorization inválido - %s", r.RemoteAddr)
		return false
	}

	// Comparação segura contra timing attacks
	if subtle.ConstantTimeCompare([]byte(parts[1]), []byte(auth.APIKey)) != 1 {
		log.Printf("❌ [AUTH] API Key inválida - %s", r.RemoteAddr)
		return false
	}

	// Verifica User-Agent se configurado
	if auth.AllowedUserAgent != "" {
		userAgent := r.Header.Get("User-Agent")
		if !strings.Contains(userAgent, auth.AllowedUserAgent) {
			log.Printf("❌ [AUTH] User-Agent não autorizado: '%s' - %s", userAgent, r.RemoteAddr)
			return false
		}
	}

	log.Printf("✅ [AUTH] Acesso autorizado - %s", r.RemoteAddr)
	return true
}

// RequireAuth é um middleware que protege endpoints
func (auth *AuthConfig) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.ValidateRequest(r) {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"Unauthorized","message":"API key inválida ou aplicativo não autorizado"}`, http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}
