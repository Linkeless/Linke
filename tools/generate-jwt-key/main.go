package main

import (
	"fmt"
	"os"

	"linke/internal/shared/config"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--help" {
		fmt.Println("JWT Key Generator")
		fmt.Println("Usage: go run tools/generate-jwt-key/main.go")
		fmt.Println("")
		fmt.Println("Generates a cryptographically secure JWT key suitable for production use.")
		fmt.Println("The generated key will be 64 characters long (256 bits of entropy).")
		fmt.Println("")
		fmt.Println("Usage in environment:")
		fmt.Println("  export JWT_SECRET=$(go run tools/generate-jwt-key/main.go)")
		return
	}

	key := config.GenerateSecureJWTKey()
	fmt.Print(key)
}
