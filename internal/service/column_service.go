package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/model"
)

type ColumnService struct {
	db *db.DB
}

func NewColumnService(db *db.DB) *ColumnService {
	return &ColumnService{db: db}
}

func (s *ColumnService) List(ctx context.Context, boardID string) ([]model.Column, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, board_id, name, position, wip_limit, created_at FROM columns WHERE board_id = ? ORDER BY position ASC`,
		boardID,
	)
	if err != nil {
		return nil, fmt.Errorf("list columns: %w", err)
	}
	defer rows.Close()

	var cols []model.Column
	for rows.Next() {
		var c model.Column
		if err := rows.Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.WIPLimit, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}
		cols = append(cols, c)
	}
	return cols, rows.Err()
}

func (s *ColumnService) Create(ctx context.Context, boardID, name string, position, wipLimit int) (*model.Column, error) {
	c := model.Column{
		ID:        uuid.NewString(),
		BoardID:   boardID,
		Name:      name,
		Position:  position,
		WIPLimit:  wipLimit,
		CreatedAt: time.Now().UTC(),
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO columns (id, board_id, name, position, wip_limit, created_at) VALUES (?, ?, ?, ?, ?, ?)`,
		c.ID, c.BoardID, c.Name, c.Position, c.WIPLimit, c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert column: %w", err)
	}
	return &c, nil
}

func (s *ColumnService) Update(ctx context.Context, id, name string, position, wipLimit int) (*model.Column, error) {
	_, err := s.db.ExecContext(ctx,
		`UPDATE columns SET name = ?, position = ?, wip_limit = ? WHERE id = ?`,
		name, position, wipLimit, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update column %s: %w", id, err)
	}

	var c model.Column
	err = s.db.QueryRowContext(ctx,
		`SELECT id, board_id, name, position, wip_limit, created_at FROM columns WHERE id = ?`, id,
	).Scan(&c.ID, &c.BoardID, &c.Name, &c.Position, &c.WIPLimit, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get column %s: %w", id, err)
	}
	return &c, nil
}

func (s *ColumnService) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM columns WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete column %s: %w", id, err)
	}
	return nil
}
