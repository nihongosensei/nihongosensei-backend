package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

type ResolvedEntryHandler struct {
	Dictionary dictionary.Dictionary
	Logger     *slog.Logger
}

func (h *ResolvedEntryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var idString = r.PathValue("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	entry, err := h.Dictionary.SelectResolvedEntry(r.Context(), id)
	if err != nil {
		if errors.Is(err, dictionary.ErrEntryNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("error during SelectResolvedEntry", "error", err)
		return
	}

	resp, err := json.Marshal(entry)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("failed to create response json for resolved entry", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
