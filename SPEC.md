# Goban: a self-hosted kanban board with an MCP server

> This is the spec goban was built from, kept as it was written. The code has
> moved on in places: configuration is command-line flags, not the `GOBAN_*`
> environment variables below, and due dates, attachments and task keys were
> added once the board was in daily use. Where the two disagree, the code and
> the [README](README.md) are right.

> A lightweight, self-hosted kanban board written in Go. Named for the board used to play the game of Go: a board for managing boards, written in Go. Designed for both human use (via web UI) and AI agent use (via MCP server). Runs as a single binary with embedded assets and SQLite storage.

## Overview

Goban is a personal/small-team kanban board that prioritises simplicity and API-first design. The primary consumers are:

1. **Luke**: via a web UI for visual board management
2. **Claude Code**: via an MCP server for programmatic task creation, movement, and querying

The system runs on a home server and is accessed remotely from a workstation where Claude Code operates.

## Architecture

```
┌─────────────────────────────────────────────┐
│                 Single Go Binary            │
│                                             │
│  ┌──────────┐  ┌──────────┐  ┌───────────┐ │
│  │  Web UI  │  │ REST API │  │MCP Server │ │
│  │ (embed)  │  │ (Chi)    │  │(Streamable│ │
│  │          │  │          │  │   HTTP)   │ │
│  └────┬─────┘  └────┬─────┘  └─────┬─────┘ │
│       │              │              │       │
│       └──────────────┼──────────────┘       │
│                      │                      │
│              ┌───────┴────────┐             │
│              │  Service Layer │             │
│              └───────┬────────┘             │
│                      │                      │
│              ┌───────┴────────┐             │
│              │   SQLite (CGo) │             │
│              └────────────────┘             │
└─────────────────────────────────────────────┘
```

### Key Design Decisions

- **Single binary**: Go's `embed` package for static assets. No separate frontend build/deploy.
- **SQLite**: Simple, file-based, zero-config. Use `github.com/mattn/go-sqlite3` (CGo) or `modernc.org/sqlite` (pure Go, easier cross-compile).
- **Chi router**: Lightweight, idiomatic, stdlib-compatible. Not Gin, not Echo.
- **MCP via Streamable HTTP**: Must be network-accessible (not stdio) since Claude Code runs on a different machine. Use `github.com/mark3labs/mcp-go` SDK.
- **No auth initially**: Runs on a private network. Can add basic auth or API key later.

## Data Model

### Board

| Field       | Type     | Notes                          |
|-------------|----------|--------------------------------|
| id          | TEXT     | UUID, primary key              |
| name        | TEXT     | e.g. "Trading Platform"        |
| description | TEXT     | Optional                       |
| created_at  | DATETIME | UTC                            |
| updated_at  | DATETIME | UTC                            |

### Column

| Field      | Type    | Notes                              |
|------------|---------|------------------------------------|
| id         | TEXT    | UUID, primary key                  |
| board_id   | TEXT    | FK → Board                         |
| name       | TEXT    | e.g. "To Do", "In Progress"       |
| position   | INTEGER | Sort order (0-indexed)             |
| wip_limit  | INTEGER | 0 = unlimited                      |
| created_at | DATETIME| UTC                                |

Default columns on board creation: `To Do`, `In Progress`, `Review`, `Done`

### Task

| Field       | Type     | Notes                                    |
|-------------|----------|------------------------------------------|
| id          | TEXT     | UUID, primary key                        |
| board_id    | TEXT     | FK → Board                               |
| column_id   | TEXT     | FK → Column                              |
| title       | TEXT     | Short description                        |
| description | TEXT     | Markdown body, optional                  |
| priority    | TEXT     | `low`, `medium`, `high`, `critical`      |
| labels      | TEXT     | Comma-separated tags                     |
| position    | INTEGER  | Sort order within column                 |
| created_at  | DATETIME | UTC                                      |
| updated_at  | DATETIME | UTC                                      |
| completed_at| DATETIME | Set when moved to final column, nullable |

