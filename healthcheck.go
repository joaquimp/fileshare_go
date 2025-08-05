// +build ignore

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// healthcheck realiza uma verificação simples de saúde do servidor
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	client := http.Client{
		Timeout: time.Second * 5,
	}

	resp, err := client.Get(fmt.Sprintf("http://localhost:%s/status", port))
	if err != nil {
		fmt.Printf("Health check failed: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Health check failed: status %d\n", resp.StatusCode)
		os.Exit(1)
	}

	fmt.Println("Health check passed")
	os.Exit(0)
}
