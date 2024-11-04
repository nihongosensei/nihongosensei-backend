package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/lib/pq"
	"github.com/nihongosensei/nihongosensei-go/internal/handler"
	"github.com/nihongosensei/nihongosensei-go/internal/postgres"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

func getLogLevelFromEnv() slog.Level {
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

func handleRequests() {
	port := getEnv("PORT", "3001")
	slog.SetLogLoggerLevel(getLogLevelFromEnv())

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	connStr := "user=wadoku password=wadoku dbname=postgres sslmode=disable"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	goquDB := goqu.New("postgres", db)

	mux := http.NewServeMux()

	d := &postgres.PostgresDictionary{
		DB:   db,
		Goqu: goquDB,
	}

	mux.Handle("POST /api/furigana", &handler.TokenizeHandler{Tokenizer: t})

	mux.Handle("GET /api/entry/id/{id}", &handler.EntryByIdHandler{Dictionary: d})
	mux.Handle("GET /api/entry/id/{id}/resolved", &handler.ResolvedEntryHandler{Dictionary: d})
	mux.Handle("GET /api/entry/text/{text}", &handler.EntryByTextHandler{Dictionary: d})
	mux.Handle("POST /api/entries", &handler.EntriesByFilter{Dictionary: d})

	mux.Handle("GET /api/kanji/{kanji}", &handler.KanjiByCharHandler{Dictionary: d})

	slog.Info("Server is listening on port", "port", port)

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), mux)

	if errors.Is(err, http.ErrServerClosed) {
		slog.Info("server closed")
	} else if err != nil {
		slog.Error("error starting server", "error", err)
		os.Exit(1)
	}
}

func main() {
	handleRequests()
}
