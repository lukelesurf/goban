package service_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/service"
)

// newTestDB opens a fresh in-process SQLite database for each test.
// t.TempDir() gives us an isolated directory that is removed after the test.
func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	d, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { d.Close() })
	return d
}

// seedBoard creates a board and returns the three services wired to the same DB.
func seedBoard(t *testing.T, ctx context.Context, d *db.DB, name string) (
	*service.BoardService, *service.ColumnService, *service.TaskService, string,
) {
	t.Helper()
	boards := service.NewBoardService(d)
	columns := service.NewColumnService(d)
	tasks := service.NewTaskService(d)

	board, err := boards.Create(ctx, name, "")
	if err != nil {
		t.Fatalf("create board %q: %v", name, err)
	}
	return boards, columns, tasks, board.ID
}
