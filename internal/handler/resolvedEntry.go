package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
	"log/slog"
	"net/http"
	"strconv"
)

type ResolvedEntryHandler struct {
	Dictionary dictionary.Dictionary
}

func (h *ResolvedEntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "resolvedEntry")
	defer span.End()

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
		slog.Error("error during SelectEntry", "error", err)
		return
	}

	// Select synonyms
	var synonyms *[]dictionary.Entry
	if entry.Synonyms != nil && len(entry.Synonyms) > 0 {
		synonyms, err = h.Dictionary.SelectEntryByIds(entry.Synonyms)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			slog.Error("failed to select synonyms", "error", err)
			return
		}
	}

	// Select Kanjis
	var kanjis *[]dictionary.Kanji
	if entry.Kanjis != nil && len(entry.Kanjis) > 0 {
		kanjis, err = h.Dictionary.SelectKanjis(entry.Kanjis)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			slog.Error("failed to select kanjis", "error", err)
			return
		}
	}

	// Select References
	refs, err := h.Dictionary.SelectMainRefsByTargetId(id)

	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("failed to select main refs by target id", "error", err)
		return
	}

	// Prepare Output
	entryDto := dictionary.ResolvedEntry{
		Entry:     entry,
		Synonyms:  synonyms,
		Relations: refs,
		Kanjis:    kanjis,
	}

	resp, err := json.Marshal(entryDto)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("failed to create response json for resolved entry", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
