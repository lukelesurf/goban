package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/lukelesurf/goban/internal/db"
	"github.com/lukelesurf/goban/internal/service"
	"github.com/lukelesurf/goban/internal/ui"
)

var dbPath string

var rootCmd = &cobra.Command{
	Use:   "goban",
	Short: "A self-hosted kanban board",
}

func main() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "goban.db", "path to SQLite database file")

	database, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	boards := service.NewBoardService(database)
	columns := service.NewColumnService(database)
	tasks := service.NewTaskService(database)

	rootCmd.AddCommand(
		newServeCmd(database, ui.Assets),
		newBoardCmd(boards, columns),
		newTaskCmd(boards, tasks),
	)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
