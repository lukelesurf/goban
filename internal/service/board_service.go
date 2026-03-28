package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/model"
)

var defaultColumns = []string{"To Do", "In Progress", "Review", "Done"}

type BoardService struct {
	db *db.DB
}

func NewBoardService(db *db.DB) *BoardService {
	return &BoardService{db: db}
}

func (s *BoardService) List(ctx context.Context) ([]model.Board, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, name, description, created_at, updated_at FROM boards ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}
	defer rows.Close()

	var boards []model.Board
	for rows.Next() {
		var b model.Board
		if err := rows.Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan board: %w", err)
		}
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

func (s *BoardService) Get(ctx context.Context, id string) (*model.Board, error) {
	var b model.Board
	err := s.db.QueryRowContext(ctx, `SELECT id, name, description, created_at, updated_at FROM boards WHERE id = ?`, id).
		Scan(&b.ID, &b.Name, &b.Description, &b.CreatedAt, &b.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get board %s: %w", id, err)
	}
	return &b, nil
}

func (s *BoardService) Create(ctx context.Context, name, description string) (*model.Board, error) {
	now := time.Now().UTC()
	b := model.Board{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO boards (id, name, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`,
		b.ID, b.Name, b.Description, b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert board: %w", err)
	}

	for i, colName := range defaultColumns {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO columns (id, board_id, name, position, created_at) VALUES (?, ?, ?, ?, ?)`,
			uuid.NewString(), b.ID, colName, i, now,
		)
		if err != nil {
			return nil, fmt.Errorf("insert default column %q: %w", colName, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit board create: %w", err)
	}

	return &b, nil
}

func (s *BoardService) Update(ctx context.Context, id, name, description string) (*model.Board, error) {
	now := time.Now().UTC()
	_, err := s.db.ExecContext(ctx,
		`UPDATE boards SET name = ?, description = ?, updated_at = ? WHERE id = ?`,
		name, description, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("update board %s: %w", id, err)
	}
	return s.Get(ctx, id)
}

func (s *BoardService) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM boards WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete board %s: %w", id, err)
	}
	return nil
}
