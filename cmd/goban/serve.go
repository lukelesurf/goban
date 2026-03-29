package main

import (
	"embed"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/lukelesurf/goban/internal/api"
	"github.com/lukelesurf/goban/internal/db"
	gobanmcp "github.com/lukelesurf/goban/internal/mcp"
	"github.com/lukelesurf/goban/internal/service"
)

func newServeCmd(database *db.DB, uiAssets embed.FS) *cobra.Command {
	var addr string
	var mcpAddr string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the HTTP server (REST API + MCP)",
		RunE: func(cmd *cobra.Command, args []string) error {
			boards := service.NewBoardService(database)
			columns := service.NewColumnService(database)
			tasks := service.NewTaskService(database)
			storageDir := filepath.Join(filepath.Dir(dbPath), "attachments")
			attachments := service.NewAttachmentService(database, storageDir)

			mcpServer := gobanmcp.NewServer(boards, columns, tasks, attachments)

			router := api.NewRouter(
				api.NewBoardHandler(boards, columns, tasks),
				api.NewColumnHandler(columns),
				api.NewTaskHandler(tasks, attachments),
				api.NewAttachmentHandler(attachments),
				uiAssets,
			)

			router.Mount("/mcp", mcpServer)

			fmt.Printf("goban API  listening on %s\n", addr)
			fmt.Printf("goban MCP  listening on %s/mcp\n", mcpAddr)

			go func() {
				if err := mcpServer.Start(mcpAddr); err != nil {
					fmt.Printf("mcp server error: %v\n", err)
				}
			}()

			return http.ListenAndServe(addr, router)
		},
	}

	cmd.Flags().StringVar(&addr, "addr", ":8081", "address for REST API")
	cmd.Flags().StringVar(&mcpAddr, "mcp-addr", ":9091", "address for MCP server")
	return cmd
}
