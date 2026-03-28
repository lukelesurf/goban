package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lukelesurf/goban/internal/model"
	"github.com/lukelesurf/goban/internal/service"
)

type TaskHandler struct {
	tasks *service.TaskService
}

func NewTaskHandler(tasks *service.TaskService) *TaskHandler {
	return &TaskHandler{tasks: tasks}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	q := r.URL.Query()
	tasks, err := h.tasks.List(r.Context(), boardID, service.TaskFilter{
		ColumnID: q.Get("column_id"),
		Priority: q.Get("priority"),
		Label:    q.Get("label"),
		Search:   q.Get("search"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	task, err := h.tasks.Get(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	comments, err := h.tasks.ListComments(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"task":     task,
		"comments": comments,
	})
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	boardID := chi.URLParam(r, "id")
	var body struct {
		Title       string         `json:"title"`
		Description string         `json:"description"`
		Priority    model.Priority `json:"priority"`
		Labels      string         `json:"labels"`
		DueDate     *string        `json:"due_date"` // YYYY-MM-DD or omit
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	c := service.TaskCreate{
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
		Labels:      body.Labels,
	}
	if body.DueDate != nil && *body.DueDate != "" {
		t, err := time.Parse("2006-01-02", *body.DueDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid due_date, use YYYY-MM-DD")
			return
		}
		c.DueDate = &t
	}

	task, err := h.tasks.Create(r.Context(), boardID, c)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Title       *string         `json:"title"`
		Description *string         `json:"description"`
		Priority    *model.Priority `json:"priority"`
		Labels      *string         `json:"labels"`
		DueDate     *string         `json:"due_date"` // YYYY-MM-DD to set, "" to clear
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u := service.TaskUpdate{
		Title:       body.Title,
		Description: body.Description,
		Priority:    body.Priority,
		Labels:      body.Labels,
	}
	if body.DueDate != nil {
		if *body.DueDate == "" {
			u.ClearDueDate = true
		} else {
			t, err := time.Parse("2006-01-02", *body.DueDate)
			if err != nil {
				writeError(w, http.StatusBadRequest, "invalid due_date, use YYYY-MM-DD")
				return
			}
			u.DueDate = &t
		}
	}

	task, err := h.tasks.Update(r.Context(), id, u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "task not found")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Move(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		ColumnID string `json:"column_id"`
		Position int    `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.ColumnID == "" {
		writeError(w, http.StatusBadRequest, "column_id is required")
		return
	}

	task, err := h.tasks.Move(r.Context(), id, body.ColumnID, body.Position)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, task)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.tasks.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ListComments(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	comments, err := h.tasks.ListComments(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, comments)
}

func (h *TaskHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Author string `json:"author"`
		Body   string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Body == "" {
		writeError(w, http.StatusBadRequest, "body is required")
		return
	}

	comment, err := h.tasks.AddComment(r.Context(), id, body.Author, body.Body)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, comment)
}
