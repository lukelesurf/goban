package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/lukelesurf/goban/internal/model"
	"github.com/lukelesurf/goban/internal/service"
)

func TestTaskCreate_DefaultPriority(t *testing.T) {
	ctx := context.Background()
	_, _, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	task, err := tasks.Create(ctx, boardID, service.TaskCreate{Title: "no priority set"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.Priority != model.PriorityMedium {
		t.Errorf("want default priority %q, got %q", model.PriorityMedium, task.Priority)
	}
}

func TestTaskCreate_PlacedInFirstColumn(t *testing.T) {
	ctx := context.Background()
	_, columns, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	cols, _ := columns.List(ctx, boardID)
	firstCol := cols[0]

	task, err := tasks.Create(ctx, boardID, service.TaskCreate{Title: "new task"})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	if task.ColumnID != firstCol.ID {
		t.Errorf("want task in first column %q, got %q", firstCol.ID, task.ColumnID)
	}
}

func TestTaskList_Filters(t *testing.T) {
	ctx := context.Background()
	_, columns, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	cols, _ := columns.List(ctx, boardID)
	colA, colB := cols[0].ID, cols[1].ID

	// Seed some tasks across two columns.
	tasks.Create(ctx, boardID, service.TaskCreate{Title: "alpha", Priority: model.PriorityHigh, Labels: "frontend"})
	tasks.Create(ctx, boardID, service.TaskCreate{Title: "beta", Priority: model.PriorityLow, Labels: "backend"})
	t3, _ := tasks.Create(ctx, boardID, service.TaskCreate{Title: "gamma", Priority: model.PriorityHigh})
	tasks.Move(ctx, t3.ID, colB, 0)

	tests := []struct {
		name   string
		filter service.TaskFilter
		want   int
	}{
		{"no filter returns all", service.TaskFilter{}, 3},
		{"filter by column", service.TaskFilter{ColumnID: colA}, 2},
		{"filter by column B", service.TaskFilter{ColumnID: colB}, 1},
		{"filter by priority", service.TaskFilter{Priority: "high"}, 2},
		{"filter by label", service.TaskFilter{Label: "frontend"}, 1},
		{"search title", service.TaskFilter{Search: "alph"}, 1},
		{"search no match", service.TaskFilter{Search: "zzz"}, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			results, err := tasks.List(ctx, boardID, tc.filter)
			if err != nil {
				t.Fatalf("list tasks: %v", err)
			}
			if len(results) != tc.want {
				t.Errorf("want %d tasks, got %d", tc.want, len(results))
			}
		})
	}
}

func TestTaskMove_SetsCompletedAt(t *testing.T) {
	ctx := context.Background()
	_, columns, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	cols, _ := columns.List(ctx, boardID)
	lastCol := cols[len(cols)-1]

	task, _ := tasks.Create(ctx, boardID, service.TaskCreate{Title: "finish me"})
	if task.CompletedAt != nil {
		t.Fatal("new task should not have completed_at")
	}

	moved, err := tasks.Move(ctx, task.ID, lastCol.ID, 0)
	if err != nil {
		t.Fatalf("move task: %v", err)
	}
	if moved.CompletedAt == nil {
		t.Error("task moved to last column should have completed_at set")
	}
}

func TestTaskMove_DoesNotSetCompletedAtForMiddleColumn(t *testing.T) {
	ctx := context.Background()
	_, columns, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	cols, _ := columns.List(ctx, boardID)

	task, _ := tasks.Create(ctx, boardID, service.TaskCreate{Title: "in progress"})
	moved, err := tasks.Move(ctx, task.ID, cols[1].ID, 0)
	if err != nil {
		t.Fatalf("move task: %v", err)
	}
	if moved.CompletedAt != nil {
		t.Error("task moved to a middle column should not have completed_at")
	}
}

func TestTaskUpdate_Partial(t *testing.T) {
	ctx := context.Background()
	_, _, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	task, _ := tasks.Create(ctx, boardID, service.TaskCreate{
		Title:    "original",
		Priority: model.PriorityLow,
	})

	newTitle := "updated"
	updated, err := tasks.Update(ctx, task.ID, service.TaskUpdate{Title: &newTitle})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}

	if updated.Title != "updated" {
		t.Errorf("want title %q, got %q", "updated", updated.Title)
	}
	// Priority was not included in the update — should be unchanged.
	if updated.Priority != model.PriorityLow {
		t.Errorf("priority should be unchanged, got %q", updated.Priority)
	}
}

func TestTaskUpdate_ClearDueDate(t *testing.T) {
	ctx := context.Background()
	_, _, tasks, boardID := seedBoard(t, ctx, newTestDB(t), "Board")

	due := time.Now().Add(24 * time.Hour)
	task, _ := tasks.Create(ctx, boardID, service.TaskCreate{Title: "has due date", DueDate: &due})

	cleared, err := tasks.Update(ctx, task.ID, service.TaskUpdate{ClearDueDate: true})
	if err != nil {
		t.Fatalf("update task: %v", err)
	}
	if cleared.DueDate != nil {
		t.Error("due_date should be nil after ClearDueDate")
	}
}
