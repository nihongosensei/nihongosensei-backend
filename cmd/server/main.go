package main

import (
	"errors"
	"fmt"
	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
	"github.com/nihongosensei/nihongosensei-go/internal/furigana"
	"log"
	"net/http"
	"os"
)

func getEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}

	return value
}

func handleRequests() {
	port := getEnv("PORT", "3001")

	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()

	mux.Handle("POST /furigana", &furigana.TokenizeHandler{Tokenizer: t})

	log.Printf("Server is listening on port %s\n", port)

	err = http.ListenAndServe(fmt.Sprintf(":%s", port), mux)

	if errors.Is(err, http.ErrServerClosed) {
		log.Println("server closed")
	} else if err != nil {
		log.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}

func main() {
	handleRequests()
}
