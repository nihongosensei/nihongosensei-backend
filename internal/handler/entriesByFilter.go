package handler

import (
	"encoding/json"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
	"log"
	"net/http"
)

type EntriesByFilter struct {
	Dictionary dictionary.Dictionary
}

type EntriesByFilterPostBody struct {
	Query  string `json:"query"`
	Limit  uint   `json:"limit,omitempty"`
	Offset uint   `json:"offset,omitempty"`
	Jlpt   int16  `json:"jlpt,omitempty"`
}

type EntriesByFilterResponse struct {
	Entries []dictionary.Entry `json:"entries"`
	Count   int                `json:"count"`
}

func (h *EntriesByFilter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var input EntriesByFilterPostBody
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		log.Printf("error while decoding input json: %v\n", err)
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
	if input.Limit > 0 && input.Limit < 50 {
		filter.Limit = input.Limit
	} else {
		filter.Limit = 10
	}

	entries, err := h.Dictionary.SelectEntries(filter)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	entriesCount, err := h.Dictionary.SelectEntriesCount(filter)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(&EntriesByFilterResponse{
		Entries: entries,
		Count:   entriesCount,
	})
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}
}
