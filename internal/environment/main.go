package environment

import (
	"log/slog"
	"os"
	"strings"
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

func LogLevel() slog.Level {
	logLevelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	switch logLevelStr {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
