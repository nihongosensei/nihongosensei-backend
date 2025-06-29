package environment

import (
	"os"
)

func Get(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

// Require the given env key. If the key is empty, panic.
func Require(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(key + " is required")
	}

	return value
}
