package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
	"github.com/nihongosensei/nihongosensei-go/internal/japanese"
	"log"
	"net/http"
	"strconv"
)

type ResolvedEntryHandler struct {
	Dictionary dictionary.Dictionary
}

type ResolvedEntry struct {
	Entry     dictionary.Entry                        `json:"entry"`
	Synonyms  []dictionary.Entry                      `json:"synonyms"`
	Relations []dictionary.EntryRefWithResolvedSource `json:"relatedEntries"`
	Kanjis    []dictionary.Kanji                      `json:"kanjis"`
}

func (h *ResolvedEntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var idString = r.PathValue("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	entry, err := h.Dictionary.SelectEntry(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	// Select synonyms
	synonymIds := mapset.NewSet[int]()
	for _, ruigo := range entry.EntryJson.Ruigos.Ruigo {
		synonymIds.Add(ruigo.Id)
	}

	synonyms, err := h.Dictionary.SelectEntryByIds(synonymIds)

	// Select Kanjis
	kanjiSet := mapset.NewSet[string]()
	for _, orth := range entry.EntryJson.Form.Orth {
		kanjiSet.Append(japanese.FindKanjis(orth.Value)...)
	}
	kanjis, err := h.Dictionary.SelectKanjis(kanjiSet)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	// Select References
	refs, err := h.Dictionary.SelectMainRefsByTargetId(id)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	// Prepare Output
	entryDto := ResolvedEntry{
		Entry:     entry,
		Synonyms:  synonyms,
		Relations: refs,
		Kanjis:    kanjis,
	}

	resp, err := json.Marshal(entryDto)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
