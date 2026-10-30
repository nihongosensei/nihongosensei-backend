package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/nihongosensei/nihongosensei-backend/internal/elasticsearch"
	"github.com/nihongosensei/nihongosensei-backend/internal/environment"
	"github.com/nihongosensei/nihongosensei-backend/internal/handler"
	"github.com/nihongosensei/nihongosensei-backend/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func newHTTPHandler(logger *slog.Logger) (http.Handler, error) {

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		return nil, fmt.Errorf("failed to create tokenizer: %w", err)
	}

	mux := http.NewServeMux()

	esClient, err := elasticsearch.CreateClientFromEnv()

	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	des := &elasticsearch.ElasticsearchDictionary{
		TypedClient: esClient,
		Logger:      logger,
	}

	// handle is a replacement for mux.Handle
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, otelhttp.NewHandler(telemetry.TracingMiddleware(handler), pattern))
	}

	handle("POST /api/furigana", &handler.TokenizeHandler{Tokenizer: t, Logger: logger})

	handle("GET /api/entry/id/{id}", &handler.EntryByIdHandler{Dictionary: des, Logger: logger})
	handle("GET /api/entry/id/{id}/resolved", &handler.ResolvedEntryHandler{Dictionary: des, Logger: logger})
	handle("POST /api/entries", &handler.EntriesByFilter{Dictionary: des, Logger: logger})

	handle("GET /api/kanji/{kanji}", &handler.KanjiByCharHandler{Dictionary: des, Logger: logger})
	handle("POST /api/kanjis", &handler.KanjisByFilter{Dictionary: des, Logger: logger})

	// We do not include the health endpoint in otel for now to avoid unnecessary spans
	mux.HandleFunc("GET /health", healthHandler)

	return mux, nil
}

type HealthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}

func run() (err error) {
	// Handle SIGINT (CTRL+C) and SIGTERM (sent by docker/kubernetes) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
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

	port := environment.Get("PORT", "3001")

	// create logger and pass down
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{}))

	handler, err := newHTTPHandler(logger)
	if err != nil {
		return err
	}

	// Start HTTP server.
	srv := &http.Server{
		Addr:        fmt.Sprintf(":%s", port),
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		// Default read timeout is short; increase to be more permissive for real clients.
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      handler,
	}
	srvErr := make(chan error, 1)
	go func() {
		srvErr <- srv.ListenAndServe()
	}()

	// Wait for interruption.
	select {
	case err = <-srvErr:
		// If the server returned ErrServerClosed as part of a normal shutdown, treat that as nil.
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		// Error when starting HTTP server.
		return
	case <-ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		stop()
	}

	// When Shutdown is called, ListenAndServe immediately returns ErrServerClosed.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
		// if shutdown fails, return that error
		err = errors.Join(err, shutdownErr)
	}
	return
}

func main() {
	if err := run(); err != nil {
		// Write to stderr directly: telemetry replaces the default slog handler, which
		// also receives output of the log package, so the error would not be visible.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
