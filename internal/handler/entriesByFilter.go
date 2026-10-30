package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

type EntriesByFilter struct {
	Dictionary dictionary.Dictionary
	Logger     *slog.Logger
}

type EntriesByFilterPostBody struct {
	Query  string   `json:"query"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
	Jlpt   int16    `json:"jlpt,omitempty"`
	Kanjis []string `json:"kanjis,omitempty"`
	Ids    []int    `json:"ids,omitempty"`
	Text   string   `json:"text,omitempty"`
	// Sort criteria in priority order.
	Sort []dictionary.EntrySort `json:"sort,omitempty"`
}

const maxSortCriteria = 3

// validateSort checks the constraints which cannot be expressed by the enums.
func validateSort(sort []dictionary.EntrySort, hasQuery bool) bool {
	if len(sort) > maxSortCriteria {
		return false
	}
	seen := map[dictionary.EntrySortField]bool{}
	for _, s := range sort {
		if seen[s.Field] || (s.Field == dictionary.EntrySortScore && !hasQuery) {
			return false
		}
		seen[s.Field] = true
	}
	return true
}

func (h *EntriesByFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input EntriesByFilterPostBody
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&input); err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	if !validateSort(input.Sort, len(input.Query) > 0) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	filter := dictionary.EntryFilter{Sort: input.Sort}
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

	entries, err := h.Dictionary.SelectEntries(r.Context(), filter)
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
