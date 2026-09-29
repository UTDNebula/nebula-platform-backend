package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Initialize this file to load environment variables from .env file
func init() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}

	for {
		envPath := filepath.Join(dir, ".env")
		if _, err := os.Stat(envPath); err == nil {
			_ = godotenv.Load(envPath)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}

func GetPortString() string {
	portNumber, exist := os.LookupEnv("PORT")
	if !exist || strings.TrimSpace(portNumber) == "" {
		portNumber = "8081"
	}

	port := fmt.Sprintf(":%s", portNumber)
	return port
}

func GetEnvMongoURI() (string, error) {
	uri, exist := os.LookupEnv("MONGODB_URI")
	if !exist {
		return "", fmt.Errorf("Error loading 'MONGODB_URI' from the .env file")
	}

	return uri, nil
}
