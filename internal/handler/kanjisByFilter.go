package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

type KanjisByFilter struct {
	Dictionary dictionary.Dictionary
	Logger     *slog.Logger
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
	var input KanjisByFilterPostBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
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

	entries, err := h.Dictionary.SelectKanjis(r.Context(), filter)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("error during SelectEntries", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(entries); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("error while creating entriesByFilter response json", "error", err)
		return
	}
}
