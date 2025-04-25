package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/nihongosensei/nihongosensei-backend/internal/elasticsearch"
	"github.com/nihongosensei/nihongosensei-backend/internal/environment"
	"github.com/nihongosensei/nihongosensei-backend/internal/handler"
	"github.com/nihongosensei/nihongosensei-backend/internal/telemetry"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func newHTTPHandler() http.Handler {

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	esClient, err := elasticsearch.CreateClientFromEnv()

	if err != nil {
		panic(err)
	}

	des := &elasticsearch.ElasticsearchDictionary{
		TypedClient: esClient,
	}

	// handle is a replacement for mux.Handle
	// which enriches the handler's HTTP instrumentation with the pattern as the http.route.
	handle := func(pattern string, handler http.Handler) {
		mux.Handle(pattern, otelhttp.NewHandler(telemetry.TracingMiddleware(handler), pattern))
	}

	handle("POST /api/furigana", &handler.TokenizeHandler{Tokenizer: t})

	handle("GET /api/entry/id/{id}", &handler.EntryByIdHandler{Dictionary: des})
	handle("GET /api/entry/id/{id}/resolved", &handler.ResolvedEntryHandler{Dictionary: des})
	handle("POST /api/entries", &handler.EntriesByFilter{Dictionary: des})

	handle("GET /api/kanji/{kanji}", &handler.KanjiByCharHandler{Dictionary: des})
	handle("POST /api/kanjis", &handler.KanjisByFilter{Dictionary: des})

	// We do not include the health endpoint in otel for now to avoid unnecessary spans
	mux.HandleFunc("GET /health", healthHandler)

	return mux
}

type HealthResponse struct {
	Status string `json:"status"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
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

	port := environment.Get("PORT", "3001")

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
	slog.SetLogLoggerLevel(environment.LogLevel())

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}
