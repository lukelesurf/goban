package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lukelesurf/goban/internal/service"
)

type ColumnHandler struct {
	columns *service.ColumnService
}

func NewColumnHandler(columns *service.ColumnService) *ColumnHandler {
	return &ColumnHandler{columns: columns}
}

func (h *ColumnHandler) List(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	cols, err := h.columns.List(r.Context(), boardID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cols)
}

func (h *ColumnHandler) Create(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	var body struct {
		Name     string `json:"name"`
		Position int    `json:"position"`
		WIPLimit int    `json:"wip_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	col, err := h.columns.Create(r.Context(), boardID, body.Name, body.Position, body.WIPLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, col)
}

func (h *ColumnHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name     string `json:"name"`
		Position int    `json:"position"`
		WIPLimit int    `json:"wip_limit"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	col, err := h.columns.Update(r.Context(), id, body.Name, body.Position, body.WIPLimit)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "column not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, col)
}

func (h *ColumnHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.columns.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
