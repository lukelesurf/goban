package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/model"
)

type TaskService struct {
	db *db.DB
}

func NewTaskService(db *db.DB) *TaskService {
	return &TaskService{db: db}
}

type TaskFilter struct {
	ColumnID string
	Priority string
	Label    string
	Search   string
}

func (s *TaskService) List(ctx context.Context, boardID string, f TaskFilter) ([]model.Task, error) {
	q := `SELECT id, board_id, column_id, title, description, priority, labels, position, created_at, updated_at, completed_at
		  FROM tasks WHERE board_id = ?`
	args := []any{boardID}

	if f.ColumnID != "" {
		q += " AND column_id = ?"
		args = append(args, f.ColumnID)
	}
	if f.Priority != "" {
		q += " AND priority = ?"
		args = append(args, f.Priority)
	}
	if f.Label != "" {
		q += " AND labels LIKE ?"
		args = append(args, "%"+f.Label+"%")
	}
	if f.Search != "" {
		q += " AND (title LIKE ? OR description LIKE ?)"
		args = append(args, "%"+f.Search+"%", "%"+f.Search+"%")
	}

	q += " ORDER BY position ASC"

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.BoardID, &t.ColumnID, &t.Title, &t.Description,
			&t.Priority, &t.Labels, &t.Position, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *TaskService) Get(ctx context.Context, id string) (*model.Task, error) {
	var t model.Task
	err := s.db.QueryRowContext(ctx,
		`SELECT id, board_id, column_id, title, description, priority, labels, position, created_at, updated_at, completed_at
		 FROM tasks WHERE id = ?`, id,
	).Scan(&t.ID, &t.BoardID, &t.ColumnID, &t.Title, &t.Description,
		&t.Priority, &t.Labels, &t.Position, &t.CreatedAt, &t.UpdatedAt, &t.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("get task %s: %w", id, err)
	}
	return &t, nil
}

func (s *TaskService) Create(ctx context.Context, boardID, title, description string, priority model.Priority, labels string) (*model.Task, error) {
	if priority == "" {
		priority = model.PriorityMedium
	}

	var columnID string
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM columns WHERE board_id = ? ORDER BY position ASC LIMIT 1`, boardID,
	).Scan(&columnID)
	if err != nil {
		return nil, fmt.Errorf("find first column: %w", err)
	}

	var maxPos int
	s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(position), -1) FROM tasks WHERE column_id = ?`, columnID,
	).Scan(&maxPos)

	now := time.Now().UTC()
	t := model.Task{
		ID:          uuid.NewString(),
		BoardID:     boardID,
		ColumnID:    columnID,
		Title:       title,
		Description: description,
		Priority:    priority,
		Labels:      labels,
		Position:    maxPos + 1,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO tasks (id, board_id, column_id, title, description, priority, labels, position, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID, t.BoardID, t.ColumnID, t.Title, t.Description, t.Priority, t.Labels, t.Position, t.CreatedAt, t.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	return &t, nil
}

type TaskUpdate struct {
	Title       *string
	Description *string
	Priority    *model.Priority
	Labels      *string
}

func (s *TaskService) Update(ctx context.Context, id string, u TaskUpdate) (*model.Task, error) {
	setClauses := []string{"updated_at = ?"}
	args := []any{time.Now().UTC()}

	if u.Title != nil {
		setClauses = append(setClauses, "title = ?")
		args = append(args, *u.Title)
	}
	if u.Description != nil {
		setClauses = append(setClauses, "description = ?")
		args = append(args, *u.Description)
	}
	if u.Priority != nil {
		setClauses = append(setClauses, "priority = ?")
		args = append(args, *u.Priority)
	}
	if u.Labels != nil {
		setClauses = append(setClauses, "labels = ?")
		args = append(args, *u.Labels)
	}

	args = append(args, id)
	_, err := s.db.ExecContext(ctx,
		`UPDATE tasks SET `+strings.Join(setClauses, ", ")+` WHERE id = ?`,
		args...,
	)
	if err != nil {
		return nil, fmt.Errorf("update task %s: %w", id, err)
	}
	return s.Get(ctx, id)
}

func (s *TaskService) Move(ctx context.Context, taskID, columnID string, position int) (*model.Task, error) {
	var lastColumnID string
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM columns WHERE board_id = (SELECT board_id FROM tasks WHERE id = ?) ORDER BY position DESC LIMIT 1`,
		taskID,
	).Scan(&lastColumnID)
	if err != nil {
		return nil, fmt.Errorf("find last column: %w", err)
	}

	var completedAt any
	if columnID == lastColumnID {
		completedAt = time.Now().UTC()
	}

	_, err = s.db.ExecContext(ctx,
		`UPDATE tasks SET column_id = ?, position = ?, completed_at = ?, updated_at = ? WHERE id = ?`,
		columnID, position, completedAt, time.Now().UTC(), taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("move task %s: %w", taskID, err)
	}
	return s.Get(ctx, taskID)
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM tasks WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete task %s: %w", id, err)
	}
	return nil
}

func (s *TaskService) AddComment(ctx context.Context, taskID, author, body string) (*model.Comment, error) {
	if author == "" {
		author = "claude"
	}
	c := model.Comment{
		ID:        uuid.NewString(),
		TaskID:    taskID,
		Author:    author,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO comments (id, task_id, author, body, created_at) VALUES (?, ?, ?, ?, ?)`,
		c.ID, c.TaskID, c.Author, c.Body, c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert comment: %w", err)
	}
	return &c, nil
}

func (s *TaskService) ListComments(ctx context.Context, taskID string) ([]model.Comment, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, task_id, author, body, created_at FROM comments WHERE task_id = ? ORDER BY created_at ASC`,
		taskID,
	)
	if err != nil {
		return nil, fmt.Errorf("list comments: %w", err)
	}
	defer rows.Close()

	var comments []model.Comment
	for rows.Next() {
		var c model.Comment
		if err := rows.Scan(&c.ID, &c.TaskID, &c.Author, &c.Body, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, rows.Err()
}
