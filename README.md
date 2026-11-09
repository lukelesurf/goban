# goban

A self-hosted kanban board built for working alongside Claude Code. It's one
Go binary on SQLite, and it serves three things from the same service layer:

- a web UI, for me
- a REST API, for scripts
- an MCP server over streamable HTTP, so Claude Code can read and move the same
  tasks I'm looking at

The name is Go plus kanban. It also turns out to be the word for the board the
game of Go is played on, which suits a board for boards.

Every other project on my home server keeps its tickets here. When Claude Code
picks up a ticket it moves the card to In Progress, comments on what it found,
and moves it to Done once the change is committed and verified. The board ends
up as the record of what each session did.

## What it does

- Boards with columns, created with a default set (To Do, In Progress, Review,
  Done). Cards drag between columns, and every column stretches to the full
  board height so there's always somewhere to drop.
- Tasks carry a Markdown description, priority, labels, a due date and comments.
- Each task gets a short key from its board's key and a counter, such as
  `GOBAN-9`. Anywhere that takes a task ID also takes its key, and boards
  resolve by key too, so an agent can say `move GOBAN-9` and not need a UUID.
- File attachments on tasks, up to 25 MB each, stored next to the database.
- A default board that opens on start-up. The board you're on survives a refresh.

## MCP tools

`list_boards`, `get_board`, `get_board_summary`, `create_board`, `create_task`,
`get_task`, `update_task`, `move_task`, `delete_task`, `add_comment`,
`list_tasks`, `attach_file` and `list_attachments`.

Two prompts come with them: `plan_project` breaks a project into a new board
of tasks, and `resume_work` finds the next task on a board and picks it up.

## Running it

With Docker:

```bash
docker build -t goban .
docker run -d -p 8080:8080 -p 9090:9090 -v goban-data:/data goban
```

The web UI is on port 8080 with the REST API under `/api/v1`, and MCP is
at `/mcp` on port 9090. Without Docker
you need Go and a C compiler, because the SQLite driver uses CGo:

```bash
go run ./cmd/goban --db goban.db serve      # dev ports: :8081 and :9091
```

The repo has a Nix flake with the toolchain, so `nix develop` (or direnv) gets
you a working shell.

Point Claude Code at the MCP server in `.mcp.json`:

```json
{
  "mcpServers": {
    "goban": { "type": "streamable-http", "url": "http://localhost:9090/mcp" }
  }
}
```

There is a CLI over the same service layer for quick jobs and scripts:
`goban board list`, `goban task create "title" --board GOBAN`,
`goban task move GOBAN-9 --column <id>`, and so on.

There's no authentication. It's meant for a private network. Don't put it on
the internet as it is.

## Layout

```
cmd/goban/          CLI entry point: serve, board and task commands (cobra)
internal/model/     boards, columns, tasks, comments, attachments
internal/db/        SQLite schema and migrations
internal/service/   all business rules; the UI, API, MCP and CLI only call this
internal/api/       REST handlers (chi)
internal/mcp/       MCP tools and prompts (mcp-go)
internal/ui/        the web UI, one embedded HTML file (Alpine.js, Tailwind, SortableJS)
```

Keeping every rule in `internal/service` is the main design decision. A task
moved by Claude over MCP, by me in the browser or by a script through the CLI
goes through the same code, so the three can't drift apart.

## How this was built

I wrote [SPEC.md](SPEC.md) first, then built goban with Claude Code over about
two weeks, and it has been running my boards ever since. The spec is kept as it
was written. A few things it lists as non-goals, such as due dates and
attachments, turned up as soon as the board was in daily use.

Tests cover the service layer, which is where the rules live: board and task
creation, key generation and collisions, filters, moves, partial updates
and the default board. The HTTP and MCP layers
are thin adapters over it and aren't tested directly.

This repo is a published copy of a private one, with its history rewritten for
publication. The module path and hostnames changed, and the CI that deployed it
to my server was replaced by a test-only workflow.

## Licence

MIT, see [LICENSE](LICENSE).
