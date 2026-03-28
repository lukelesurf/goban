package service_test

import (
	"context"
	"testing"
)

func TestBoardCreate_DefaultColumns(t *testing.T) {
	ctx := context.Background()
	d := newTestDB(t)
	_, columns, _, boardID := seedBoard(t, ctx, d, "My Board")

	cols, err := columns.List(ctx, boardID)
	if err != nil {
		t.Fatalf("list columns: %v", err)
	}

	want := []string{"To Do", "In Progress", "Review", "Done"}
	if len(cols) != len(want) {
		t.Fatalf("want %d columns, got %d", len(want), len(cols))
	}
	for i, col := range cols {
		if col.Name != want[i] {
			t.Errorf("column[%d]: want %q, got %q", i, want[i], col.Name)
		}
	}
}

func TestBoardSetDefault_Exclusivity(t *testing.T) {
	ctx := context.Background()
	d := newTestDB(t)
	boards, _, _, idA := seedBoard(t, ctx, d, "A")
	_, _, _, idB := seedBoard(t, ctx, d, "B")

	// Set A as default, then switch to B — A should no longer be default.
	if _, err := boards.SetDefault(ctx, idA); err != nil {
		t.Fatalf("set default A: %v", err)
	}
	if _, err := boards.SetDefault(ctx, idB); err != nil {
		t.Fatalf("set default B: %v", err)
	}

	all, err := boards.List(ctx)
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}

	for _, b := range all {
		switch b.ID {
		case idB:
			if !b.IsDefault {
				t.Errorf("board B should be default")
			}
		case idA:
			if b.IsDefault {
				t.Errorf("board A should no longer be default after switching to B")
			}
		}
	}
}

func TestBoardCreate_KeyGeneratedFromName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"Goban", "GOBAN"},
		{"My Board", "MYBOA"},
		{"AB", "AB"},
		{"123 Numbers!", "NUMBE"},
		{"", "BOARD"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			boards, _, _, _ := seedBoard(t, ctx, newTestDB(t), tc.name)
			all, err := boards.List(ctx)
			if err != nil {
				t.Fatalf("list boards: %v", err)
			}
			if all[0].Key != tc.want {
				t.Errorf("want key %q, got %q", tc.want, all[0].Key)
			}
		})
	}
}

func TestBoardCreate_KeyCollisionHandled(t *testing.T) {
	ctx := context.Background()
	d := newTestDB(t)
	boards, _, _, _ := seedBoard(t, ctx, d, "Goban")
	_, _, _, _ = seedBoard(t, ctx, d, "Goban") // same name → collision

	all, err := boards.List(ctx)
	if err != nil {
		t.Fatalf("list boards: %v", err)
	}
	keys := map[string]bool{}
	for _, b := range all {
		if keys[b.Key] {
			t.Errorf("duplicate key %q", b.Key)
		}
		keys[b.Key] = true
	}
}

func TestBoardClearDefault(t *testing.T) {
	ctx := context.Background()
	d := newTestDB(t)
	boards, _, _, boardID := seedBoard(t, ctx, d, "Solo")

	if _, err := boards.SetDefault(ctx, boardID); err != nil {
		t.Fatalf("set default: %v", err)
	}
	if err := boards.ClearDefault(ctx, boardID); err != nil {
		t.Fatalf("clear default: %v", err)
	}

	b, err := boards.Get(ctx, boardID)
	if err != nil {
		t.Fatalf("get board: %v", err)
	}
	if b.IsDefault {
		t.Error("board should not be default after ClearDefault")
	}
}
