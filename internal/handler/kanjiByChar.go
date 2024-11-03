package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
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

// Checks if the given kanji is one Kanji
func containsExactlyOneKanji(s string) bool {
	kanjiCount := 0

	for _, r := range s {
		if unicode.In(r, unicode.Han) {
			kanjiCount++
		}
		if kanjiCount > 1 {
			return false
		}
	}

	return kanjiCount == 1
}

func (h *KanjiByCharHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var kanjiChar = r.PathValue("kanji")
	if !containsExactlyOneKanji(kanjiChar) {
		http.Error(w, "kanji must contain exactly one kanji", http.StatusBadRequest)
		return
	}

	kanji, err := h.Dictionary.SelectKanji(kanjiChar)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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
