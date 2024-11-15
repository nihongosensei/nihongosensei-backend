package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/doug-martin/goqu/v9"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	_ "github.com/lib/pq"
	"github.com/nihongosensei/nihongosensei-backend/internal/handler"
	"github.com/nihongosensei/nihongosensei-backend/internal/postgres"
	"github.com/nihongosensei/nihongosensei-backend/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

// Requires the given env key. If the key is empty, panic.
func requireEnv(key string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		panic(key + " is required")
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

func newHTTPHandler() http.Handler {

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	d := &postgres.PostgresDictionary{
		Goqu: createDatabase(),
	}

	// handle is a replacement for mux.Handle
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, otelhttp.NewHandler(telemetry.TracingMiddleware(handler), pattern))
	}

	handle("POST /api/furigana", &handler.TokenizeHandler{Tokenizer: t})

	handle("GET /api/entry/id/{id}", &handler.EntryByIdHandler{Dictionary: d})
	handle("GET /api/entry/id/{id}/resolved", &handler.ResolvedEntryHandler{Dictionary: d})
	handle("GET /api/entry/text/{text}", &handler.EntryByTextHandler{Dictionary: d})
	handle("POST /api/entries", &handler.EntriesByFilter{Dictionary: d})

	handle("GET /api/kanji/{kanji}", &handler.KanjiByCharHandler{Dictionary: d})

	return mux
}

func createDatabase() *goqu.Database {
	dbUser := requireEnv("DB_USER")
	dbPassword := requireEnv("DB_PASSWORD")
	dbDatabase := requireEnv("DB_DATABASE")

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbDatabase)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	return goqu.New("postgres", db)
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Set up OpenTelemetry.
	otelShutdown, err := telemetry.SetupOTelSDK(ctx)
	if err != nil {
		return
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = errors.Join(err, otelShutdown(context.Background()))
	}()

	port := getEnv("PORT", "3001")

	// Start HTTP server.
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		BaseContext:  func(_ net.Listener) context.Context { return ctx },
		ReadTimeout:  time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      newHTTPHandler(),
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	err = srv.Shutdown(context.Background())
	return
}

func main() {
	slog.SetLogLoggerLevel(getLogLevelFromEnv())

	go telemetry.ServeMetrics()

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
