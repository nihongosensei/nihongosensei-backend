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
	Query  string `json:"query"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
	// JLPT levels (1-5). Entries matching any of them are returned.
	Jlpt   []int16  `json:"jlpt,omitempty"`
	Kanjis []string `json:"kanjis,omitempty"`
	Ids    []int    `json:"ids,omitempty"`
	Text   string   `json:"text,omitempty"`
	// Grammar types. Entries matching any of them are returned.
	GrammarType []string `json:"grammar_type,omitempty"`
	// Usage labels. Entries matching any of them are returned.
	Usg []string `json:"usg,omitempty"`
	// Aggregations to compute, see dictionary.EntryFilter.
	Aggregations []dictionary.EntryAggregationField `json:"aggregations,omitempty"`
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

// validateAggregations rejects duplicate fields.
func validateAggregations(aggs []dictionary.EntryAggregationField) bool {
	seen := map[dictionary.EntryAggregationField]bool{}
	for _, a := range aggs {
		if seen[a] {
			return false
		}
		seen[a] = true
	}
	return true
}

// validateJlpt checks that all levels are within 1-5.
func validateJlpt(levels []int16) bool {
	for _, l := range levels {
		if l < 1 || l > 5 {
			return false
		}
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

	if !validateSort(input.Sort, len(input.Query) > 0) ||
		!validateAggregations(input.Aggregations) ||
		!validateJlpt(input.Jlpt) {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	filter := dictionary.EntryFilter{
		Sort:         input.Sort,
		Jlpt:         input.Jlpt,
		GrammarType:  input.GrammarType,
		Usg:          input.Usg,
		Aggregations: input.Aggregations,
	}
	if len(input.Query) > 0 {
		filter.Query = input.Query
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
