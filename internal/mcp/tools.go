package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/lukelesurf/goban/internal/model"
	"github.com/lukelesurf/goban/internal/service"
)

func registerTools(s *server.MCPServer, boards *service.BoardService, columns *service.ColumnService, tasks *service.TaskService) {
	s.AddTool(mcp.NewTool("list_boards",
		mcp.WithDescription("List all kanban boards"),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		bs, err := boards.List(ctx)
		if err != nil {
			return nil, err
		}
		return jsonResult(bs)
	})

	s.AddTool(mcp.NewTool("get_board",
		mcp.WithDescription("Get board details with columns and tasks"),
		mcp.WithString("board_id", mcp.Required(), mcp.Description("Board ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		boardID, err := req.RequireString("board_id")
		if err != nil {
			return nil, err
		}

		board, err := boards.Get(ctx, boardID)
		if err != nil {
			return nil, fmt.Errorf("get board: %w", err)
		}

		cols, err := columns.List(ctx, boardID)
		if err != nil {
			return nil, fmt.Errorf("list columns: %w", err)
		}

		type colWithTasks struct {
			model.Column
			Tasks []model.Task `json:"tasks"`
		}

		colsWithTasks := make([]colWithTasks, 0, len(cols))
		for _, c := range cols {
			ts, err := tasks.List(ctx, boardID, service.TaskFilter{ColumnID: c.ID})
			if err != nil {
				return nil, fmt.Errorf("list tasks for column %s: %w", c.ID, err)
			}
			colsWithTasks = append(colsWithTasks, colWithTasks{Column: c, Tasks: ts})
		}

		return jsonResult(map[string]any{
			"board":   board,
			"columns": colsWithTasks,
		})
	})

	s.AddTool(mcp.NewTool("create_board",
		mcp.WithDescription("Create a new board with default columns"),
		mcp.WithString("name", mcp.Required(), mcp.Description("Board name")),
		mcp.WithString("description", mcp.Description("Optional description")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return nil, err
		}
		description := req.GetString("description", "")
		b, err := boards.Create(ctx, name, description)
		if err != nil {
			return nil, err
		}
		return jsonResult(b)
	})

	s.AddTool(mcp.NewTool("create_task",
		mcp.WithDescription("Add a task to a board (placed in the first column)"),
		mcp.WithString("board_id", mcp.Required(), mcp.Description("Board ID")),
		mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
		mcp.WithString("description", mcp.Description("Task description (Markdown)")),
		mcp.WithString("priority", mcp.Description("low, medium, high, or critical (default: medium)")),
		mcp.WithString("labels", mcp.Description("Comma-separated labels")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		boardID, err := req.RequireString("board_id")
		if err != nil {
			return nil, err
		}
		title, err := req.RequireString("title")
		if err != nil {
			return nil, err
		}
		description := req.GetString("description", "")
		priority := req.GetString("priority", "")
		labels := req.GetString("labels", "")

		t, err := tasks.Create(ctx, boardID, service.TaskCreate{
				Title:       title,
				Description: description,
				Priority:    model.Priority(priority),
				Labels:      labels,
			})
		if err != nil {
			return nil, err
		}
		return jsonResult(t)
	})

	s.AddTool(mcp.NewTool("get_task",
		mcp.WithDescription("Get full task detail including comments"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return nil, err
		}
		t, err := tasks.Get(ctx, taskID)
		if err != nil {
			return nil, err
		}
		comments, err := tasks.ListComments(ctx, taskID)
		if err != nil {
			return nil, err
		}
		return jsonResult(map[string]any{"task": t, "comments": comments})
	})

	s.AddTool(mcp.NewTool("update_task",
		mcp.WithDescription("Update task fields"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("title", mcp.Description("New title")),
		mcp.WithString("description", mcp.Description("New description")),
		mcp.WithString("priority", mcp.Description("New priority")),
		mcp.WithString("labels", mcp.Description("New labels")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return nil, err
		}

		u := service.TaskUpdate{}
		if v := req.GetString("title", ""); v != "" {
			u.Title = &v
		}
		if v := req.GetString("description", ""); v != "" {
			u.Description = &v
		}
		if v := req.GetString("priority", ""); v != "" {
			p := model.Priority(v)
			u.Priority = &p
		}
		if v := req.GetString("labels", ""); v != "" {
			u.Labels = &v
		}

		t, err := tasks.Update(ctx, taskID, u)
		if err != nil {
			return nil, err
		}
		return jsonResult(t)
	})

	s.AddTool(mcp.NewTool("move_task",
		mcp.WithDescription("Move a task to a different column"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("column_id", mcp.Required(), mcp.Description("Target column ID")),
		mcp.WithNumber("position", mcp.Description("Position within the column (default: 0)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return nil, err
		}
		columnID, err := req.RequireString("column_id")
		if err != nil {
			return nil, err
		}
		position := req.GetInt("position", 0)

		t, err := tasks.Move(ctx, taskID, columnID, position)
		if err != nil {
			return nil, err
		}
		return jsonResult(t)
	})

	s.AddTool(mcp.NewTool("delete_task",
		mcp.WithDescription("Delete a task"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return nil, err
		}
		if err := tasks.Delete(ctx, taskID); err != nil {
			return nil, err
		}
		return mcp.NewToolResultText("deleted"), nil
	})

	s.AddTool(mcp.NewTool("add_comment",
		mcp.WithDescription("Add a comment to a task"),
		mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID")),
		mcp.WithString("body", mcp.Required(), mcp.Description("Comment body (Markdown)")),
		mcp.WithString("author", mcp.Description("Author name (default: claude)")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		taskID, err := req.RequireString("task_id")
		if err != nil {
			return nil, err
		}
		body, err := req.RequireString("body")
		if err != nil {
			return nil, err
		}
		author := req.GetString("author", "")
		c, err := tasks.AddComment(ctx, taskID, author, body)
		if err != nil {
			return nil, err
		}
		return jsonResult(c)
	})

	s.AddTool(mcp.NewTool("list_tasks",
		mcp.WithDescription("List and search tasks on a board"),
		mcp.WithString("board_id", mcp.Required(), mcp.Description("Board ID")),
		mcp.WithString("column_id", mcp.Description("Filter by column ID")),
		mcp.WithString("priority", mcp.Description("Filter by priority")),
		mcp.WithString("label", mcp.Description("Filter by label (contains match)")),
		mcp.WithString("search", mcp.Description("Search title and description")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		boardID, err := req.RequireString("board_id")
		if err != nil {
			return nil, err
		}
		ts, err := tasks.List(ctx, boardID, service.TaskFilter{
			ColumnID: req.GetString("column_id", ""),
			Priority: req.GetString("priority", ""),
			Label:    req.GetString("label", ""),
			Search:   req.GetString("search", ""),
		})
		if err != nil {
			return nil, err
		}
		return jsonResult(ts)
	})

	s.AddTool(mcp.NewTool("get_board_summary",
		mcp.WithDescription("Get quick stats: task count per column"),
		mcp.WithString("board_id", mcp.Required(), mcp.Description("Board ID")),
	), func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		boardID, err := req.RequireString("board_id")
		if err != nil {
			return nil, err
		}

		board, err := boards.Get(ctx, boardID)
		if err != nil {
			return nil, err
		}

		cols, err := columns.List(ctx, boardID)
		if err != nil {
			return nil, err
		}

		type colSummary struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Count    int    `json:"task_count"`
			WIPLimit int    `json:"wip_limit"`
		}

		summaries := make([]colSummary, 0, len(cols))
		for _, c := range cols {
			ts, err := tasks.List(ctx, boardID, service.TaskFilter{ColumnID: c.ID})
			if err != nil {
				return nil, err
			}
			summaries = append(summaries, colSummary{
				ID:       c.ID,
				Name:     c.Name,
				Count:    len(ts),
				WIPLimit: c.WIPLimit,
			})
		}

		return jsonResult(map[string]any{
			"board":   board,
			"columns": summaries,
		})
	})
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return mcp.NewToolResultText(string(data)), nil
}
