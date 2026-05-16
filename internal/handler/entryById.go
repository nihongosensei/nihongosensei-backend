package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nihongosensei/nihongosensei-backend/internal/dictionary"
)

type EntryByIdHandler struct {
	Dictionary dictionary.Dictionary
	Logger     *slog.Logger
}

func (h *EntryByIdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "entryById")
	defer span.End()

	var idString = r.PathValue("id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
		return
	}

	entry, err := h.Dictionary.SelectEntry(r.Context(), id)
	if err != nil {
		if errors.Is(err, dictionary.ErrEntryNotFound) {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("error during SelectEntry", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(entry); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		h.Logger.Error("error while creating entryById response json", "error", err)
		return
	}
}
