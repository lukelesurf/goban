package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/lukelesurf/goban/internal/service"
)

func newBoardCmd(boards *service.BoardService, columns *service.ColumnService) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "board",
		Short: "Manage boards",
	}

	cmd.AddCommand(
		boardListCmd(boards),
		boardCreateCmd(boards),
		boardShowCmd(boards, columns),
		boardDeleteCmd(boards),
	)

	return cmd
}

func boardListCmd(boards *service.BoardService) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all boards",
		RunE: func(cmd *cobra.Command, args []string) error {
			bs, err := boards.List(context.Background())
			if err != nil {
				return err
			}
			if len(bs) == 0 {
				fmt.Println("no boards found")
				return nil
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tDESCRIPTION")
			for _, b := range bs {
				fmt.Fprintf(w, "%s\t%s\t%s\n", b.ID, b.Name, b.Description)
			}
			return w.Flush()
		},
	}
}

func boardCreateCmd(boards *service.BoardService) *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := boards.Create(context.Background(), args[0], description)
			if err != nil {
				return err
			}
			fmt.Printf("created board %s (%s)\n", b.Name, b.ID)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "board description")
	return cmd
}

func boardShowCmd(boards *service.BoardService, columns *service.ColumnService) *cobra.Command {
	return &cobra.Command{
		Use:   "show <id>",
		Short: "Show a board and its columns",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := boards.Get(context.Background(), args[0])
			if err != nil {
				return err
			}
			fmt.Printf("Board: %s\nDescription: %s\n\n", b.Name, b.Description)

			cols, err := columns.List(context.Background(), b.ID)
			if err != nil {
				return err
			}
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "POSITION\tNAME\tWIP LIMIT\tID")
			for _, c := range cols {
				wip := "-"
				if c.WIPLimit > 0 {
					wip = fmt.Sprintf("%d", c.WIPLimit)
				}
				fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", c.Position, c.Name, wip, c.ID)
			}
			return w.Flush()
		},
	}
}

func boardDeleteCmd(boards *service.BoardService) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a board",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := boards.Delete(context.Background(), args[0]); err != nil {
				return err
			}
			fmt.Printf("deleted board %s\n", args[0])
			return nil
		},
	}
}
