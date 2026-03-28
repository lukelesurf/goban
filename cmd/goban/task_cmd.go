package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lukelesurf/goban/internal/model"
	"github.com/lukelesurf/goban/internal/service"
)

func newTaskCmd(boards *service.BoardService, tasks *service.TaskService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	cmd.AddCommand(
		taskListCmd(tasks),
		taskCreateCmd(boards, tasks),
		taskShowCmd(tasks),
		taskMoveCmd(tasks),
		taskDeleteCmd(tasks),
		taskCommentCmd(tasks),
	)

	return cmd
}

func taskListCmd(tasks *service.TaskService) *cobra.Command {
	var boardID, columnID, priority, label, search string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks on a board",
		RunE: func(cmd *cobra.Command, args []string) error {
			ts, err := tasks.List(context.Background(), boardID, service.TaskFilter{
				ColumnID: columnID,
				Priority: priority,
				Label:    label,
				Search:   search,
			})
			if err != nil {
				return err
			}
			if len(ts) == 0 {
				fmt.Println("no tasks found")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tTITLE\tPRIORITY\tCOLUMN")
			for _, t := range ts {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.ID, t.Title, t.Priority, t.ColumnID)
			}
			return w.Flush()
		},
	}

	cmd.Flags().StringVar(&boardID, "board", "", "board ID (required)")
	cmd.Flags().StringVar(&columnID, "column", "", "filter by column ID")
	cmd.Flags().StringVar(&priority, "priority", "", "filter by priority")
	cmd.Flags().StringVar(&label, "label", "", "filter by label")
	cmd.Flags().StringVar(&search, "search", "", "search title and description")
	cmd.MarkFlagRequired("board")

	return cmd
}

func taskCreateCmd(_ *service.BoardService, tasks *service.TaskService) *cobra.Command {
	var boardID, description, priority, labels string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := tasks.Create(context.Background(), boardID, service.TaskCreate{
				Title:       args[0],
				Description: description,
				Priority:    model.Priority(priority),
				Labels:      labels,
			})
			if err != nil {
				return err
			}
			fmt.Printf("created task %s (%s)\n", t.Title, t.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&boardID, "board", "", "board ID (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "task description")
	cmd.Flags().StringVar(&priority, "priority", "medium", "priority (low, medium, high, critical)")
	cmd.Flags().StringVar(&labels, "labels", "", "comma-separated labels")
	cmd.MarkFlagRequired("board")

	return cmd
}

func taskShowCmd(tasks *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a task and its comments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := tasks.Get(context.Background(), args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Title:       %s\n", t.Title)
			fmt.Printf("Priority:    %s\n", t.Priority)
			fmt.Printf("Labels:      %s\n", t.Labels)
			fmt.Printf("Description: %s\n\n", t.Description)

			comments, err := tasks.ListComments(context.Background(), t.ID)
			if err != nil {
				return err
			}
			if len(comments) == 0 {
				fmt.Println("no comments")
				return nil
			}
			for _, c := range comments {
				fmt.Printf("[%s] %s: %s\n", c.CreatedAt.Format("2006-01-02 15:04"), c.Author, c.Body)
			}
			return nil
		},
	}
}

func taskMoveCmd(tasks *service.TaskService) *cobra.Command {
	var columnID string
	var position int

	cmd := &cobra.Command{
		Use:   "move <id>",
		Short: "Move a task to a different column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			t, err := tasks.Move(context.Background(), args[0], columnID, position)
			if err != nil {
				return err
			}
			fmt.Printf("moved task %s to column %s\n", t.Title, t.ColumnID)
			return nil
		},
	}

	cmd.Flags().StringVar(&columnID, "column", "", "target column ID (required)")
	cmd.Flags().IntVar(&position, "position", 0, "position within the column")
	cmd.MarkFlagRequired("column")

	return cmd
}

func taskDeleteCmd(tasks *service.TaskService) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := tasks.Delete(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("deleted task %s\n", args[0])
			return nil
		},
	}
}

func taskCommentCmd(tasks *service.TaskService) *cobra.Command {
	var author string

	cmd := &cobra.Command{
		Use:   "comment <task-id> <body>",
		Short: "Add a comment to a task",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := tasks.AddComment(context.Background(), args[0], author, args[1])
			if err != nil {
				return err
			}
			fmt.Printf("added comment %s\n", c.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&author, "author", "", "comment author (defaults to 'claude')")
	return cmd
}
