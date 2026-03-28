package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/model"
)

var defaultColumns = []string{"To Do", "In Progress", "Review", "Done"}

var nonLetters = regexp.MustCompile(`[^a-zA-Z]`)

// boardKeyBase derives a short uppercase identifier from a board name (e.g. "My Board" → "MYBOA").
func boardKeyBase(name string) string {
	letters := nonLetters.ReplaceAllString(name, "")
	base := strings.ToUpper(letters)
	if len(base) > 5 {
		base = base[:5]
	}
	if base == "" {
		base = "BOARD"
	}
	return base
}

type BoardService struct {
	db *db.DB
}

func NewBoardService(db *db.DB) *BoardService {
	return &BoardService{db: db}
}

const boardColumns = `id, name, description, key, is_default, created_at, updated_at`

func scanBoard(row interface{ Scan(...any) error }, b *model.Board) error {
	return row.Scan(&b.ID, &b.Name, &b.Description, &b.Key, &b.IsDefault, &b.CreatedAt, &b.UpdatedAt)
}

func (s *BoardService) List(ctx context.Context) ([]model.Board, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT `+boardColumns+` FROM boards ORDER BY created_at ASC`)
	if err != nil {
		return nil, fmt.Errorf("list boards: %w", err)
	}
	defer rows.Close()

	var boards []model.Board
	for rows.Next() {
		var b model.Board
		if err := scanBoard(rows, &b); err != nil {
			return nil, fmt.Errorf("scan board: %w", err)
		}
		boards = append(boards, b)
	}
	return boards, rows.Err()
}

func (s *BoardService) Get(ctx context.Context, id string) (*model.Board, error) {
	var b model.Board
	err := scanBoard(s.db.QueryRowContext(ctx, `SELECT `+boardColumns+` FROM boards WHERE id = ?`, id), &b)
	if err != nil {
		return nil, fmt.Errorf("get board %s: %w", id, err)
	}
	return &b, nil
}

func (s *BoardService) SetDefault(ctx context.Context, id string) (*model.Board, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `UPDATE boards SET is_default = 0`); err != nil {
		return nil, fmt.Errorf("clear defaults: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE boards SET is_default = 1 WHERE id = ?`, id); err != nil {
		return nil, fmt.Errorf("set default: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit set default: %w", err)
	}
	return s.Get(ctx, id)
}

func (s *BoardService) ClearDefault(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE boards SET is_default = 0 WHERE id = ?`, id)
	return err
}

func (s *BoardService) Create(ctx context.Context, name, description string) (*model.Board, error) {
	now := time.Now().UTC()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Generate a unique board key within the transaction.
	base := boardKeyBase(name)
	key := base
	for i := 2; ; i++ {
		var count int
		tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM boards WHERE key = ?`, key).Scan(&count)
		if count == 0 {
			break
		}
		key = fmt.Sprintf("%s%d", base, i)
	}

	b := model.Board{
		ID:          uuid.NewString(),
		Name:        name,
		Description: description,
		Key:         key,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO boards (id, name, description, key, is_default, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		b.ID, b.Name, b.Description, b.Key, b.IsDefault, b.CreatedAt, b.UpdatedAt,
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
