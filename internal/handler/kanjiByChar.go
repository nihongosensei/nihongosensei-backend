package handler

import (
	"encoding/json"
	"errors"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
	"log/slog"
	"net/http"
	"unicode"
)

type KanjiByCharHandler struct {
	Dictionary dictionary.Dictionary
}

type KanjiByCharDTO struct {
	Kanji   string `json:"kanji"`
	Strokes int16  `json:"strokes"`
	Grade   int16  `json:"grade"`
	Kunyomi string `json:"kunyomi,omitempty"`
	Onyomi  string `json:"onyomi,omitempty"`
	Radical string `json:"radical,omitempty"`
}

// ContainsExactlyOneKanji checks if the given text contains exactly one kanji
func ContainsExactlyOneKanji(s string) bool {
	kanjiCount := 0
	for _, r := range s {
		if unicode.In(r, unicode.Han) {
			kanjiCount++
		} else {
			return false
		}
		if kanjiCount > 1 {
			return false
		}
	}

	return kanjiCount == 1
}

func (h *KanjiByCharHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "kanjiByChar")
	defer span.End()

	var kanjiChar = r.PathValue("kanji")
	if !ContainsExactlyOneKanji(kanjiChar) {
		http.Error(w, "kanji must contain exactly one kanji", http.StatusBadRequest)
		return
	}

	kanji, err := h.Dictionary.SelectKanji(kanjiChar)
	if err != nil {
		if errors.Is(err, dictionary.ErrKanjiNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error during SelectKanji", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(kanji)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error while creating kanjiByChar response json", "error", err)
		return
	}
}