### Comment

| Field      | Type     | Notes                     |
|------------|----------|---------------------------|
| id         | TEXT     | UUID, primary key         |
| task_id    | TEXT     | FK → Task                 |
| author     | TEXT     | Free text (e.g. "claude") |
| body       | TEXT     | Markdown                  |
| created_at | DATETIME | UTC                       |

## REST API

Base path: `/api/v1`

### Boards
- `GET    /boards`: List all boards
- `POST   /boards`: Create board
- `GET    /boards/:id`: Get board with columns and task counts
- `PUT    /boards/:id`: Update board
- `DELETE /boards/:id`: Delete board and all children

### Columns
- `GET    /boards/:id/columns`: List columns for board
- `POST   /boards/:id/columns`: Add column
- `PUT    /columns/:id`: Update column (name, position, wip_limit)
- `DELETE /columns/:id`: Delete column (must be empty or reassign tasks)

### Tasks
- `GET    /boards/:id/tasks`: List tasks (filterable by column, priority, label)
- `POST   /boards/:id/tasks`: Create task (defaults to first column)
- `GET    /tasks/:id`: Get task with comments
- `PUT    /tasks/:id`: Update task fields
- `DELETE /tasks/:id`: Delete task
- `POST   /tasks/:id/move`: Move task to column (with position). Body: `{"column_id": "...", "position": 0}`
- `GET    /tasks/:id/comments`: List comments
- `POST   /tasks/:id/comments`: Add comment

### Query Parameters (for task listing)
- `column_id`: filter by column
- `priority`: filter by priority
- `label`: filter by label (contains match)
- `search`: full-text search on title and description

## MCP Server

Transport: **Streamable HTTP** on a configurable port (default 9090), path `/mcp`.

### Tools

| Tool                | Description                                        | Parameters                                                  |
|---------------------|----------------------------------------------------|-------------------------------------------------------------|
| `list_boards`       | List all kanban boards                              | _(none)_                                                    |
| `get_board`         | Get board details with columns and tasks            | `board_id` (string, required)                               |
| `create_board`      | Create a new board with default columns             | `name` (string), `description` (string, optional)           |
| `create_task`       | Add a task to a board                               | `board_id`, `title`, `description?`, `priority?`, `labels?` |
| `move_task`         | Move a task to a different column                   | `task_id`, `column_id`, `position?`                         |
| `update_task`       | Update task fields                                  | `task_id`, and any of: `title`, `description`, `priority`, `labels` |
| `get_task`          | Get full task detail including comments              | `task_id`                                                   |
| `delete_task`       | Delete a task                                       | `task_id`                                                   |
| `add_comment`       | Add a comment to a task                             | `task_id`, `body`, `author?` (defaults to "claude")         |
| `list_tasks`        | List/search tasks on a board                        | `board_id`, `column_id?`, `priority?`, `label?`, `search?`  |
| `get_board_summary` | Quick stats: task count per column, overdue, etc.   | `board_id`                                                  |

### MCP Prompts

| Prompt              | Description                                                      |
|---------------------|------------------------------------------------------------------|
| `plan_project`      | Create a kanban board for a project, break it into tasks         |
| `resume_work`       | Find the board, pick the next task, and work on it               |

### Tool Response Format

All tool responses should return well-structured JSON that Claude can reason about. Example for `get_board`:

```json
{
  "board": {
    "id": "abc-123",
    "name": "Trading Platform",
    "columns": [
      {
        "name": "To Do",
        "wip_limit": 0,
        "tasks": [
          {"id": "t-1", "title": "Add stop-loss logic", "priority": "high", "labels": "strategy,risk"}
        ]
      }
    ]
  }
}
```

## Web UI

A minimal, functional web UI embedded in the binary. Not trying to win design awards. It just needs to be usable.

