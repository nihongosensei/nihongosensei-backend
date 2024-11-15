package handler

import (
	"encoding/json"
	"github.com/nihongosensei/nihongosensei-backend/internal/furigana"
	"log/slog"
	"net/http"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

type TokenizeHandler struct {
	Tokenizer *tokenizer.Tokenizer
}

type PostBody struct {
	Text string `json:"text"`
}

func (h *TokenizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "furigana")
	defer span.End()

	var input PostBody
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if input.Text == "" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
		return
	}

	f, err := furigana.TextToFurigana(h.Tokenizer, input.Text)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error during TextToFurigana", "error", err)
		return
	}

	err = json.NewEncoder(w).Encode(f)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error while creating furigana response json", "error", err)
		return
	}
}
