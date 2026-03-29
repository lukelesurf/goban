package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/lukelesurf/goban/internal/service"
)

type AttachmentHandler struct {
	attachments *service.AttachmentService
}

func NewAttachmentHandler(attachments *service.AttachmentService) *AttachmentHandler {
	return &AttachmentHandler{attachments: attachments}
}

func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")
	attachments, err := h.attachments.List(r.Context(), taskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, attachments)
}

func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	taskID := chi.URLParam(r, "id")

	r.Body = http.MaxBytesReader(w, r.Body, 25<<20) // 25 MB
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large (max 25 MB)")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	att, err := h.attachments.Upload(r.Context(), taskID, header.Filename, file)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, att)
}

func (h *AttachmentHandler) Download(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	att, err := h.attachments.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "attachment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.Filename))
	w.Header().Set("Content-Type", att.ContentType)
	http.ServeFile(w, r, h.attachments.FilePath(att))
}

func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.attachments.Delete(r.Context(), id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "attachment not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