### Tech Stack
- Vanilla HTML/CSS/JS or Alpine.js + Tailwind (via CDN in dev, embedded in prod)
- Drag-and-drop for task movement (HTML5 drag API or SortableJS)
- Go `html/template` for server-rendered pages, or a small SPA with fetch calls to the REST API

### Views
1. **Board List**: cards showing each board with task count summary
2. **Board View**: columnar kanban layout with tasks as cards, drag to move
3. **Task Detail**: slide-out or modal with full description, comments, edit form

### Embedded Assets
```go
//go:embed ui/dist/*
var uiAssets embed.FS
```

Serve from root `/` with the API mounted at `/api/v1` and MCP at `/mcp`.

## Configuration

Via environment variables (12-factor style):

| Variable       | Default          | Description                    |
|----------------|------------------|--------------------------------|
| `GOBAN_PORT`   | `8080`           | HTTP port for web UI + API     |
| `GOBAN_MCP_PORT` | `9090`         | MCP server port                |
| `GOBAN_DB_PATH`| `./goban.db`     | SQLite database file path      |
| `GOBAN_HOST`   | `0.0.0.0`        | Bind address                   |

## Project Structure

```
goban/
├── cmd/
│   └── goban/
│       └── main.go              # Entry point, wire everything up
├── internal/
│   ├── db/
│   │   ├── sqlite.go            # DB connection, migrations
│   │   └── migrations.go        # SQL migration statements
│   ├── model/
│   │   ├── board.go
│   │   ├── column.go
│   │   ├── task.go
│   │   └── comment.go
│   ├── service/
│   │   ├── board_service.go     # Business logic
│   │   ├── task_service.go
│   │   └── column_service.go
│   ├── api/
│   │   ├── router.go            # Chi router setup
│   │   ├── board_handler.go
│   │   ├── task_handler.go
│   │   ├── column_handler.go
│   │   └── middleware.go        # Logging, CORS, etc.
│   └── mcp/
│       ├── server.go            # MCP server setup
│       ├── tools.go             # Tool definitions
│       └── prompts.go           # MCP prompts
├── ui/
│   ├── dist/                    # Built/static frontend assets
│   ├── index.html
│   ├── app.js
│   └── style.css
├── go.mod
├── go.sum
├── Dockerfile
├── docker-compose.yml
├── SPEC.md
└── README.md
```

## Deployment

### Docker Compose

```yaml
services:
  goban:
    build: .
    ports:
      - "8080:8080"   # Web UI + REST API
      - "9090:9090"   # MCP Server
    volumes:
      - goban-data:/data
    environment:
      - GOBAN_DB_PATH=/data/goban.db
      - GOBAN_HOST=0.0.0.0
    restart: unless-stopped

volumes:
  goban-data:
```

### Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache gcc musl-dev   # needed for CGo SQLite
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -o goban ./cmd/goban

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/goban /usr/local/bin/goban
EXPOSE 8080 9090
ENTRYPOINT ["goban"]
```

### Claude Code MCP Configuration

In the project's `.claude/mcp.json` or global Claude Code config:

```json
{
  "mcpServers": {
    "goban": {
      "type": "streamable-http",
      "url": "http://localhost:9090/mcp"
    }
  }
}
```

> **Note**: Verify `mcp-go`'s support for streamable HTTP transport. If only stdio is available in the SDK, use `mcp-proxy` (https://github.com/punkpeye/mcp-proxy) to bridge stdio→HTTP, or implement SSE transport directly.

## Non-Goals (for v1)

- User authentication / multi-tenancy
- Due dates and calendar integration
- File attachments
- Email notifications
- Sub-tasks (keep it flat for now)
- Real-time websocket updates (polling or manual refresh is fine)

## Future Ideas (v2+)

- API key auth header
- Webhooks on task transitions
- GitHub integration (auto-create tasks from issues)
- Time tracking
- Archive completed tasks
- Dark mode toggle in UI