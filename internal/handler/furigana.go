package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nihongosensei/nihongosensei-backend/internal/furigana"

	"github.com/ikawaha/kagome/v2/tokenizer"
)

type TokenizeHandler struct {
	Tokenizer *tokenizer.Tokenizer
	Logger    *slog.Logger
}

type PostBody struct {
	Text string `json:"text"`
}

func (h *TokenizeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input PostBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
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
		h.Logger.Error("error during TextToFurigana", "error", err)
		return
	}

	if err := json.NewEncoder(w).Encode(f); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("error while creating furigana response json", "error", err)
		return
	}
}
