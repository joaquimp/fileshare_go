package main

import (
	"crypto/rand"
	"encoding/hex"
)

// generateSecureToken gera um token criptograficamente seguro
// n é o número de bytes aleatórios (o token final será n*2 caracteres em hex)
func generateSecureToken(n int) (string, error) {
	bytes := make([]byte, n)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
