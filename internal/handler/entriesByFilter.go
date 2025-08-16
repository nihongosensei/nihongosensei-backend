package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

type EntriesByFilter struct {
	Dictionary dictionary.Dictionary
}

type EntriesByFilterPostBody struct {
	Query  string   `json:"query"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
	Jlpt   int16    `json:"jlpt,omitempty"`
	Kanjis []string `json:"kanjis,omitempty"`
	Ids    []int    `json:"ids,omitempty"`
	Text   string   `json:"text,omitempty"`
}

func (h *EntriesByFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "entriesByFilter")
	defer span.End()

	var input EntriesByFilterPostBody
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	filter := dictionary.EntryFilter{}
	if len(input.Query) > 0 {
		filter.Query = input.Query
	}
	if input.Jlpt > 0 && input.Jlpt <= 5 {
		filter.Jlpt = input.Jlpt
	}
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
	if len(input.Ids) > 0 {
		filter.Ids = input.Ids
	}
	if len(input.Text) > 0 {
		filter.Text = input.Text
	}

	entries, err := h.Dictionary.SelectEntries(filter)
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
