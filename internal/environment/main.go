package environment

import (
	"fmt"
	"os"
)

func Get(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

// Require returns the value of the given env key. If the key is empty, an error is returned.
func Require(key string) (string, error) {
	value := os.Getenv(key)
	if len(value) == 0 {
		return "", fmt.Errorf("environment variable %s is required", key)
	}

	return value, nil
}
