package handler

import (
	"encoding/json"
	"errors"
	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
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

	entry, err := h.Dictionary.SelectResolvedEntry(id)
	if err != nil {
		if errors.Is(err, dictionary.ErrEntryNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("error during SelectResolvedEntry", "error", err)
		return
	}

	resp, err := json.Marshal(entry)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		slog.Error("failed to create response json for resolved entry", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
