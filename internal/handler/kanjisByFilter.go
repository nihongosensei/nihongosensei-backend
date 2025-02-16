package handler

import (
	"encoding/json"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
	"log/slog"
	"net/http"
)

type KanjisByFilter struct {
	Dictionary dictionary.Dictionary
}

type KanjisByFilterPostBody struct {
	Limit   int      `json:"limit,omitempty"`
	Offset  int      `json:"offset,omitempty"`
	Strokes []int16  `json:"strokes,omitempty"`
	Grade   []int16  `json:"grade,omitempty"`
	Radical []string `json:"radical,omitempty"`
	Kanjis  []string `json:"kanjis,omitempty"`
}

func (h *KanjisByFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "kanjisByFilter")
	defer span.End()

	var input KanjisByFilterPostBody
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	filter := dictionary.KanjiFilters{}
	if input.Offset > 0 {
		filter.Offset = input.Offset
	} else {
		filter.Offset = 0
	}
	if input.Limit > 0 && input.Limit <= 50 {
		filter.Limit = input.Limit
	} else {
		filter.Limit = 10
	}
	if len(input.Kanjis) > 0 {
		filter.Kanjis = input.Kanjis
	}
	if len(input.Radical) > 0 {
		filter.Radical = input.Radical
	}
	if len(input.Grade) > 0 {
		filter.Grade = input.Grade
	}
	if len(input.Strokes) > 0 {
		filter.Strokes = input.Strokes
	}

	entries, err := h.Dictionary.SelectKanjis(filter)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error during SelectEntries", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(entries)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error while creating entriesByFilter response json", "error", err)
		return
	}
}
