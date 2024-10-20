package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/nihongosensei/nihongosensei-go/internal/dictionary"
	"log"
	"net/http"
	"strconv"
)

type EntryByIdHandler struct {
	Dictionary dictionary.Dictionary
}

func (h *EntryByIdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(entry)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("error while creating response json: %v\n", err)
		return
	}
}
