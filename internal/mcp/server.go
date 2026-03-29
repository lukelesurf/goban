package mcp

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/lukelesurf/goban/internal/service"
)

func NewServer(boards *service.BoardService, columns *service.ColumnService, tasks *service.TaskService, attachments *service.AttachmentService) *server.StreamableHTTPServer {
	s := server.NewMCPServer("goban", "1.0.0")

	registerTools(s, boards, columns, tasks, attachments)
	registerPrompts(s)

	return server.NewStreamableHTTPServer(s)
}
